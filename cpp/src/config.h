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

// 5x5
// 0 walls = 268108
// 1 walls = 2988669
// 2 walls = 13341759
// 3 walls = 41965437
// 4 walls = 95002637
// 0-3 wls = 58295867
// 0-4 wls = 153298505

// const uint64_t MaxID = 1149; // 4x4
const uint64_t MaxID = 268108; // 5x5
// const uint64_t MaxID = 243502785; // 6x6
// const uint64_t MaxID = 2689860; // 5x5, 1 wall
// const uint64_t MaxID = 3331773541; // 6x6, 1 wall

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
