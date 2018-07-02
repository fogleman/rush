#include "move.h"

Move::Move(int piece, int steps) :
    m_Piece(piece),
    m_Steps(steps)
{}

int Move::Piece() const {
    return m_Piece;
}

int Move::Steps() const {
    return m_Steps;
}
