#pragma once

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

    size_t operator()(const Board &board) const;
    bool operator==(const Board& other) const;
private:
    bb m_Mask;
    bb m_HorzMask;
    bb m_VertMask;
    std::vector<Piece> m_Pieces;
};

std::ostream& operator<<(std::ostream &stream, const Board &board);

bool operator<(const Board &b1, const Board &b2);

struct BoardMaskHash {
public:
    size_t operator()(const Board &board) const {
        return std::hash<bb>()(board.HorzMask()) ^ std::hash<bb>()(board.VertMask());
    }
};

struct BoardMaskEqual {
public:
    bool operator()(const Board &b1, const Board &b2) const {
        return b1.HorzMask() == b2.HorzMask() && b1.VertMask() == b2.VertMask();
    }
};
