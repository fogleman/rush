# Running the enumeration on a GPU

An analysis of whether the C++ solver on the `claude-performance` branch could be
rewritten for a GPU, and how it would be structured. All numbers below were
measured on that branch; the methodology is at the end.

If the goal is the top N hardest 7x7 puzzles rather than a complete database, see
[Wanting only the top N](#wanting-only-the-top-n) — it changes less than one
would hope, and the reasons are worth knowing before starting.

## Short answer

Yes, and the payoff is large enough to matter: 7x7 is currently out of reach on a
CPU by roughly an order of magnitude, and a single GPU plausibly closes that gap.

But not by porting `Cluster::Explore` to CUDA. A direct port assigns one cluster
to one thread, and that mapping loses, because the work is not distributed the
way the cluster count suggests: **92% of clusters contain 13% of the work.** The
design has to parallelize *inside* the expensive clusters, not just across
clusters.

The good news is that the whole program collapses onto a single primitive —
*bounded breadth-first search over an implicitly defined graph, batched across
many independent graphs* — and that primitive is one of the best-understood
things to write for a GPU. The work already done on `claude-performance` (packed
integer state, open-addressed table, CSR adjacency, table-free move generation)
happens to be most of what a GPU port needs anyway.

## What the program does today

```
Enumerator                       every distinct board layout, in DFS order
  └─ Cluster::Explore            per layout:
       ├─ forward BFS            reach the connected component; abort if some
       │                         reachable state sorts below the input
       ├─ backward BFS           multi-source from all goal states -> distances
       ├─ Solver::Solve          shortest solution from the hardest state
       ├─ Solver::SolvableWithin per removable piece: is it still as hard?
       └─ histogram              states per distance
```

Only the last stage produces output, and it fires for about one input in five
hundred.

## Measured shape of the workload

6x6, no walls, 1-in-199 sample of the enumeration (1,223,633 clusters explored):

| | |
|---|---|
| canonical | 33.3% of inputs |
| solvable (and canonical) | 13.6% |
| minimal (reported as a puzzle) | 0.194% |
| forward expansions per input | 140.4 |
| average out-degree of a state | 10.9 |
| expansions spent in clusters that turn out non-canonical | 19.1% |
| largest cluster in the sample | 463,118 states |
| BFS eccentricity (levels in the backward pass) | ≤ 29, median 1 |
| average states per cluster | 164.7 |

Single-threaded time inside `Explore`:

| phase | share |
|---|---|
| forward BFS | 71% |
| backward BFS | 5% |
| minimality — `Solver::Solve` | 13% |
| minimality — removal tests | 11% |

Enumeration itself is negligible at 6x6: 243.5M boards in 7.9 s, versus ~4
core-hours for the exploration. That flips at 7x7, discussed below.

### Where the work actually lives

Clusters bucketed by state count, with the share of forward expansions each
bucket accounts for:

| cluster size | clusters | cumulative | expansions |
|---|---|---|---|
| ≤ 255 states | 92.19% | 92.19% | **13.4%** |
| 256 – 8,191 | 7.56% | 99.75% | **53.4%** |
| ≥ 8,192 | 0.25% | 100% | **33.3%** |

This single table drives the whole design. Restated:

- A thread-per-cluster kernel with a 256-state cap disposes of 92% of the
  clusters and 13% of the runtime. It is a nice-to-have, not the design.
- A quarter of a percent of clusters hold a third of the work. Those clusters
  have 8k–460k states spread over ≤ 29 BFS levels, so their frontiers are
  thousands of states wide — **the expensive clusters are exactly the ones with
  abundant internal parallelism.** That is an unusually friendly property.
- Skew is getting worse with board size: the top 1% of clusters hold 19.6% of
  states at 5x5 and 49.0% at 6x6. Any design that treats a cluster as an
  indivisible unit of work will degrade as the board grows.

## Architecture

Six stages. Stages 2–4 are the same kernel with different parameters.

### Stage 0 — seed generation, on device

`Enumerator` is a DFS over per-row and per-column `PositionEntry` lists with
bitmask compatibility tests. Split it at the row/column boundary:

- The host (or a first kernel) enumerates **row prefixes** — the choice of
  entries for every row. Measured: 25.8M of them at 6x6, averaging 9.4 boards
  each.
- One device thread per prefix runs the remaining column DFS on an explicit
  stack of depth ≤ `BoardSize`, emitting board seeds into a global queue with a
  warp-aggregated `atomicAdd`.

25.8M work items at the outer level and ~9 boards each is plenty of parallelism
without any load-balancing effort. Streaming prefixes rather than boards also
cuts host→device traffic by ~9x, and a seed is compact: `numPieces` plus
`(base, size, horizontal)` per piece is 2 bytes per piece, under 30 bytes total.

IDs stay reproducible: count the column combinations per prefix in a cheap first
pass, prefix-sum the counts, and `id = base[prefix] + local`. Prefixes generated
in DFS order then reproduce today's numbering exactly, which matters for
resuming a long run and for diffing GPU output against CPU output.

**This stage is not optional at 7x7.** Enumeration measures 31M boards/s, so
561.3e9 boards is ~5.0 h single-threaded — which matches the `5h42m` note next
to the 7x7 `MaxID` in `config.h`. Because every worker in `main.cpp` runs the
full enumeration and discards `id % wn != wi`, that 5 h is a floor on wall-clock
time no matter how many workers you add. At 6x6 the replication costs nothing;
at 7x7 it is the dominant term.

### Stage 1 — descriptors and bucketing

One thread per seed. Build the per-cluster piece descriptor (below), compute an a
priori bound on the cluster's state count, and append the seed to one of three
tier queues. This is where the load-balancing decision is made, from a bound
that costs a dozen multiplies.

### Stage 2 — forward BFS, batched and level-synchronous

The core kernel. Input: a batch of clusters, each with a segment of a shared hash
table. State: a flat frontier array of `(clusterSlot, nodeIndex)`.

```
for level in 0 .. maxLevels:
    expand(frontier) -> nextFrontier      # one launch
    if nextFrontier empty: break
```

Because the frontier is flat, work is measured in states rather than clusters,
and the skew problem disappears: a batch mixing 8-state clusters with
400,000-state clusters produces one big well-balanced frontier per level. A
cluster that finishes early simply stops contributing.

Level count is bounded by the graph's radius, measured at ≤ 29 for 6x6, so this
is ~30 launches per batch — launch overhead is irrelevant.

**Expansion, two-phase, to kill divergence.** A state has 10.9 neighbours on
average but they arrive from a nested loop (per piece, then per slide distance),
so one-thread-per-state diverges badly. Instead:

1. Thread *t* handles one `(state, piece)` pair and counts that piece's legal
   slides — a couple of instructions given the descriptor.
2. Warp/block prefix-sum over those counts.
3. Each lane picks up exactly one neighbour slot and generates it.

This is the standard fine-grained gather from the GPU BFS literature, and it
matters here: `numPieces` is ~12 at 6x6, so step 1 already gives ~12 lanes of
work per state, and step 3 makes the tail uniform.

**Deduplication.** Per-cluster segment of a global open-addressing table, linear
probing at ≤ 0.75 load, `atomicCAS` on the key. A lane that loses the CAS
re-reads and compares. Node indexes come from an `atomicAdd` on the cluster's
counter. Buckets aligned to 32 bytes so a warp-cooperative probe of four slots is
one transaction.

Segmented tables replace the generation-stamp trick in `Cluster::NewGeneration`.
The stamp exists to avoid clearing a reused table; on a GPU you size each
cluster's segment from Stage 1's bound and clear only what you allocated, which
is a coalesced write and nearly free.

**Canonicality abort.** The test — does any neighbour's `(horz, vert)` sort below
the input's — is evaluated before insertion, and a thread that finds one stores
into a per-cluster `abort` byte immediately. Threads check the flag when they
pick up a frontier entry. Compared with today's instant `return`, you can lose up
to the remainder of one level. Non-canonical clusters account for 19% of
expansions in total, so the worst case is bounded and small.

**CSR adjacency.** Record `(nodeIndex -> neighbour indexes)` during the forward
pass exactly as `m_Edges`/`m_EdgeStart` already do, so the backward pass never
regenerates a move or hashes anything. On a GPU the tradeoff is sharper than on a
CPU: regenerating costs one random table probe per edge, while CSR costs a 4-byte
sequential write plus a 4-byte read. CSR wins on bandwidth but costs ~10.9 × 4
bytes per state, and that memory is what caps batch size — so keep CSR for the
large tiers and regenerate for the small ones.

### Stage 3 — backward BFS

Multi-source BFS from every goal state, over the CSR built in Stage 2. Textbook
GPU BFS; same batched level-synchronous loop. Moves are reversible, so the graph
is undirected and no transpose is needed.

The "hardest state" reduction — furthest from a goal, ties broken by
lexicographically smallest `(horz, vert)` — becomes a final min-reduction over
nodes at `maxDistance`. It is order-independent (BFS distances are final on first
assignment, and min over a total order is associative), so the GPU picks the same
state as the CPU, bit for bit.

### Stage 4 — minimality

This is 24% of the time today and cannot be left on the host at 7x7 scale. It
also does not need a separate solver.

`Solver::Solve(m_Unsolved)` recomputes, by unbounded iterative deepening, a
shortest solution that the backward pass has already implicitly found. Its only
use is the set of pieces that move. Walking from the hardest state to a goal by
always stepping to a neighbour with `distance - 1` yields a shortest solution by
construction, in `maxDistance` steps with no search at all.

The verdict is unchanged even though the path may differ. If piece *i* moves in
*some* optimal solution, deleting its moves from that sequence leaves every other
move legal — removing a piece only frees cells — and reaches the goal in strictly
fewer moves. So every piece that moves on *any* shortest path is provably
non-blocking and correctly skipped. Verified empirically: over 62,106 clusters at
5x5 and 33,370 at 6x6, the descent-based version produced **zero** differing
verdicts and the identical number of removal tests.

That leaves the removal tests, which are `SolvableWithin(reduced, maxDistance-1)`
— a bounded reachability question on a slightly different board. Express it as
Stage 2 with a level cap and an early goal test, and the separate recursive
`Solver` disappears from the GPU code entirely. Every remaining piece of the
program is one batched bounded BFS.

Cost: 7.8% of inputs reach this stage with `maxDistance > 0` — over half of
solvable clusters are already solved — each spawning up to `numPieces` bounded
searches. Measured 20,540 searches per 244,236 inputs, so it roughly doubles the
number of BFS work items. Same kernel, bigger queue.

### Stage 5 — reduce and report

Distance histogram by `atomicAdd` into the cluster's bin array; compact the
surviving puzzles; copy back sorted by id. Output volume is tiny (one line per
~500 inputs), so this stage is free.

## Data structure changes

**Replace `PieceInfo`'s lookup tables with arithmetic.** `PieceInfo` currently
holds `mask`, `firstCell` and `lastCell` for every offset: 3 × `BoardSize` × 8
bytes per piece, about 3 KB per cluster at 6x6. That is fatal per-thread state on
a GPU, and it is unnecessary — every entry is one shift of a single unit mask:

```
mask[k]      = unit << (k * stride)
firstCell[k] = low  << (k * stride)        low  = unit & -unit
lastCell[k]  = high << (k * stride)        high = low << ((size - 1) * stride)
```

for both orientations (`unit` is `((1<<size)-1) << base` horizontally, and the
column pattern `Σ 1<<(j*BoardSize)` vertically). So a piece needs `unit`,
`stride`, `size`, `maxOffset` — and `base`, `size`, `horz` in two bytes is enough
to rebuild `unit` from two constants in `__constant__` memory. The whole cluster
descriptor drops from ~3 KB to ~40 bytes, which fits in registers or a shared
slab and can be broadcast across a warp.

**Mixed-radix state packing.** `Cluster::State` spends a fixed
`OffsetBits = 3` per piece: 54 bits at 6x6, 72 bits at 7x7, which forces
`unsigned __int128` and its multi-instruction arithmetic. Packing in mixed radix
instead — piece *i* contributes `offset_i × Π_{j<i}(maxOffset_j + 1)` — puts most
6x6 clusters inside **32 bits** (twelve size-2 pieces give 5^12 ≈ 2^27.9),
halving key traffic and turning every key operation into full-rate 32-bit IMAD.
Clusters whose radix product overflows 32 bits select a 64-bit kernel
specialization; the choice is a per-cluster property known in Stage 1, so it
never becomes a per-thread branch.

The same product is a free upper bound on the cluster's state count, which is
what Stage 1 uses to size the hash segment and pick a tier. It is loose for large
boards but tight enough for the 92% of clusters that are small — exactly where a
conservative allocation would hurt most.

**Drop the masks from `Node`.** Today a node is `{state, horz, vert, distance}` =
28 bytes. `horz`/`vert` are needed only for the canonicality test and the
hardest-state tie-break. The first is evaluated on the generated child, where the
parent's masks are already in registers and the child's follow from one
`ANDN`/`OR` pair, as `ForEachMove` already does. The second is a one-off
reduction that can recompute masks from the state. So a node is `state` plus a
one-byte distance, and per-state memory falls by roughly 3x — which directly buys
a larger batch.

## Load balancing

Three tiers, assigned in Stage 1 from the radix bound, with overflow re-queued
rather than predicted:

| tier | segment | clusters | of the work | mapping |
|---|---|---|---|---|
| A | 512 slots, global slab | 95.4% | 19.9% | thread per cluster |
| B | 16k slots, shared memory | 4.4% | 46.9% | warp or block per cluster |
| C | global, batched frontier | 0.25% | 33.3% | whole grid, many clusters |

Tier A overflows into B, B into C. Because tier C uses the flat-frontier kernel,
it is self-balancing and needs no size prediction at all — which means tiers A
and B are pure optimizations, and a correct implementation can consist of tier C
alone. That is the right order to build it in.

## Memory budget

Per cluster at 6x6, in the batched tier: nodes at ~9 bytes, table at ~1.33 slots
per state × 8 bytes, CSR at 10.9 × 4 bytes per state. Call it 60 bytes per state,
so ~10 KB for an average cluster (164.7 states) and ~28 MB for the largest
observed. A batch of 100k average clusters is under 1 GB, leaving room for the
outliers to be processed a few at a time. Batch size is the one real tuning knob: it trades
memory against per-level parallelism, and 5M edge-generations per level is
already enough to saturate a large GPU.

## Expected speedup

The forward pass reduces to random probes into a table. The CPU manages 3.6M
expansions/s single-threaded, so at 10.9 neighbours each, ~40M neighbour
insertions/s — ~25 ns apiece, which is one to two cache misses per insertion
against tables of up to a few MB. A GPU with ~3 TB/s of HBM achieving 20–25% of
peak on scattered 32-byte accesses lands somewhere around 10^10 insertions/s,
i.e. two to three orders of magnitude over one core.

Discount that heavily — for warp divergence, atomic contention, the abort
granularity loss, tail clusters, and the fact that the CPU's small clusters live
in L2 — and **20–50x versus a well-used multicore CPU** is a defensible
expectation. It is a bandwidth-and-latency-bound irregular workload, not a
compute-bound one, so this is a memory-system win, not a FLOPs win.

What that buys, extrapolated: the full 6x6 solve is ~4 core-hours (13.5 s of
exploration per 1-in-997 sample). 7x7 has 2,305x more inputs, and per-input cost
grew 9.3x from 5x5 to 6x6 (6.0 µs to 55.3 µs), so 7x7 is plausibly ~10^5
core-hours — two months on a 64-core machine, which is presumably why it is
commented out in `config.h`. At 30x, one GPU brings that to a few days.

Treat the 7x7 figure as an order-of-magnitude estimate; the per-input growth
factor is the weak link, and the honest way to pin it down is to solve 7x7
partially and measure.

## Risks

- **Divergence in move generation is the main unknown.** The two-phase expansion
  should handle it, but the slide-length loop is data-dependent and cannot be
  fully flattened.
- **Tail clusters.** A 460k-state cluster at 6x6 is fine; the largest 7x7 cluster
  is unknown and could be orders of magnitude bigger. The batched-frontier design
  handles it as long as it fits in memory, and it needs a spill path if it does
  not.
- **64-bit keys at 7x7.** `MaxPieces` is 24 there, so the fixed-width state needs
  128 bits. The mixed-radix scheme is what avoids this, and its effectiveness at
  7x7 is unmeasured — worth checking the actual maximum piece count before
  committing.
- **This is not a good SIMT workload in the abstract.** Every step is a
  dependent, unpredictable memory access. GPUs win here through concurrency, not
  through the usual mechanisms, so the speedup is sensitive to getting occupancy
  and batch size right.

## Three things worth doing to the CPU code first

Independent of any GPU work, and each measured above:

1. **Delete the `Solver::Solve` call in the minimality check** and recover the
   shortest solution by descending the distance field. Verified identical
   verdicts over 95,476 clusters; worth ~13% of exploration time.
2. **Parallelize the enumeration** instead of replicating it in every worker.
   Free at 6x6, but at 7x7 it is a ~5 h floor on wall-clock time regardless of
   `NumWorkers`.
3. **Skip the cheap canonicality pre-filter.** A two-move lookahead — no hash
   table, no memory — proves 40.0% of 6x6 inputs non-canonical, which sounds
   compelling. It is not: the clusters it catches are the cheap ones, and
   non-canonical clusters only account for 19% of expansions in total. Measured
   and recorded here so nobody else pursues it. (One move alone rejects nothing:
   `PositionEntry::Require` already guarantees it.)

## Wanting only the top N

If the goal is the N hardest 7x7 puzzles for N in the 10k–100k range, rather than
a complete database, the pipeline changes — but the change is worth about 2x, not
orders of magnitude, and the two shortcuts that would give orders of magnitude
both measure as dead ends. Those negative results are the useful part of this
section.

### What the threshold is worth

Sizing first. At 6x6 there are ~0.47M minimal puzzles in total, distributed like
this in the tail (1-in-997 sample, so counts are scaled):

| threshold | puzzles at or above |
|---|---|
| ≥ 15 moves | ~114,000 |
| ≥ 16 moves | ~88,000 |
| ≥ 20 moves | ~25,000 |
| ≥ 23 moves | ~15,000 |
| ≥ 24 moves | ~9,000 |

So top-100k sits at about 15–16 moves and top-10k at about 23–24. The 100k figure
rests on ~100 sampled puzzles and is reliable; the 10k figure rests on ~10 and is
not, and a 1-in-997 sample will not contain the genuinely hardest puzzles at all.
Either way, N of 10k–100k lands deep in the tail.

Which is where the temptation lies, because the work is not there:

| clusters with eccentricity ≥ T | share of forward expansions |
|---|---|
| ≥ 5 | 58.5% |
| ≥ 10 | 17.3% |
| ≥ 15 | 4.1% |
| ≥ 20 | 0.81% |
| ≥ 25 | 0.16% |

Over 99% of the exploration goes into clusters that a top-10k run will discard.
If you could identify them in advance you would be done in an afternoon.

### Two dead ends

**There is no sound prune.** Eccentricity is a global property of the cluster: it
is the maximum over all states of the distance to the nearest goal, so you cannot
know it without enumerating the whole connected component. Nor can a smaller
layout bound a larger one. Removing a piece never lengthens the optimal solution
— that is what makes the minimality test work — so a sub-layout's difficulty is a
*lower* bound on its supersets, never an upper one. Pruning needs an upper bound.
Adding a piece can raise the difficulty arbitrarily, so no upper bound exists.

**No cheap layout feature predicts hardness.** This one had to be measured rather
than argued. Four features computable from the layout alone, with no search, each
tested as a one-sided cut against the hard clusters at 6x6 (446,808 layouts, 951
with eccentricity ≥ 15, 166 with ≥ 20):

| feature | best usable cut | expansions kept | recall at ≥15 | recall at ≥20 |
|---|---|---|---|---|
| piece count | ≥ 9 | 94.2% | 99.4% | 100% |
| piece count | ≥ 10 | 82.3% | 95.9% | 99.4% |
| occupancy | ≥ 22 | 78.6% | 95.8% | 98.8% |
| vertical pieces left of target | ≥ 2 | 90.4% | 97.8% | 99.4% |
| positional freedom `Σ log2(maxOffset+1)` | ≥ 20 | 88.8% | 97.2% | 100% |

The best trade on offer discards 18–21% of the work for a 4% loss of the puzzles
you were looking for. Every feature's distribution over hard clusters has
essentially the same shape as its distribution over all clusters — hard 6x6
puzzles have 10–13 pieces and 25–29 occupied cells, which is simply where most
layouts are. Hardness is not visible in coarse structure. Combining the features
is unlikely to rescue this given how weak each is alone.

A related intuition also fails: at 5x5 the largest clusters are the easy ones
(the biggest, 4,768 states, needs 3 moves), which suggests capping cluster size.
At 6x6 the correlation reverses — mean cluster size climbs from 37 states at
eccentricity 0 to ~11,000 at 24, and the largest cluster containing a ≥15-move
puzzle has 74,266 states. A cap that keeps every hard puzzle saves ~5% of the
work.

### What the threshold actually buys

Per-cluster, once a threshold `T` is active:

- **Skip the minimality check unless the cluster clears `T`.** 24% of exploration
  time today, and at a tail threshold it becomes a rounding error.
- **Cap the backward pass at `T` levels** and skip the distance histogram, the
  hardest-state extraction and the tie-break for everything below. Most of the
  remaining 5%.
- The forward pass survives intact, because that is the part that cannot be
  avoided.

That leaves roughly the forward pass alone: about a 1.3x saving. Combine it with
the goal-position enumeration below and the total is ~2x versus building the full
database.

### Enumerate goal positions instead

This is the largest single win available, it is not specific to top-N, and it is
provable rather than heuristic. Instead of enumerating every packed layout, pin
the primary piece on the target and pack only the other pieces.

It is sound and complete for solvable clusters. Every solvable cluster contains at
least one goal state; take the lexicographically smallest. If any non-primary
piece in it could slide to a lower offset, the result is still a goal state — the
primary has not moved — and has a smaller `(horz, vert)` key, contradicting
minimality. So the lex-smallest goal state has every non-primary piece at offset
zero or blocked from behind, which is exactly what `PositionEntry::Require`
encodes. Canonicality is then decided among the cluster's goal states only.

Measured, with the goal enumeration and goal-mode canonicality both implemented:

| | packed (today) | goal positions |
|---|---|---|
| positions emitted, 6x6 | 243,502,786 | 88,914,655 |
| distinct solvable clusters found | 33.27M | 33.14M |
| minimal puzzles found | ~0.48M | ~0.46M |
| expansions on non-canonical clusters | 19.6% | 13.4% |
| expansions on canonical-but-unsolvable | 17.6% | **0%** |
| expansions per distinct solvable cluster | 1000.9 | **706.3** |
| seconds per 1M distinct solvable clusters | 441.0 | **339.9** |

At 5x5, where a full run is cheap enough to check exactly, the two enumerations
agree to the last puzzle: 62,106 solvable clusters and 1,730 minimal puzzles from
either, out of 268,109 packed positions versus 124,886 goal positions. The 6x6
figures differ by a few percent only because the two runs sample different
1-in-997 subsets.

Every enumerated position is solvable by construction, which is what removes the
canonical-but-unsolvable clusters — a fifth of the expansions today, and pure
waste for any purpose.

### What this means for the GPU design

Less than it means for the CPU. The primitive is unchanged: a bounded BFS is
exactly what a threshold test wants. Two simplifications:

- No distance histograms and no per-cluster minimality state, so per-state memory
  drops and batches get bigger.
- The minimality stage nearly vanishes, removing a kernel path — though the
  separate `Solver` port was already unnecessary given the gradient-descent
  result above.

Tier C still dominates and still needs the batched-frontier design. The hardest
6x6 puzzles live in clusters averaging 3,000–11,000 states, squarely in tiers B
and C, so the skew handling is not optional even for a top-N run.

### The strategic choice

The measurements do not pick between these; the guarantee you want does.

1. **Provable top-N.** Full sweep of the goal enumeration with the threshold
   active. No shortcuts beyond the ~2x, because none exist. This is the only
   route that can claim these *are* the N hardest. On a GPU at 20–50x, plausibly
   days rather than months.
2. **N very hard puzzles, no guarantee.** Stochastic search in position space.
   Difficulty is monotone under adding pieces, which makes hill-climbing
   well-behaved: grow a board, keep what gets harder. `anneal.go` and
   `generator.go` already do this. Hours, no GPU, no completeness claim — and for
   most uses of a puzzle list, indistinguishable from option 1.
3. **Hybrid.** Anneal first to establish a high `T` cheaply, then sweep with that
   `T` active to certify it. The sweep costs the same as option 1, but you have a
   usable list on day one and a proof later.

Option 2 is much cheaper than option 1 and the gap is not closable, so the
decision is really about whether the word "top" has to be literally true.

## Methodology

Numbers come from an instrumented copy of `cpp/src` at `claude-performance`,
built with `-O3 -march=native -DNDEBUG`, single-threaded. 5x5 figures are a full
run over all 268,109 inputs. 6x6 figures sample the enumeration at 1-in-199
(1,223,633 clusters) or 1-in-997 where noted; the enumerator itself always runs
to completion, so enumeration timings are exact. Counters added: forward
expansions and nodes, backward relaxations, recorded edges, per-phase wall time,
node-count and eccentricity histograms, and a shadow implementation of the
minimality check for the verification described in Stage 4.

For the top-N section, `Enumerator` gained a goal mode restricting the primary
row to the single entry with the primary on the target and clearing its
`Require`, and `Cluster` gained a matching mode that applies the canonicality
comparison only to states satisfying `IsSolved`. The layout-feature study
computes piece count, `popcount(Mask())`, the number of vertical pieces in
columns left of the target, and `Σ log2(maxOffset+1)` before exploring, then
cross-tabulates each against the cluster's eccentricity.
