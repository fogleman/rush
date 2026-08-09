#include "cluster.h"

#include <algorithm>
#include <cassert>
#include <limits>

namespace {

const int32_t UnknownDistance = std::numeric_limits<int32_t>::max();

// The table is grow-only and shared by every cluster this instance explores.
//
// Restarting it small per cluster looks appealing, since cluster sizes are
// wildly skewed and most are tiny, but it measures slower at both 5x5 and 6x6:
// the skew cuts the other way. Growing back up rehashes every node already
// inserted, and that cost lands on the big clusters, which is where most of the
// states are -- at 6x6 one percent of clusters hold forty percent of the states.
// Sizing from the previous cluster's node count does not rescue it either.
const size_t InitialTableSize = 4096;

// Ceiling on recorded adjacency, so one enormous cluster cannot balloon a
// worker's memory. Past it the backward pass falls back to regenerating moves.
const size_t MaxEdges = 4 << 20;

// Load the table to at most 3/4, which linear probing tolerates well.
bool TableIsFull(const size_t numNodes, const size_t tableSize) {
    return (numNodes + 1) * 4 > tableSize * 3;
}

uint64_t Mix(uint64_t x) {
    x ^= x >> 33;
    x *= 0xff51afd7ed558ccdULL;
    x ^= x >> 29;
    return x;
}

// A packed state is wider than 64 bits on large boards, so fold it a word at a
// time. For the common case the loop runs once and this is just Mix.
template <class T>
uint64_t Hash(const T state) {
    uint64_t h = 0;
    for (size_t shift = 0; shift < 8 * sizeof(T); shift += 64) {
        h = Mix(h ^ (uint64_t)(state >> shift));
    }
    return h;
}

}  // namespace

void Cluster::BuildPieceInfo(const Board &input) {
    m_NumPieces = input.Pieces().size();
    assert(m_NumPieces <= MaxPieces);
    for (int i = 0; i < m_NumPieces; i++) {
        const Piece &piece = input.Pieces()[i];
        PieceInfo &info = m_Pieces[i];
        info.stride = piece.Stride();
        info.size = piece.Size();
        info.horz = piece.Stride() == H;
        const int line = info.horz
            ? piece.Position() / BoardSize
            : piece.Position() % BoardSize;
        info.base = info.horz ? line * BoardSize : line;
        info.maxOffset = BoardSize - info.size;
        for (int k = 0; k <= info.maxOffset; k++) {
            bb mask = 0;
            for (int j = 0; j < info.size; j++) {
                mask |= (bb)1 << (info.base + (k + j) * info.stride);
            }
            info.mask[k] = mask;
            info.firstCell[k] = (bb)1 << (info.base + k * info.stride);
            info.lastCell[k] = (bb)1 << (info.base + (k + info.size - 1) * info.stride);
        }
    }
}

Cluster::State Cluster::StartState(const Board &input) const {
    State state = 0;
    for (int i = 0; i < m_NumPieces; i++) {
        const PieceInfo &info = m_Pieces[i];
        const int offset = (input.Pieces()[i].Position() - info.base) / info.stride;
        state = WithOffset(state, i, offset);
    }
    return state;
}

Board Cluster::ToBoard(const State state) const {
    // Built from scratch rather than by walking pieces to their new offsets one
    // at a time: the intermediate configurations of such a walk can overlap,
    // which corrupts the board's incrementally maintained masks.
    Board board;
    for (int i = 0; i < m_NumPieces; i++) {
        const PieceInfo &info = m_Pieces[i];
        const int position = info.base + Offset(state, i) * info.stride;
        board.AddPiece(Piece(position, info.size, info.stride));
    }
    return board;
}

template <class F>
bool Cluster::ForEachPieceMove(const Node &node, const int i, F fn) const {
    const PieceInfo &info = m_Pieces[i];
    if (info.size == 1) {
        // a wall never moves
        return true;
    }
    const bb all = node.horz | node.vert;
    const int offset = Offset(node.state, i);
    // the board with this piece lifted off it
    const bb restHorz = info.horz ? (node.horz & ~info.mask[offset]) : node.horz;
    const bb restVert = info.horz ? node.vert : (node.vert & ~info.mask[offset]);
    const auto visit = [&](const int k) {
        return fn(
            WithOffset(node.state, i, k),
            info.horz ? (restHorz | info.mask[k]) : restHorz,
            info.horz ? restVert : (restVert | info.mask[k]));
    };
    // Offsets are bounded by the piece's own line, so sliding needs no edge
    // masks: only the cell being entered has to be free.
    for (int k = offset - 1; k >= 0 && (all & info.firstCell[k]) == 0; k--) {
        if (!visit(k)) {
            return false;
        }
    }
    for (int k = offset + 1;
         k <= info.maxOffset && (all & info.lastCell[k]) == 0; k++) {
        if (!visit(k)) {
            return false;
        }
    }
    return true;
}

template <class F>
bool Cluster::ForEachMove(const Node &node, F fn) const {
    for (int i = 0; i < m_NumPieces; i++) {
        if (!ForEachPieceMove(node, i, fn)) {
            return false;
        }
    }
    return true;
}

// Retires every entry currently in the table: bumping the stamp is what makes
// reusing the buffer across clusters, and growing within one, free of clearing.
void Cluster::NewGeneration() {
    if (++m_Generation == 0) {
        // The stamp wrapped, so entries left by older clusters can no longer be
        // told apart from fresh ones. Clearing outright is the only way back,
        // and it happens once every four billion generations.
        std::fill(m_Table.begin(), m_Table.end(), 0);
        m_Generation = 1;
    }
}

void Cluster::BeginCluster() {
    m_Nodes.clear();
    m_Edges.clear();
    m_EdgeStart.clear();
    m_HaveEdges = true;
    NewGeneration();
    if (m_Table.empty()) {
        GrowTable();
    }
}

void Cluster::GrowTable() {
    const size_t size = m_Table.empty() ? InitialTableSize : m_Table.size() * 2;
    // assign zeroes the buffer, which retires every entry already in it
    m_Table.assign(size, 0);
    m_TableMask = size - 1;
    for (size_t i = 0; i < m_Nodes.size(); i++) {
        uint64_t slot = Hash(m_Nodes[i].state) & m_TableMask;
        while ((m_Table[slot] >> 32) == m_Generation) {
            slot = (slot + 1) & m_TableMask;
        }
        m_Table[slot] = ((uint64_t)m_Generation << 32) | (i + 1);
    }
}

uint32_t Cluster::Insert(const State state, const bb horz, const bb vert) {
    if (TableIsFull(m_Nodes.size(), m_Table.size())) {
        GrowTable();
    }
    uint64_t slot = Hash(state) & m_TableMask;
    while (true) {
        const uint64_t entry = m_Table[slot];
        if ((entry >> 32) != m_Generation) {
            // Node indexes are stored biased by one so that a zeroed table
            // reads as empty; a cluster can therefore hold 2^32-1 states.
            assert(m_Nodes.size() + 1 < UINT32_MAX);
            m_Nodes.push_back({state, horz, vert, UnknownDistance});
            m_Table[slot] = ((uint64_t)m_Generation << 32) | m_Nodes.size();
            return (uint32_t)(m_Nodes.size() - 1);
        }
        const uint32_t index = (uint32_t)entry - 1;
        if (m_Nodes[index].state == state) {
            return index;
        }
        slot = (slot + 1) & m_TableMask;
    }
}

// Looks up a state that is known to be present. Every slot probed on the way to
// it was filled by this cluster, because Insert walked the same chain and
// stopped at the first slot this cluster had not filled.
uint32_t Cluster::Find(const State state) const {
    uint64_t slot = Hash(state) & m_TableMask;
    while (true) {
        const uint64_t entry = m_Table[slot];
        assert((entry >> 32) == m_Generation);
        const uint32_t index = (uint32_t)entry - 1;
        if (m_Nodes[index].state == state) {
            return index;
        }
        slot = (slot + 1) & m_TableMask;
    }
}

void Cluster::Explore(const Board &input) {
    m_Canonical = false;
    m_Solvable = false;
    m_Minimal = false;
    m_NumStates = 0;
    m_Radius = 0;
    m_MaxDistance = 0;
    // reset rather than leave the previous cluster's board behind: this
    // instance is reused, and results that are not filled in below should read
    // as empty rather than as stale
    m_Unsolved = Board();

    BuildPieceInfo(input);
    BeginCluster();

    const bb inputHorz = input.HorzMask();
    const bb inputVert = input.VertMask();
    const State startState = StartState(input);
    // Goal enumeration always seeds from a goal state; anything else only
    // reaches here if some other caller hands over an arbitrary board, and the
    // radius shortcut below is not sound for those.
    const bool seedIsGoal = IsSolved(startState);
    Insert(startState, inputHorz, inputVert);

    // Forward pass: reach every state in the cluster. If any *goal* state sorts
    // before the input then the input is not this cluster's canonical
    // representative, and whichever goal state is will report the cluster
    // instead, so there is nothing left to do here. Where vertical symmetry
    // applies, mirrored goal states join that comparison, so the representative
    // is chosen once for a cluster and its mirror together.
    //
    // The test is restricted to goal states because the enumeration only emits
    // goal states: the input is one, and the state that would sort below it has
    // to be a candidate the enumeration could have produced. Comparing against
    // every state instead would discard clusters nothing else reports.
    //
    // Nodes are visited in the order they were inserted, which is breadth-first
    // order, so the levels are contiguous runs and tracking where the current
    // one ends is all it takes to know the radius.
    size_t levelEnd = 1;
    for (size_t i = 0; i < m_Nodes.size(); i++) {
        if (i == levelEnd) {
            m_Radius++;
            levelEnd = m_Nodes.size();
        }
        // by value: inserting below can reallocate m_Nodes
        const Node node = m_Nodes[i];
        if (IsSolved(node.state)) {
            m_Solvable = true;
        }
        // hoisted out of the loop below, where it is a perfectly predicted branch
        const bool recordEdges = m_HaveEdges;
        if (recordEdges) {
            m_EdgeStart.push_back((uint32_t)m_Edges.size());
        }
        const bool canonical = ForEachMove(node,
            [&](const State state, const bb horz, const bb vert)
        {
            if (IsSolved(state)) {
                if (horz < inputHorz ||
                    (horz == inputHorz && vert < inputVert)) {
                    return false;
                }
                // The mirror of this goal state is a goal state of the mirror
                // cluster, and one representative is wanted for the pair, so it
                // competes with the input on the same footing. Comparing only the
                // input against its own mirror -- which the enumerator already
                // does, and cheaply -- would not be enough: reflection does not
                // preserve the ordering, so a cluster's smallest goal state and
                // its mirror's smallest goal state can both beat their own
                // mirrors, and the pair would be reported twice.
                if (DoVertSymmetry) {
                    const bb mirrorHorz = ReflectV(horz);
                    if (mirrorHorz < inputHorz ||
                        (mirrorHorz == inputHorz && ReflectV(vert) < inputVert)) {
                        return false;
                    }
                }
            }
            const uint32_t index = Insert(state, horz, vert);
            if (recordEdges) {
                m_Edges.push_back(index);
            }
            return true;
        });
        if (!canonical) {
            // don't count a non-canonical cluster as solvable either
            m_Solvable = false;
            return;
        }
        if (m_HaveEdges && m_Edges.size() > MaxEdges) {
            // Give up on adjacency rather than grow without bound. Checked once
            // per node, so a node's run may spill slightly past the ceiling.
            m_HaveEdges = false;
        }
    }
    if (m_HaveEdges) {
        // sentinel, so node i's run always ends at m_EdgeStart[i + 1]
        m_EdgeStart.push_back((uint32_t)m_Edges.size());
    }

    m_Canonical = true;
    m_NumStates = m_Nodes.size();
    if (!m_Solvable) {
        return;
    }

    // The seed is a goal state, so no state can be further from the goal set
    // than it is from the seed: the forward radius is an upper bound on the
    // cluster's eccentricity. A cluster that cannot reach the threshold stops
    // here, with no backward pass, no hardest-state extraction, no minimality
    // test and no histogram -- which is the whole point of the threshold, since
    // the forward pass is the only part that cannot be avoided.
    if (seedIsGoal && m_Radius < MinMoves) {
        return;
    }

    // Backward pass: breadth-first from every goal state at once, which gives
    // each state its distance to the nearest goal.
    m_Queue.clear();
    for (size_t i = 0; i < m_Nodes.size(); i++) {
        if (IsSolved(m_Nodes[i].state)) {
            m_Nodes[i].distance = 0;
            m_Queue.push_back(i);
        }
    }

    // the input itself stands in until something further away turns up
    uint32_t hardestIndex = 0;
    State hardest = m_Nodes[0].state;
    bb hardestHorz = inputHorz;
    bb hardestVert = inputVert;

    const auto relax = [&](const uint32_t index, const int32_t distance) {
        Node &neighbor = m_Nodes[index];
        if (neighbor.distance <= distance) {
            return;
        }
        neighbor.distance = distance;
        m_Queue.push_back(index);
        if (distance > m_MaxDistance) {
            m_MaxDistance = distance;
            hardestIndex = index;
            hardest = neighbor.state;
            hardestHorz = neighbor.horz;
            hardestVert = neighbor.vert;
        } else if (distance == m_MaxDistance) {
            if (neighbor.horz < hardestHorz ||
                (neighbor.horz == hardestHorz && neighbor.vert < hardestVert)) {
                hardestIndex = index;
                hardest = neighbor.state;
                hardestHorz = neighbor.horz;
                hardestVert = neighbor.vert;
            }
        }
    };

    for (size_t i = 0; i < m_Queue.size(); i++) {
        const uint32_t current = m_Queue[i];
        const int32_t distance = m_Nodes[current].distance + 1;
        if (m_HaveEdges) {
            // walk the recorded adjacency: no move generation, no hashing
            const uint32_t end = m_EdgeStart[current + 1];
            for (uint32_t e = m_EdgeStart[current]; e < end; e++) {
                relax(m_Edges[e], distance);
            }
        } else {
            const Node node = m_Nodes[current];
            ForEachMove(node, [&](const State state, const bb, const bb) {
                relax(Find(state), distance);
                return true;
            });
        }
    }

    // The radius is only an upper bound, so a cluster can still fall short of
    // the threshold here. Stop before the minimality test, which is the
    // expensive part of what remains.
    if (m_MaxDistance < MinMoves) {
        return;
    }

    m_Unsolved = ToBoard(hardest);

    // A puzzle is minimal when no piece can be removed without making it easier
    // -- see BoardIsMinimal, which is one bounded search per piece rather than a
    // full iterative-deepening ladder.
    bool minimal = true;
    if (m_MaxDistance == 0) {
        // Already solved, so removing anything leaves it solved. Minimal only
        // if there is nothing to remove.
        minimal = m_NumPieces == 1;
    } else {
        // Any piece that moves in some shortest solution is provably
        // non-blocking: deleting its moves from that solution leaves every
        // other move legal, since removing a piece only frees cells, and
        // reaches the goal in strictly fewer moves. Walking down the distance
        // field the backward pass just built produces such a solution in
        // m_MaxDistance steps with no search at all.
        m_PieceMoved.assign(m_NumPieces, false);
        uint32_t current = hardestIndex;
        for (int step = 0; step < m_MaxDistance; step++) {
            const Node node = m_Nodes[current];
            const int32_t want = node.distance - 1;
            bool found = false;
            for (int i = 0; i < m_NumPieces && !found; i++) {
                ForEachPieceMove(node, i,
                    [&](const State state, const bb, const bb)
                {
                    const uint32_t index = Find(state);
                    if (m_Nodes[index].distance != want) {
                        return true;
                    }
                    m_PieceMoved[i] = true;
                    current = index;
                    found = true;
                    return false;
                });
            }
            // the distance field is a gradient: every state above zero has a
            // neighbor one step closer
            assert(found);
        }
        minimal = BoardIsMinimal(m_Unsolved);
    }

    // The mirror cluster is not explored -- that is the point of the symmetry --
    // so the puzzle it would have reported has to be settled here, and minimality
    // is a property of the one state that gets picked, not of the cluster. Two
    // states equally far from a goal can disagree about it, and the mirror
    // cluster's pick is not the reflection of this one, since the tie-break is
    // lexicographic and reflection does not preserve that order. Falling back to
    // it keeps a puzzle that the mirror would have reported on its own, which is
    // what makes the reduction exact rather than merely close: a mirror pair
    // yields a puzzle whenever either half of it would have.
    //
    // The two picks coincide -- and the fallback does nothing -- exactly when the
    // cluster is its own mirror, so a self-symmetric puzzle is never reported
    // twice or tested twice.
    if (DoVertSymmetry && !minimal && m_MaxDistance > 0) {
        // What the mirror cluster would have picked: of the states furthest from a
        // goal, the one whose reflection sorts first. Reflection preserves
        // distances, so the mirror's furthest states are exactly the reflections
        // of these. Found in one pass over the nodes, here on the fallback path
        // rather than tracked during the backward pass, where it would have cost
        // two reflections for every relaxed node to serve the rare case that needs
        // it -- and the pass costs nothing next to the searches below.
        const Node *pick = nullptr;
        bb pickHorz = 0;
        bb pickVert = 0;
        for (const Node &node : m_Nodes) {
            if (node.distance != m_MaxDistance) {
                continue;
            }
            const bb horz = ReflectV(node.horz);
            const bb vert = ReflectV(node.vert);
            if (pick == nullptr || horz < pickHorz ||
                (horz == pickHorz && vert < pickVert)) {
                pick = &node;
                pickHorz = horz;
                pickVert = vert;
            }
        }
        assert(pick != nullptr);
        if (pickHorz != hardestHorz || pickVert != hardestVert) {
            const Board board = ToBoard(pick->state).Reflected();
            // No solution has been walked for this board, so every piece has to
            // be tested; the descent above only ever skips work, it never decides
            // anything.
            m_PieceMoved.assign(m_NumPieces, false);
            if (BoardIsMinimal(board)) {
                m_Unsolved = board;
                minimal = true;
            }
        }
    }

    if (!minimal) {
        return;
    }
    m_Minimal = true;

    m_Distances.assign(m_MaxDistance + 1, 0);
    for (const Node &node : m_Nodes) {
        m_Distances[node.distance]++;
    }
}

// Removing a piece can never increase the distance to a goal -- every existing
// solution survives -- so "the distance is unchanged" is the same question as "no
// shorter solution exists", which is one bounded search per piece. Pieces flagged
// in m_PieceMoved are skipped, having already been proved non-blocking; clear it
// to test every piece.
bool Cluster::BoardIsMinimal(const Board &input) {
    for (int i = 1; i < m_NumPieces; i++) {
        if (m_PieceMoved[i]) {
            continue;
        }
        Board board(input);
        board.RemovePiece(i);
        if (!Reduced().SolvableWithin(board, m_MaxDistance - 1)) {
            return false;
        }
    }
    return true;
}

Cluster &Cluster::Reduced() {
    if (!m_Reduced) {
        m_Reduced.reset(new Cluster());
    }
    return *m_Reduced;
}

bool Cluster::SolvableWithin(const Board &input, const int maxLevels) {
    BuildPieceInfo(input);
    BeginCluster();
    // no backward pass here, so the adjacency would only cost memory
    m_HaveEdges = false;

    const State startState = StartState(input);
    if (IsSolved(startState)) {
        return true;
    }
    if (maxLevels <= 0) {
        return false;
    }
    Insert(startState, input.HorzMask(), input.VertMask());

    bool solved = false;
    const auto expand = [&](const State state, const bb horz, const bb vert) {
        if (IsSolved(state)) {
            solved = true;
            return false;
        }
        Insert(state, horz, vert);
        return true;
    };

    // Breadth-first, as in the forward pass: the levels are contiguous runs of
    // m_Nodes, so the cap is a comparison. Nodes at level maxLevels - 1 are the
    // last ones expanded, since their neighbors are exactly maxLevels moves out.
    size_t levelEnd = 1;
    int level = 0;
    for (size_t i = 0; i < m_Nodes.size(); i++) {
        if (i == levelEnd) {
            if (++level >= maxLevels) {
                break;
            }
            levelEnd = m_Nodes.size();
        }
        // by value: inserting below can reallocate m_Nodes
        const Node node = m_Nodes[i];
        if (!ForEachMove(node, expand)) {
            break;
        }
    }
    return solved;
}
