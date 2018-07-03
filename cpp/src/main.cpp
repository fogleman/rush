#include <cstdlib>
#include <iostream>
#include <unordered_set>

#include "board.h"
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

void handler(uint64_t counter, int group, const Board &board) {
    if (counter % 1000000 == 0) {
        print(board);
    }
}

int main() {
    Enumerator enumerator;
    enumerator.Enumerate(handler);
    return 0;

    // Board board("BCDDE.BCF.EGB.FAAGHHHI.G..JIKKLLJMM."); // 51 moves
    Board board("BB.C...D.CEE.DAAFGH.IIFGH.JKK.LLJ..."); // 541934 states

    vector<Move> moves;
    unordered_set<BoardKey> seen;

    for (int i = 0; i < 5000000; i++) {
        // if (board.Pieces()[0].Position() == Target) {
        //     cout << i << endl;
        //     break;
        // }
        seen.emplace(board.Key());

        board.Moves(moves);
        const int index = rand() % moves.size();
        board.DoMove(moves[index]);
    }

    print(board);
    cout << seen.size() << endl;

    return 0;
}
