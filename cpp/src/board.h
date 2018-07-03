#pragma once

#include <iostream>
#include <string>
#include <vector>

#include "bb.h"
#include "config.h"
#include "move.h"
#include "piece.h"

class BoardKey {
public:
    explicit BoardKey(bb horz, bb vert) :
        m_HorzMask(horz),
        m_VertMask(vert)
    {}

    bb HorzMask() const {
        return m_HorzMask;
    }

    bb VertMask() const {
        return m_VertMask;
    }

    bool operator==(const BoardKey &other) const {
        return HorzMask() == other.HorzMask() && VertMask() == other.VertMask();
    }

private:
    bb m_HorzMask;
    bb m_VertMask;
};

namespace std {
    template<> struct hash<BoardKey> {
        size_t operator()(const BoardKey &b) const {
            return std::hash<bb>()(b.HorzMask()) ^ std::hash<bb>()(b.VertMask());
        }
    };
}

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

    BoardKey Key() const {
        return BoardKey(m_HorzMask, m_VertMask);
    }

    const std::vector<Piece> &Pieces() const {
        return m_Pieces;
    }

    void AddPiece(const Piece &piece);
    void PopPiece();

    void DoMove(const int piece, const int steps);
    void DoMove(const Move &move);
    void UndoMove(const Move &move);

    void Moves(std::vector<Move> &moves) const;

    std::string String() const;

private:
    bb m_Mask;
    bb m_HorzMask;
    bb m_VertMask;
    std::vector<Piece> m_Pieces;
};

std::ostream& operator<<(std::ostream &stream, const Board &board);

bool operator<(const Board &b1, const Board &b2);
