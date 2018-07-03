#pragma once

#include <unordered_map>
#include <vector>

#include "board.h"

class Solver {
public:
    Solver(const Board &board);
    int Solve();
private:
    bool Search(int depth, int maxDepth, int previousPiece);
    Board m_Board;
    std::vector<std::vector<Move>> m_MoveBuffers;
    std::unordered_map<BoardKey, int> m_Memo;
};
