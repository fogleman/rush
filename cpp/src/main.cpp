#include <atomic>
#include <chrono>
#include <functional>
#include <iomanip>
#include <iostream>
#include <mutex>
#include <thread>
#include <vector>

#include "board.h"
#include "cluster.h"
#include "config.h"
#include "enumerator.h"
#include "solver.h"

using namespace std;
using namespace std::chrono;

typedef std::function<void(uint64_t id, const Cluster &)> CallbackFunc;

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

void worker(const int wi, const int wn, Counts &counts, CallbackFunc func) {
    Enumerator enumerator;
    // reused across clusters: its arena and hash table are the expensive part
    Cluster cluster;
    enumerator.Enumerate([&](uint64_t id, const Board &board) {
        if (id % wn != wi) {
            return;
        }
        cluster.Explore(board);
        bump(counts.in);
        if (cluster.Canonical()) bump(counts.canonical);
        if (cluster.Solvable()) bump(counts.solvable);
        if (cluster.Minimal()) bump(counts.minimal);
        if (!cluster.Canonical() || !cluster.Solvable() || !cluster.Minimal()) {
            return;
        }
        func(id, cluster);
    });
}

int main() {
    // uint64_t lastID = 0;
    // Enumerator enumerator;
    // enumerator.Enumerate([&](uint64_t id, const Board &board) {
    //     lastID = std::max(lastID, id);
    // });
    // cout << lastID << endl;
    // return 0;

    mutex m;

    const int wn = NumWorkers;
    std::vector<Counts> counts(wn);
    uint64_t maxSeenID = 0;

    auto start = steady_clock::now();

    // only called for clusters that yield a puzzle
    auto callback = [&](uint64_t id, const Cluster &c) {
        lock_guard<mutex> lock(m);

        const Totals totals = sum(counts);
        maxSeenID = std::max(maxSeenID, id);
        const Board &unsolved = c.Unsolved();
        const double pct = (double)maxSeenID / (double)MaxID;
        const double hrs = duration<double>(steady_clock::now() - start).count() / 3600;
        const double est = pct > 0 ? hrs / pct : 0;

        // print results to stdout
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

        // print progress info to stderr
        cerr
            << fixed
            << pct << " pct "
            << hrs << " hrs "
            << est << " est - "
            << totals.in << " inp "
            << totals.canonical << " can "
            << totals.solvable << " slv "
            << totals.minimal << " min"
            << endl;
    };

    std::vector<std::thread> threads;
    for (int wi = 0; wi < wn; wi++) {
        threads.push_back(
            std::thread(worker, wi, wn, std::ref(counts[wi]), callback));
    }
    for (int wi = 0; wi < wn; wi++) {
        threads[wi].join();
    }

    // print final stats to stderr
    const Totals totals = sum(counts);
    const double pct = (double)maxSeenID / (double)MaxID;
    const double hrs = duration<double>(steady_clock::now() - start).count() / 3600;
    const double est = pct > 0 ? hrs / pct : 0;
    cerr
        << fixed
        << 1.0 << " pct "
        << hrs << " hrs "
        << est << " est - "
        << totals.in << " inp "
        << totals.canonical << " can "
        << totals.solvable << " slv "
        << totals.minimal << " min"
        << endl;
    return 0;
}

int main2() {
    // // 51 83 13 BCDDE.BCF.EGB.FAAGHHHI.G..JIKKLLJMM. 4780
    Board board("BCDDE.BCF.EGB.FAAGHHHI.G..JIKKLLJMM.");

    Solver solver;
    for (int i = 0; i < 100; i++) {
        solver.Solve(board);
    }
    // Solver solver(board);
    // const int numMoves = solver.Solve();
    // cout << numMoves << endl;

    // // 15 32 12 BB.C...D.CEE.DAAFGH.IIFGH.JKK.LLJ... 541934
    // Board board("BB.C...D.CEE.DAAFGH.IIFGH.JKK.LLJ...");

    // // 24 43 13 B..CDDBEEC.F.G.AAF.GHHIJKKL.IJ..L.MM 278666
    // // Board board("B..CDDBEEC.F.G.AAF.GHHIJKKL.IJ..L.MM");

    // Cluster cluster(board);

    // cout << "canonical: " << cluster.Canonical() << endl;
    // cout << "solvable:  " << cluster.Solvable() << endl;
    // cout << "states:    " << cluster.NumStates() << endl;
    // cout << "moves:     " << cluster.NumMoves() << endl;
    // cout << "counts:    ";

    // for (int count : cluster.DistanceCounts()) {
    //     cout << count << ",";
    // }
    // cout << endl;
    // cout << endl;

    // cout << "unsolved:" << endl;
    // cout << cluster.Unsolved().String2D() << endl;
    // cout << "solved:" << endl;
    // cout << cluster.Solved().String2D() << endl;

    return 0;
}
