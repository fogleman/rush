#pragma once

#include <unordered_map>
#include <vector>

#include "board.h"

struct BoardKeyHash {
    size_t operator()(const BoardKey &key) const {
        const uint64_t h1 = std::get<0>(key);
        const uint64_t h2 = std::get<1>(key);
        uint64_t h = h1 ^ (h2 * 0x9e3779b97f4a7c15ULL + (h1 << 6) + (h1 >> 2));
        h ^= h >> 33;
        h *= 0xff51afd7ed558ccdULL;
        h ^= h >> 33;
        return h;
    }
};

class Solution {
public:
    Solution();
    explicit Solution(const std::vector<Move> &moves);

    bool Solvable() const {
        return m_Solvable;
    }

    const std::vector<Move> &Moves() const {
        return m_Moves;
    }

    int NumMoves() const {
        return m_Moves.size();
    }

private:
    bool m_Solvable;
    std::vector<Move> m_Moves;
};

class Solver {
public:
    Solution Solve(Board &board);
    int CountMoves(Board &board);
private:
    bool Search(Board &board, int depth, int maxDepth, int previousPiece);
    std::vector<Move> m_Moves;
    std::vector<std::vector<Move>> m_MoveBuffers;
    std::unordered_map<BoardKey, int, BoardKeyHash> m_Memo;
};
