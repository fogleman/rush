#include "solver.h"

Solution::Solution(const std::vector<Move> &moves) :
    m_Moves(moves)
{
}

int Solver::Deepen(Board &board, int maxDepth) {
    if (board.Solved()) {
        return 0;
    }
    m_Memo.clear();
    for (int i = 1; maxDepth < 0 || i <= maxDepth; i++) {
        m_Moves.resize(i);
        m_MoveBuffers.resize(i);
        if (Search(board, 0, i, -1)) {
            return i;
        }
    }
    return -1;
}

Solution Solver::Solve(Board &board) {
    // Search() fills m_Moves in place, so on success it already holds exactly
    // the solution; this also truncates it to empty for an already-solved board.
    m_Moves.resize(Deepen(board, -1));
    return Solution(m_Moves);
}

int Solver::CountMoves(Board &board) {
    return Deepen(board, -1);
}

bool Solver::SolvableWithin(Board &board, int maxDepth) {
    return Deepen(board, maxDepth) >= 0;
}

bool Solver::Search(Board &board, int depth, int maxDepth, int previousPiece) {
    int height = maxDepth - depth;
    if (height == 0) {
        return board.Solved();
    }

    const auto item = m_Memo.find(board.Key());
    if (item != m_Memo.end() && item->second >= height) {
        return false;
    }
    m_Memo[board.Key()] = height;

    // count occupied squares between primary piece and target
    const bb boardMask = board.Mask();
    const auto &primary = board.Pieces()[0];
    const int i0 = primary.Position() + primary.Size();
    const int i1 = Target + primary.Size() - 1;
    int minMoves = 0;
    for (int i = i0; i <= i1; i++) {
        const bb mask = (bb)1 << i;
        if ((mask & boardMask) != 0) {
            minMoves++;
        }
    }
    if (minMoves >= height) {
        return false;
    }

    auto &moves = m_MoveBuffers[depth];
    board.Moves(moves);
    for (const auto &move : moves) {
        if (move.Piece() == previousPiece) {
            continue;
        }
        board.DoMove(move);
        bool solved = Search(board, depth + 1, maxDepth, move.Piece());
        board.UndoMove(move);
        if (solved) {
            m_Moves[depth] = move;
            return true;
        }
    }

    return false;
}
