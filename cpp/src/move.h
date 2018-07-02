#pragma once

class Move {
public:
    explicit Move(int piece, int steps);
    int Piece() const;
    int Steps() const;
private:
    int m_Piece;
    int m_Steps;
};
