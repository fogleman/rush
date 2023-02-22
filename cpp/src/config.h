#pragma once

#include <array>

#include "bb.h"

const int BoardSize = 5;
const int PrimaryRow = 2;
const int PrimarySize = 2;
const int MinPieceSize = 2;
const int MaxPieceSize = 3;
const int MinWalls = 0;
const int MaxWalls = 0;
const int NumWorkers = 4;

// const uint64_t MaxID = 1348; // 4x4
// const uint64_t MaxID = 9803; // 4x4, 0-1 walls
// const uint64_t MaxID = 33952; // 4x4, 0-2 walls
// const uint64_t MaxID = 76837; // 4x4, 0-3 walls

const uint64_t MaxID = 268108; // 5x5
// const uint64_t MaxID = 2988669; // 5x5, 0-1 walls
// const uint64_t MaxID = 16330429; // 5x5, 0-2 walls

// const uint64_t MaxID = 243502785; // 6x6
// const uint64_t MaxID = 3670622351; // 6x6, 0-1 walls
// const uint64_t MaxID = 27403231254; // 6x6, 0-2 walls

// const uint64_t MaxID = 561276504436; // 7x7 - 5h42m

const int BoardSize2 = BoardSize * BoardSize;
const int Target = PrimaryRow * BoardSize + BoardSize - PrimarySize;
const int H = 1; // horizontal stride
const int V = BoardSize; // vertical stride
const bool DoWalls = MinPieceSize == 1;

const std::array<bb, BoardSize> RowMasks = []() {
    std::array<bb, BoardSize> rowMasks;
    for (int y = 0; y < BoardSize; y++) {
        bb mask = 0;
        for (int x = 0; x < BoardSize; x++) {
            const int i = y * BoardSize + x;
            mask |= (bb)1 << i;
        }
        rowMasks[y] = mask;
    }
    return rowMasks;
}();

const std::array<bb, BoardSize> ColumnMasks = []() {
    std::array<bb, BoardSize> columnMasks;
    for (int x = 0; x < BoardSize; x++) {
        bb mask = 0;
        for (int y = 0; y < BoardSize; y++) {
            const int i = y * BoardSize + x;
            mask |= (bb)1 << i;
        }
        columnMasks[x] = mask;
    }
    return columnMasks;
}();

const bb TopRow = RowMasks.front();
const bb BottomRow = RowMasks.back();
const bb LeftColumn = ColumnMasks.front();
const bb RightColumn = ColumnMasks.back();
