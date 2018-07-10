#pragma once

class Move {
public:
    Move() = default;
    explicit Move(int piece, int steps);

    int Piece() const {
        return m_Piece;
    }

    int Steps() const {
        return m_Steps;
    }

private:
    int m_Piece;
    int m_Steps;
};
