#pragma once

#include <boost/version.hpp>
#include <vector>

#include "board.h"

// The memo is cleared and refilled on every search, so a node-based map costs
// an allocation and a free per state visited. Flat storage avoids both.
#if BOOST_VERSION >= 108100
#include <boost/unordered/unordered_flat_map.hpp>
typedef boost::unordered_flat_map<BoardKey, int> SolverMemo;
#else
#include <boost/unordered_map.hpp>
typedef boost::unordered_map<BoardKey, int> SolverMemo;
#endif

class Solution {
public:
    explicit Solution(const std::vector<Move> &moves);

    const std::vector<Move> &Moves() const {
        return m_Moves;
    }

    int NumMoves() const {
        return m_Moves.size();
    }

private:
    std::vector<Move> m_Moves;
};

class Solver {
public:
    // Solve and CountMoves both loop forever on an unsolvable board.
    Solution Solve(Board &board);
    int CountMoves(Board &board);

    // True if the board can be solved in at most maxDepth moves. Cheaper than
    // CountMoves when only a comparison against a bound is needed, and safe on
    // unsolvable boards.
    bool SolvableWithin(Board &board, int maxDepth);

private:
    // Iterative deepening. maxDepth < 0 searches without a bound. Returns the
    // length of an optimal solution, or -1 if none exists within maxDepth.
    int Deepen(Board &board, int maxDepth);
    bool Search(Board &board, int depth, int maxDepth, int previousPiece);

    std::vector<Move> m_Moves;
    std::vector<std::vector<Move>> m_MoveBuffers;
    SolverMemo m_Memo;
};
