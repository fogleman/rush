#pragma once

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
        return m_HorzMask | m_VertMask | m_WallMask;
    }

    bb WallMask() const {
        return m_WallMask;
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

    const std::vector<Piece> &Pieces() const {
        return m_Pieces;
    }

    int Target() const {
        return m_Target;
    }

    bool Solved() const {
        return m_Pieces[0].Position() == m_Target;
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
    bb m_WallMask;
    int m_Target;
    std::vector<Piece> m_Pieces;
};

std::ostream& operator<<(std::ostream &stream, const Board &board);

bool operator<(const Board &b1, const Board &b2);
