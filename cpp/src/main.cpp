#include <cstdlib>
#include <iostream>

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
    Board board("BCDDE.BCF.EGB.FAAGHHHI.G..JIKKLLJMM.");

    std::vector<Move> moves;
    for (int i = 0; i < 5000000; i++) {
        // if (board.Pieces()[0].Position() == Target) {
        //     std::cout << i << std::endl;
        //     break;
        // }

        board.Moves(moves);
        const int index = std::rand() % moves.size();
        board.DoMove(moves[index]);
    }

    return 0;
}
