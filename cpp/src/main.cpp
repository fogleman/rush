#include <cstdlib>
#include <iostream>
#include <unordered_set>

#include "board.h"
#include "cluster.h"
#include "config.h"
#include "enumerator.h"

using namespace std;

void print(const Board &board) {
    cout << board << endl;
    // cout << BitboardString(board.Mask()) << endl;
    // cout << BitboardString(board.HorzMask()) << endl;
    // cout << BitboardString(board.VertMask()) << endl;
    // cout << endl;
}

int main() {
    Enumerator enumerator;
    // 362,797,056
    // for (int g = enumerator.NumGroups() - 1; ; g--) {
    // for (int g = 0; ; g++) {
    //     enumerator.EnumerateGroup(g, [&](uint64_t counter, int group, const Board &board) {
    //         cout << group << endl;
    //         cout << board.String2D() << endl;
    //     });
    // }
    enumerator.EnumerateGroup(11000,[&](uint64_t counter, int group, const Board &board) {
        if (group == 11000) {
            cout << group << " " << counter << endl;
            cout << board.String2D() << endl;
        }
    });
    return 0;

    // 51 83 13 BCDDE.BCF.EGB.FAAGHHHI.G..JIKKLLJMM. 4780
    // Board board("BCDDE.BCF.EGB.FAAGHHHI.G..JIKKLLJMM.");

    // 15 32 12 BB.C...D.CEE.DAAFGH.IIFGH.JKK.LLJ... 541934
    Board board("BB.C...D.CEE.DAAFGH.IIFGH.JKK.LLJ...");

    // 24 43 13 B..CDDBEEC.F.G.AAF.GHHIJKKL.IJ..L.MM 278666
    // Board board("B..CDDBEEC.F.G.AAF.GHHIJKKL.IJ..L.MM");

    Cluster cluster(board);

    cout << "canonical: " << cluster.Canonical() << endl;
    cout << "solvable:  " << cluster.Solvable() << endl;
    cout << "states:    " << cluster.NumStates() << endl;
    cout << "moves:     " << cluster.NumMoves() << endl;
    cout << "counts:    ";

    for (int count : cluster.DistanceCounts()) {
        cout << count << ",";
    }
    cout << endl;
    cout << endl;

    cout << "input:" << endl;
    cout << cluster.Input().String2D() << endl;
    cout << "unsolved:" << endl;
    cout << cluster.Unsolved().String2D() << endl;
    cout << "solved:" << endl;
    cout << cluster.Solved().String2D() << endl;

    return 0;
}
