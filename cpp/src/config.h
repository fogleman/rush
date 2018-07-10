#pragma once

#include "bb.h"

const int BoardSize = 6;
const int PrimaryRow = 2;
const int PrimarySize = 2;
const int MinPieceSize = 1;
const int MaxPieceSize = 3;
const int MaxWalls = 1;
const int NumWorkers = 4;

// const uint64_t MaxID = 1149; // 4x4
// const uint64_t MaxID = 268108; // 5x5
// const uint64_t MaxID = 243502785; // 6x6
// const uint64_t MaxID = 2689860; // 5x5, 1 wall
const uint64_t MaxID = 3331773541; // 6x6, 1 wall

const int BoardSize2 = BoardSize * BoardSize;
const int Target = PrimaryRow * BoardSize + BoardSize - PrimarySize;
const int H = 1; // horizontal stride
const int V = BoardSize; // vertical stride

const bb RightColumn = []() {
    bb result = 0;
    for (int y = 0; y < BoardSize; y++) {
        result |= (bb)1 << (y * BoardSize + BoardSize - 1);
    }
    return result;
}();

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
