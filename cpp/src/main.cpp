#include <algorithm>
#include <atomic>
#include <chrono>
#include <cstdio>
#include <cstdlib>
#include <functional>
#include <iomanip>
#include <iostream>
#include <mutex>
#include <set>
#include <string>
#include <thread>
#include <vector>

#include "board.h"
#include "cluster.h"
#include "config.h"
#include "enumerator.h"
#include "solver.h"

using namespace std;
using namespace std::chrono;

namespace {

struct Options {
    // Shard S of N takes every Nth chunk of the row-combination space, so a run
    // can be spread over machines that never talk to each other. Every shard
    // must use the same --chunk, or the shards do not partition the space.
    uint64_t shard = 0;
    uint64_t numShards = 1;
    // Row combinations per unit of work. Zero picks a default from the size of
    // the space, which is a property of the board and so is the same on every
    // machine.
    uint64_t chunkSize = 0;
    // Sweep position, so a long run can be stopped and resumed: pass the resume
    // value from the last progress line back in.
    uint64_t resume = 0;
    uint64_t limit = 0;
    int numWorkers = NumWorkers;
};

void Usage() {
    cerr <<
        "usage: main [options]\n"
        "  --shard S/N   process only shard S of N (default 0/1)\n"
        "  --chunk C     row combinations per unit of work\n"
        "  --resume C    start at row combination C\n"
        "  --limit N     stop this many row combinations after --resume\n"
        "  --workers W   worker threads (default NumWorkers from config.h)\n";
    exit(1);
}

Options Parse(const int argc, char **argv) {
    Options o;
    for (int i = 1; i < argc; i++) {
        const string arg = argv[i];
        const char *next = i + 1 < argc ? argv[i + 1] : nullptr;
        if (arg == "--shard" && next) {
            if (sscanf(argv[++i], "%llu/%llu",
                    (unsigned long long *)&o.shard,
                    (unsigned long long *)&o.numShards) != 2 ||
                o.numShards == 0 || o.shard >= o.numShards) {
                Usage();
            }
        } else if (arg == "--chunk" && next) {
            o.chunkSize = strtoull(argv[++i], nullptr, 10);
        } else if (arg == "--resume" && next) {
            o.resume = strtoull(argv[++i], nullptr, 10);
        } else if (arg == "--limit" && next) {
            o.limit = strtoull(argv[++i], nullptr, 10);
        } else if (arg == "--workers" && next) {
            o.numWorkers = atoi(argv[++i]);
        } else {
            Usage();
        }
    }
    if (o.numWorkers < 1) {
        Usage();
    }
    return o;
}

// One instance per worker, so the increments below never touch another thread's
// cache line and never need the output lock. The totals are only summed when a
// puzzle is reported, which is rare; taking a lock for every cluster instead
// would serialize all the workers billions of times per run. Only the owning
// worker writes, so relaxed atomics cost nothing but keep the reads defined.
struct Counts {
    std::atomic<uint64_t> in{0};
    std::atomic<uint64_t> canonical{0};
    std::atomic<uint64_t> solvable{0};
    std::atomic<uint64_t> minimal{0};
    char padding[64 - 4 * sizeof(std::atomic<uint64_t>)];
};

struct Totals {
    uint64_t in = 0;
    uint64_t canonical = 0;
    uint64_t solvable = 0;
    uint64_t minimal = 0;
};

void bump(std::atomic<uint64_t> &counter) {
    counter.store(
        counter.load(std::memory_order_relaxed) + 1, std::memory_order_relaxed);
}

Totals sum(const std::vector<Counts> &counts) {
    Totals totals;
    for (const Counts &c : counts) {
        totals.in += c.in.load(std::memory_order_relaxed);
        totals.canonical += c.canonical.load(std::memory_order_relaxed);
        totals.solvable += c.solvable.load(std::memory_order_relaxed);
        totals.minimal += c.minimal.load(std::memory_order_relaxed);
    }
    return totals;
}

// Hands out chunks of the row-combination space.
//
// A chunk is a range of combination indexes and nothing more, because rows never
// conflict with one another: the space is a plain odometer, so where a chunk
// starts needs no enumeration to find. That is what makes the sweep resumable
// and shardable across machines.
//
// Chunks are interleaved -- shard S takes every Nth one -- rather than split
// into contiguous ranges. Measured cost per row combination varies by more than
// 65x across the space, so contiguous ranges give wildly unbalanced shards;
// interleaving averages it out. Workers inside a process take the next chunk
// going, which balances them for the same reason.
class Sweep {
public:
    Sweep(const Options &opts, const uint64_t numRowCombos) :
        m_ChunkSize(opts.chunkSize),
        m_Shard(opts.shard),
        m_NumShards(opts.numShards),
        m_First(std::min(opts.resume, numRowCombos)),
        m_Last(opts.limit > 0
            ? std::min(opts.resume + opts.limit, numRowCombos)
            : numRowCombos)
    {
        // The first chunk of this shard at or after the resume point. Ownership
        // is a property of the chunk index alone, so a resumed run picks up
        // exactly the chunks it would have reached.
        const uint64_t firstChunk = m_First / m_ChunkSize;
        m_BaseChunk = firstChunk +
            (m_NumShards + m_Shard - firstChunk % m_NumShards) % m_NumShards;
        if (m_Last > 0) {
            const uint64_t lastChunk = (m_Last - 1) / m_ChunkSize;
            if (lastChunk >= m_BaseChunk) {
                m_NumChunks = (lastChunk - m_BaseChunk) / m_NumShards + 1;
            }
        }
    }

    // Claims the next chunk, as the combination range [begin, end).
    bool Next(uint64_t &ordinal, uint64_t &begin, uint64_t &end) {
        const uint64_t k = m_Next.fetch_add(1, std::memory_order_relaxed);
        const uint64_t chunk = ChunkOf(k);
        const uint64_t from = std::max(m_First, chunk * m_ChunkSize);
        if (from >= m_Last) {
            return false;
        }
        ordinal = k;
        begin = from;
        end = std::min(m_Last, (chunk + 1) * m_ChunkSize);
        lock_guard<mutex> lock(m_Mutex);
        m_InFlight.insert(k);
        return true;
    }

    void Finish(const uint64_t ordinal) {
        lock_guard<mutex> lock(m_Mutex);
        m_InFlight.erase(ordinal);
        m_Finished++;
    }

    // The combination to pass back as --resume. Every chunk this shard owns
    // below it is finished, so resuming there loses at most the chunks still in
    // flight and misses nothing. It is deliberately conservative: one expensive
    // chunk holds the cursor back however far the other workers have run ahead,
    // which is why coverage is reported separately.
    uint64_t Cursor() const {
        lock_guard<mutex> lock(m_Mutex);
        const uint64_t k = m_InFlight.empty()
            ? m_Next.load(std::memory_order_relaxed)
            : *m_InFlight.begin();
        return std::min(m_Last, std::max(m_First, ChunkOf(k) * m_ChunkSize));
    }

    // Fraction of this shard's chunks that are finished. Counted rather than
    // derived from the cursor: cost per chunk varies by more than an order of
    // magnitude, so workers finish far out of order and the cursor is a poor
    // measure of how much is done.
    double Fraction() const {
        if (m_NumChunks == 0) {
            return 1;
        }
        lock_guard<mutex> lock(m_Mutex);
        return (double)m_Finished / (double)m_NumChunks;
    }

private:
    uint64_t ChunkOf(const uint64_t ordinal) const {
        return m_BaseChunk + ordinal * m_NumShards;
    }

    const uint64_t m_ChunkSize;
    const uint64_t m_Shard;
    const uint64_t m_NumShards;
    const uint64_t m_First;
    const uint64_t m_Last;
    uint64_t m_BaseChunk = 0;
    uint64_t m_NumChunks = 0;

    std::atomic<uint64_t> m_Next{0};
    mutable mutex m_Mutex;
    uint64_t m_Finished = 0;
    // Chunks claimed but not yet finished. The smallest of them is what bounds
    // the resume cursor: chunks above it may already be done, but a resumed run
    // has to start somewhere it can prove nothing was missed.
    std::set<uint64_t> m_InFlight;
};

typedef std::function<void(const Cluster &)> ReportFunc;
typedef std::function<void()> ProgressFunc;

void worker(
    Sweep &sweep, Counts &counts, const Enumerator &enumerator,
    const ReportFunc &report, const ProgressFunc &progress)
{
    // reused across clusters: its arena and hash table are the expensive part
    Cluster cluster;
    // built once, not per combination: a std::function large enough to capture
    // this much heap allocates every time it is constructed
    const EnumeratorFunc onPosition = [&](const Board &board) {
        cluster.Explore(board);
        bump(counts.in);
        if (cluster.Canonical()) bump(counts.canonical);
        if (cluster.Solvable()) bump(counts.solvable);
        if (cluster.Minimal()) bump(counts.minimal);
        if (!cluster.Canonical() || !cluster.Solvable() || !cluster.Minimal()) {
            return;
        }
        report(cluster);
    };

    uint64_t ordinal = 0;
    uint64_t begin = 0;
    uint64_t end = 0;
    while (sweep.Next(ordinal, begin, end)) {
        for (uint64_t combo = begin; combo < end; combo++) {
            enumerator.EnumerateRowCombo(combo, onPosition);
        }
        sweep.Finish(ordinal);
        progress();
    }
}

}  // namespace

int main(int argc, char **argv) {
    Options opts = Parse(argc, argv);

    const Enumerator enumerator;
    const uint64_t numRowCombos = enumerator.NumRowCombos();
    if (opts.chunkSize == 0) {
        // Small enough that the expensive stretches of the space get spread over
        // the workers and a stopped run loses little, large enough that the
        // bookkeeping disappears. Derived from the board alone, so every shard
        // agrees on the chunk grid without being told what it is.
        opts.chunkSize =
            std::max<uint64_t>(1, std::min<uint64_t>(1 << 16, numRowCombos >> 14));
    }

    cerr
        << BoardSize << "x" << BoardSize
        << ", primary row " << PrimaryRow
        << ", min moves " << MinMoves
        << ", " << enumerator.NumRowEntries(0) << " row entries"
        << ", " << enumerator.NumColumnEntries(0) << " column entries"
        << ", " << numRowCombos << " row combinations"
        << ", chunk " << opts.chunkSize
        << ", shard " << opts.shard << "/" << opts.numShards
        << ", " << opts.numWorkers << " workers"
        << endl;

    Sweep sweep(opts, numRowCombos);
    std::vector<Counts> counts(opts.numWorkers);
    mutex m;
    const auto start = steady_clock::now();
    double lastReport = 0;

    // Both of these hold the output lock, so they also serialize the progress
    // line against the puzzle being printed.
    const auto printProgress = [&]() {
        const Totals totals = sum(counts);
        const double pct = sweep.Fraction();
        const double hrs =
            duration<double>(steady_clock::now() - start).count() / 3600;
        lastReport = hrs;
        cerr
            << fixed
            << pct << " pct "
            << hrs << " hrs "
            << (pct > 0 ? hrs / pct : 0) << " est - "
            << "resume " << sweep.Cursor() << " - "
            << totals.in << " inp "
            << totals.canonical << " can "
            << totals.solvable << " slv "
            << totals.minimal << " min"
            << endl;
    };

    // only called for clusters that yield a puzzle
    const ReportFunc report = [&](const Cluster &c) {
        lock_guard<mutex> lock(m);
        const Board &unsolved = c.Unsolved();
        cout
            << setfill('0')
            << setw(2) << c.NumMoves() << " "
            << unsolved << " "
            << c.NumStates() << " ";
        for (int i = 0; i < c.DistanceCounts().size(); i++) {
            if (i != 0) {
                cout << ",";
            }
            cout << c.DistanceCounts()[i];
        }
        cout << endl;
        printProgress();
    };

    // Called once per finished chunk. With a move threshold puzzles are rare
    // enough that reporting only when one turns up leaves a run silent for
    // hours, so progress also goes out on elapsed time.
    const ProgressFunc progress = [&]() {
        lock_guard<mutex> lock(m);
        const double hrs =
            duration<double>(steady_clock::now() - start).count() / 3600;
        if (hrs - lastReport >= 30.0 / 3600) {
            printProgress();
        }
    };

    std::vector<std::thread> threads;
    for (int wi = 0; wi < opts.numWorkers; wi++) {
        threads.push_back(std::thread(
            worker, std::ref(sweep), std::ref(counts[wi]), std::cref(enumerator),
            std::cref(report), std::cref(progress)));
    }
    for (std::thread &thread : threads) {
        thread.join();
    }

    lock_guard<mutex> lock(m);
    printProgress();
    return 0;
}

int main2() {
    // // 51 83 13 BCDDE.BCF.EGB.FAAGHHHI.G..JIKKLLJMM. 4780
    Board board("BCDDE.BCF.EGB.FAAGHHHI.G..JIKKLLJMM.");

    Solver solver;
    for (int i = 0; i < 100; i++) {
        solver.Solve(board);
    }

    return 0;
}
