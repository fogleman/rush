#pragma once

#include "bb.h"
#include "config.h"

class Piece {
public:
    explicit Piece(int position, int size, int stride);

    int Position() const {
        return m_Position;
    }

    int Size() const {
        return m_Size;
    }

    int Stride() const {
        return m_Stride;
    }

    bb Mask() const {
        return m_Mask;
    }

    bool Fixed() const {
        return m_Size == 1;
    }

    void Move(int steps);

private:
    int m_Position;
    int m_Size;
    int m_Stride;
    bb m_Mask;
};
