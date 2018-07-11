#pragma once

#include <boost/container/small_vector.hpp>
#include <iostream>
#include <string>
#include <tuple>
#include <vector>

#include "bb.h"
#include "config.h"
#include "move.h"
#include "piece.h"

typedef std::tuple<bb, bb> BoardKey;

class Board {
public:
    Board();
    explicit Board(std::string desc);

    bb Mask() const {
        return m_HorzMask | m_VertMask;
    }

    bb HorzMask() const {
        return m_HorzMask;
    }

    bb VertMask() const {
        return m_VertMask;
    }

    BoardKey Key() const {
        return std::make_tuple(m_HorzMask, m_VertMask);
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
    bb m_HorzMask;
    bb m_VertMask;
    boost::container::small_vector<Piece, BoardSize2> m_Pieces;
};

std::ostream& operator<<(std::ostream &stream, const Board &board);

bool operator<(const Board &b1, const Board &b2);
