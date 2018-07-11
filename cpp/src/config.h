#pragma once

#include "bb.h"

const int BoardSize = 5;
const int PrimaryRow = 2;
const int PrimarySize = 2;
const int MinPieceSize = 2;
const int MaxPieceSize = 3;
const int MinWalls = 0;
const int MaxWalls = 1;
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
// const uint64_t MaxID = 268108; // 5x5
// const uint64_t MaxID = 243502785; // 6x6
// const uint64_t MaxID = 2689860; // 5x5, 1 wall
const uint64_t MaxID = 3331773541; // 6x6, 1 wall

const int BoardSize2 = BoardSize * BoardSize;
const int Target = PrimaryRow * BoardSize + BoardSize - PrimarySize;
const int H = 1; // horizontal stride
const int V = BoardSize; // vertical stride
const bool DoWalls = MinPieceSize == 1;

const std::vector<std::vector<bb>> ZobristKeys = []() {
    const int n = BoardSize2;
    std::vector<std::vector<bb>> keys(n, std::vector<bb>(n, 0));
    std::random_device rd;
    std::mt19937 gen(rd());
    for (int i = 0; i < n; i++) {
        for (int j = 0; j < n; j++) {
            keys[i][j] = RandomBitboard(gen);
        }
    }
    return keys;
}();

const bb TopRow = []() {
    bb result = 0;
    for (int x = 0; x < BoardSize; x++) {
        result |= (bb)1 << x;
    }
    return result;
}();

const bb BottomRow = []() {
    bb result = 0;
    for (int x = 0; x < BoardSize; x++) {
        result |= (bb)1 << (BoardSize2 - x - 1);
    }
    return result;
}();

const bb LeftColumn = []() {
    bb result = 0;
    for (int y = 0; y < BoardSize; y++) {
        result |= (bb)1 << (y * BoardSize);
    }
    return result;
}();

const bb RightColumn = []() {
    bb result = 0;
    for (int y = 0; y < BoardSize; y++) {
        result |= (bb)1 << (y * BoardSize + BoardSize - 1);
    }
    return result;
}();
