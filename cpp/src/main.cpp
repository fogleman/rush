#include <chrono>
#include <cstdio>
#include <cstdlib>
#include <functional>
#include <iostream>
#include <mutex>
#include <thread>

#include "board.h"
#include "cluster.h"
#include "config.h"
#include "enumerator.h"
#include "solver.h"

using namespace std;
using namespace std::chrono;

typedef std::function<void(const Cluster &)> CallbackFunc;

void worker(const int wi, const int wn, CallbackFunc func) {
    Enumerator enumerator;
    enumerator.Enumerate([&](uint64_t id, uint64_t group, const Board &board) {
        if (id % wn != wi) {
            return;
        }
        Cluster cluster(id, group, board);
        func(cluster);
    });
}

int main() {
    const uint64_t maxID = 243502785;
    uint64_t maxSeenID = 0;

    mutex m;

    auto start = steady_clock::now();

    uint64_t numIn = 0;
    uint64_t numCanonical = 0;
    uint64_t numSolvable = 0;
    uint64_t numMinimal = 0;

    auto callback = [&](const Cluster &c) {
        lock_guard<mutex> lock(m);

        numIn++;
        if (c.Canonical()) numCanonical++;
        if (c.Solvable()) numSolvable++;
        if (c.Minimal()) numMinimal++;
        if (!c.Canonical() || !c.Solvable() || !c.Minimal()) {
            return;
        }

        maxSeenID = std::max(maxSeenID, c.ID());
        const Board &unsolved = c.Unsolved();
        const double pct = (double)maxSeenID / (double)maxID;
        const double sec = duration<double>(steady_clock::now() - start).count();

        // print results to stdout
        printf(
            "%02d %02d %s %lld %lld %d ",
            c.NumMoves(),
            (int)unsolved.Pieces().size(),
            unsolved.String().c_str(),
            c.ID(),
            c.Group(),
            c.NumStates());
        for (int i = 0; i < c.DistanceCounts().size(); i++) {
            if (i != 0) {
                printf(",");
            }
            printf("%d", c.DistanceCounts()[i]);
        }
        printf("\n");

        // print progress info to stderr
        fprintf(
            stderr,
            "%.6f pct %.3f sec - %lld inp %lld can %lld slv %lld min\n",
            pct, sec, numIn, numCanonical, numSolvable, numMinimal);
    };

    std::vector<std::thread> threads;
    const int wn = 4;
    for (int wi = 0; wi < wn; wi++) {
        threads.push_back(std::thread(worker, wi, wn, callback));
    }
    for (int wi = 0; wi < wn; wi++) {
        threads[wi].join();
    }
    return 0;
}

int main2() {
    // // 51 83 13 BCDDE.BCF.EGB.FAAGHHHI.G..JIKKLLJMM. 4780
    Board board("BCDDE.BCF.EGB.FAAGHHHI.G..JIKKLLJMM.");

    for (int i = 0; i < 100; i++) {
        Solver solver(board);
        solver.Solve();
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
