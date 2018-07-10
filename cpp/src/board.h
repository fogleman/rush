#pragma once

#include <boost/container/small_vector.hpp>
#include <iostream>
#include <string>
#include <vector>

#include "bb.h"
#include "config.h"
#include "move.h"
#include "piece.h"

class Board {
public:
    Board();
    explicit Board(std::string desc);

    bb Mask() const {
        return m_Mask;
    }

    bb HorzMask() const {
        return m_HorzMask;
    }

    bb VertMask() const {
        return m_VertMask;
    }

    bb Key() const {
        return m_Key;
    }

    const boost::container::small_vector<Piece, BoardSize2> &Pieces() const {
        return m_Pieces;
    }

    bool Solved() const {
        return m_Pieces[0].Position() == Target;
    }

    void AddPiece(const Piece &piece);
    void PopPiece();
    void RemovePiece(const int i);

    void DoMove(const int piece, const int steps);
    void DoMove(const Move &move);
    void UndoMove(const Move &move);

    void Moves(std::vector<Move> &moves) const;

    std::string String() const;
    std::string String2D() const;

private:
    bb m_Mask;
    bb m_HorzMask;
    bb m_VertMask;
    bb m_Key;
    boost::container::small_vector<Piece, BoardSize2> m_Pieces;
};

std::ostream& operator<<(std::ostream &stream, const Board &board);

bool operator<(const Board &b1, const Board &b2);
