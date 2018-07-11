#include <chrono>
#include <functional>
#include <iomanip>
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
    enumerator.Enumerate([&](uint64_t id, const Board &board) {
        if (id % wn != wi) {
            return;
        }
        Cluster cluster(id, board);
        func(cluster);
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

    uint64_t maxSeenID = 0;
    uint64_t numIn = 0;
    uint64_t numCanonical = 0;
    uint64_t numSolvable = 0;
    uint64_t numMinimal = 0;

    auto start = steady_clock::now();

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
            << numIn << " inp "
            << numCanonical << " can "
            << numSolvable << " slv "
            << numMinimal << " min"
            << endl;
    };

    std::vector<std::thread> threads;
    const int wn = NumWorkers;
    for (int wi = 0; wi < wn; wi++) {
        threads.push_back(std::thread(worker, wi, wn, callback));
    }
    for (int wi = 0; wi < wn; wi++) {
        threads[wi].join();
    }

    // print final stats to stderr
    const double pct = (double)maxSeenID / (double)MaxID;
    const double hrs = duration<double>(steady_clock::now() - start).count() / 3600;
    const double est = pct > 0 ? hrs / pct : 0;
    cerr
        << fixed
        << 1.0 << " pct "
        << hrs << " hrs "
        << est << " est - "
        << numIn << " inp "
        << numCanonical << " can "
        << numSolvable << " slv "
        << numMinimal << " min"
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
