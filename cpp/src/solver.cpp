#include "solver.h"

Solution::Solution(const std::vector<Move> &moves) :
    m_Moves(moves)
{
}

Solution Solver::Solve(Board &board) {
    m_Moves.resize(0);
    if (board.Solved()) {
        return Solution(m_Moves);
    }
    m_Memo.clear();
    for (int i = 1; ; i++) {
        m_Moves.resize(i);
        m_MoveBuffers.resize(i);
        if (Search(board, 0, i, -1)) {
            return Solution(m_Moves);
        }
    }
}

int Solver::CountMoves(Board &board) {
    if (board.Solved()) {
        return 0;
    }
    m_Memo.clear();
    for (int i = 1; ; i++) {
        m_Moves.resize(i);
        m_MoveBuffers.resize(i);
        if (Search(board, 0, i, -1)) {
            return i;
        }
    }
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
            m_Memo[board.Key()] = height - 1;
            m_Moves[depth] = move;
            return true;
        }
    }

    return false;
}
