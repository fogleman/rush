#include <chrono>
#include <cstdio>
#include <cstdlib>
#include <functional>
#include <iostream>
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
    auto start = steady_clock::now();
    auto callback = [&](const Cluster &c) {
        if (!c.Canonical() || !c.Solvable() || !c.Minimal()) {
            return;
        }
        maxSeenID = std::max(maxSeenID, c.ID());
        const Board &unsolved = c.Unsolved();
        const double pct = (double)maxSeenID / (double)maxID;
        const double sec = duration<double>(steady_clock::now() - start).count();
        // TODO: mutex
        printf(
            "%.6f %.3f %lld %lld %02d %02d %s %d\n",
            pct, sec, c.ID(), c.Group(), c.NumMoves(),
            (int)unsolved.Pieces().size(), unsolved.String().c_str(),
            c.NumStates());
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
