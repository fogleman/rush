#include <cstdlib>
#include <iostream>

#include "board.h"
#include "config.h"

int main() {
    Board board("BCDDE.BCF.EGB.FAAGHHHI.G..JIKKLLJMM.");

    for (int i = 0; i < 5000000; i++) {
        // std::cout << board << std::endl;
        // std::cout << BitboardString(board.Mask()) << std::endl;
        // std::cout << BitboardString(board.HorzMask()) << std::endl;
        // std::cout << BitboardString(board.VertMask()) << std::endl;
        // std::cout << std::endl;

        // if (board.Pieces()[0].Position() == Target) {
        //     std::cout << i << std::endl;
        //     break;
        // }

        const auto moves = board.Moves();
        const int index = std::rand() % moves.size();
        board.DoMove(moves[index]);
    }

    return 0;
}
