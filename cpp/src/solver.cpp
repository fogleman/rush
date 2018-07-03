#include "solver.h"

Solver::Solver(const Board &board) :
    m_Board(board)
{
}

int Solver::Solve() {
    if (m_Board.Solved()) {
        return 0;
    }
    for (int i = 1; ; i++) {
        m_MoveBuffers.resize(i);
        if (Search(0, i, -1)) {
            return i;
        }
    }
}

bool Solver::Search(int depth, int maxDepth, int previousPiece) {
    int height = maxDepth - depth;
    if (height == 0) {
        return m_Board.Solved();
    }

    const auto item = m_Memo.find(m_Board.Key());
    if (item != m_Memo.end() && item->second >= height) {
        return false;
    }
    m_Memo[m_Board.Key()] = height;

    // count occupied squares between primary piece and target
    const auto &primary = m_Board.Pieces()[0];
    const int i0 = primary.Position() + primary.Size();
    const int i1 = Target + primary.Size() - 1;
    int minMoves = 0;
    for (int i = i0; i <= i1; i++) {
        const bb mask = (bb)1 << i;
        if ((mask & m_Board.Mask()) != 0) {
            minMoves++;
        }
    }
    if (minMoves >= height) {
        return false;
    }

    auto &moves = m_MoveBuffers[depth];
    m_Board.Moves(moves);
    for (const auto &move : moves) {
        if (move.Piece() == previousPiece) {
            continue;
        }
        m_Board.DoMove(move);
        bool solved = Search(depth + 1, maxDepth, move.Piece());
        m_Board.UndoMove(move);
        if (solved) {
            m_Memo[m_Board.Key()] = height - 1;
            return true;
        }
    }

    return false;
}
