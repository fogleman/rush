#include <cstdlib>
#include <iostream>
#include <unordered_set>

#include "board.h"
#include "config.h"
#include "enumerator.h"
#include "search.h"

using namespace std;

void print(const Board &board) {
    cout << board << endl;
    // cout << BitboardString(board.Mask()) << endl;
    // cout << BitboardString(board.HorzMask()) << endl;
    // cout << BitboardString(board.VertMask()) << endl;
    // cout << endl;
}

int main() {
    // Enumerator enumerator;
    // enumerator.Enumerate([&](uint64_t counter, int group, const Board &board) {
    //     ReachableStates(board);
    // });
    // return 0;

    // 51 83 13 BCDDE.BCF.EGB.FAAGHHHI.G..JIKKLLJMM. 4780
    // Board board("BCDDE.BCF.EGB.FAAGHHHI.G..JIKKLLJMM.");

    // 15 32 12 BB.C...D.CEE.DAAFGH.IIFGH.JKK.LLJ... 541934
    Board board("BB.C...D.CEE.DAAFGH.IIFGH.JKK.LLJ...");

    // 24 43 13 B..CDDBEEC.F.G.AAF.GHHIJKKL.IJ..L.MM 278666
    // Board board("B..CDDBEEC.F.G.AAF.GHHIJKKL.IJ..L.MM");

    const int count = ReachableStates(board);
    cout << count << endl;

    return 0;
}
