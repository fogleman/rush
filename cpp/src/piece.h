#pragma once

#include "bb.h"
#include "config.h"

class Piece {
public:
    explicit Piece(int position, int size, int stride);
    int Position() const;
    int Size() const;
    int Stride() const;
    bb Mask() const;
    void Move(int steps);
private:
    int m_Position;
    int m_Size;
    int m_Stride;
    bb m_Mask;
};
