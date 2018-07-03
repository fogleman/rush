#pragma once

#include "bb.h"

const int BoardSize = 6;
const int PrimaryRow = 2;
const int PrimarySize = 2;
const int MinPieceSize = 2;
const int MaxPieceSize = 3;

const int BoardSize2 = BoardSize * BoardSize;
const int Target = PrimaryRow * BoardSize + BoardSize - PrimarySize;
const int H = 1; // horizontal stride
const int V = BoardSize; // vertical stride

const bb RightColumn = []() {
    bb result = 0;
    for (int y = 0; y <= BoardSize; y++) { // TODO: = needed?
        result |= (bb)1 << (y * BoardSize + BoardSize - 1);
    }
    return result;
}();
