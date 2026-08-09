#include <algorithm>
#include <atomic>
#include <chrono>
#include <cstdio>
#include <cstdlib>
#include <functional>
#include <iomanip>
#include <iostream>
#include <mutex>
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
    // Shard S of N takes a 1/N sample of the row-combination space, so a run can
    // be spread over machines that never talk to each other. The shards partition
    // the space exactly, and each one samples all of it -- see Scramble.
    uint64_t shard = 0;
    uint64_t numShards = 1;
    int numWorkers = NumWorkers;
};

void Usage() {
    cerr <<
        "usage: main [options]\n"
        "  --shard S/N   process only shard S of N (default 0/1)\n"
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

// A fixed permutation of the row-combination space, so the sweep visits it in a
// scattered order instead of front to back.
//
// Cost per combination varies by more than 65x, and the expensive combinations
// are not spread evenly through the index space. A row's entry index 0 is the
// empty row, and row BoardSize-1 is the odometer's most significant digit, so the
// emptiest boards -- the largest clusters, the longest solutions, the most
// positions per combination -- all sit at the very start. Swept front to back, a
// run spends its first hours in the most expensive corner and every rate it
// reports is wrong by more than an order of magnitude. Permuted, any prefix of a
// shard is a uniform sample of the whole space, so the projection is meaningful
// within minutes -- and so is anything else measured from a partial run.
//
// Cycle walking over a bit mixer: iterating a bijection on the smallest power of
// two that covers the space, until the value lands back inside the space, is
// itself a bijection on the space, and takes 2^bits / size iterations on average
// (1.8 at 7x7). Every step -- multiply by an odd constant, xor a right shift -- is
// invertible modulo 2^bits, so the mixer is. Fixed constants and 64-bit
// arithmetic, so every machine walks the same order.
class Scramble {
public:
    explicit Scramble(const uint64_t size) : m_Size(size) {
        while (m_Bits < 64 && ((uint64_t)1 << m_Bits) < size) {
            m_Bits++;
        }
        m_Mask = m_Bits >= 64 ? ~(uint64_t)0 : (((uint64_t)1 << m_Bits) - 1);
        // a shift of zero would collapse the xor step, and with it the bijection
        m_ShiftA = std::max(1, m_Bits / 2);
        m_ShiftB = std::max(1, m_Bits / 3);
    }

    uint64_t operator()(uint64_t i) const {
        do {
            // Offset before mixing, because zero is a fixed point of multiplies
            // and shifts alike, and combination zero -- every row empty -- is the
            // most expensive one there is. It belongs inside the loop: cycle
            // walking is only a bijection while the walk starts inside the space,
            // and offsetting beforehand would start it anywhere, letting two
            // indexes reach the same combination.
            i = (i + 0x2545f4914f6cdd1dULL) & m_Mask;
            i = (i * 0x9e3779b97f4a7c15ULL) & m_Mask;
            i ^= i >> m_ShiftA;
            i = (i * 0xbf58476d1ce4e5b9ULL) & m_Mask;
            i ^= i >> m_ShiftB;
            i = (i * 0x94d049bb133111ebULL) & m_Mask;
        } while (i >= m_Size);
        return i;
    }

private:
    uint64_t m_Size;
    int m_Bits = 0;
    uint64_t m_Mask = 0;
    int m_ShiftA = 1;
    int m_ShiftB = 1;
};

// Hands out work indexes. Index k means the combination Scramble(k), so a
// contiguous range of indexes is a scattered sample of the board space.
//
// Shard S of N owns the index range [S*size/N, (S+1)*size/N). The ranges tile the
// index space and the permutation is a bijection, so the shards cover every
// combination exactly once -- while each one still samples the whole space, which
// is what a contiguous range of raw combination indexes could never do.
//
// Workers claim a batch at a time purely to keep the counter off the critical
// path: half of all combinations are now retired by the vertical-symmetry check
// in a few hundred cycles, which is not much more than a contended atomic. Unlike
// a chunk size, the batch does not affect which combinations a shard covers, so
// shards need not agree on it.
class Sweep {
public:
    static const uint64_t BatchSize = 256;

    Sweep(const Options &opts, const uint64_t numRowCombos) :
        m_First(Split(numRowCombos, opts.shard, opts.numShards)),
        m_Last(Split(numRowCombos, opts.shard + 1, opts.numShards)),
        m_Next(m_First)
    {
    }

    // Claims the next batch, as the index range [begin, end).
    bool Next(uint64_t &begin, uint64_t &end) {
        begin = m_Next.fetch_add(BatchSize, std::memory_order_relaxed);
        if (begin >= m_Last) {
            return false;
        }
        end = std::min(m_Last, begin + BatchSize);
        return true;
    }

    void Finish(const uint64_t count) {
        m_Done.fetch_add(count, std::memory_order_relaxed);
    }

    // Fraction of this shard's combinations that are finished. An unbiased
    // estimate of the fraction of its *work* that is finished, since the
    // permutation makes what has been visited a uniform sample.
    double Fraction() const {
        if (m_Last <= m_First) {
            return 1;
        }
        return (double)m_Done.load(std::memory_order_relaxed) /
            (double)(m_Last - m_First);
    }

private:
    // The i'th of n boundaries in [0, size], i.e. floor(size * i / n) without
    // overflowing on the product. Consecutive boundaries tile the index space, so
    // shards fit together exactly whatever n is.
    static uint64_t Split(
        const uint64_t size, const uint64_t i, const uint64_t n)
    {
        return size / n * i + size % n * i / n;
    }

    const uint64_t m_First;
    const uint64_t m_Last;
    std::atomic<uint64_t> m_Next;
    std::atomic<uint64_t> m_Done{0};
};

typedef std::function<void(const Cluster &)> ReportFunc;
typedef std::function<void()> ProgressFunc;

void worker(
    Sweep &sweep, const Scramble &scramble, Counts &counts,
    const Enumerator &enumerator,
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

    uint64_t begin = 0;
    uint64_t end = 0;
    while (sweep.Next(begin, end)) {
        for (uint64_t i = begin; i < end; i++) {
            enumerator.EnumerateRowCombo(scramble(i), onPosition);
        }
        sweep.Finish(end - begin);
        progress();
    }
}

}  // namespace

int main(int argc, char **argv) {
    Options opts = Parse(argc, argv);

    const Enumerator enumerator;
    const uint64_t numRowCombos = enumerator.NumRowCombos();

    cerr
        << BoardSize << "x" << BoardSize
        << ", primary row " << PrimaryRow
        << ", min moves " << MinMoves
        << ", " << enumerator.NumRowEntries(0) << " row entries"
        << ", " << enumerator.NumColumnEntries(0) << " column entries"
        << ", " << numRowCombos << " row combinations"
        << (DoVertSymmetry ? ", vertical symmetry" : "")
        << ", shard " << opts.shard << "/" << opts.numShards
        << ", " << opts.numWorkers << " workers"
        << endl;

    const Scramble scramble(numRowCombos);
    Sweep sweep(opts, numRowCombos);
    std::vector<Counts> counts(opts.numWorkers);
    mutex m;
    const auto start = steady_clock::now();
    // Read outside the output lock by every finished batch, so it cannot just be
    // a double.
    std::atomic<double> lastReport{0};

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

    // Called once per finished batch. With a move threshold puzzles are rare
    // enough that reporting only when one turns up leaves a run silent for
    // hours, so progress also goes out on elapsed time.
    //
    // The clock is read before the lock is taken, because a batch of the cheapest
    // combinations -- the ones vertical symmetry retires outright -- takes only
    // microseconds, and every worker calls this at the end of every batch. Taking
    // the output lock each time would serialize them on it. The unlocked read can
    // be stale, which only costs a redundant lock, and the check is repeated
    // underneath.
    const ProgressFunc progress = [&]() {
        const double hrs =
            duration<double>(steady_clock::now() - start).count() / 3600;
        if (hrs - lastReport.load(std::memory_order_relaxed) < 30.0 / 3600) {
            return;
        }
        lock_guard<mutex> lock(m);
        if (hrs - lastReport.load(std::memory_order_relaxed) >= 30.0 / 3600) {
            printProgress();
        }
    };

    std::vector<std::thread> threads;
    for (int wi = 0; wi < opts.numWorkers; wi++) {
        threads.push_back(std::thread(
            worker, std::ref(sweep), std::cref(scramble), std::ref(counts[wi]),
            std::cref(enumerator), std::cref(report), std::cref(progress)));
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
