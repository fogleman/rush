#include <cstdlib>
#include <iostream>
#include <unordered_set>

#include "board.h"
#include "config.h"

void print(const Board &board) {
    std::cout << board << std::endl;
    std::cout << BitboardString(board.Mask()) << std::endl;
    std::cout << BitboardString(board.HorzMask()) << std::endl;
    std::cout << BitboardString(board.VertMask()) << std::endl;
    std::cout << std::endl;
}

int main() {
    // Board board("BCDDE.BCF.EGB.FAAGHHHI.G..JIKKLLJMM."); // 51 moves
    Board board("BB.C...D.CEE.DAAFGH.IIFGH.JKK.LLJ..."); // 541934 states

    std::vector<Move> moves;
    std::unordered_set<Board, BoardMaskHash, BoardMaskEqual> seen;

    for (int i = 0; i < 5000000; i++) {
        // if (board.Pieces()[0].Position() == Target) {
        //     std::cout << i << std::endl;
        //     break;
        // }
        seen.insert(board);

        board.Moves(moves);
        const int index = std::rand() % moves.size();
        board.DoMove(moves[index]);
    }

    print(board);
    std::cout << seen.size() << std::endl;

    return 0;
}
