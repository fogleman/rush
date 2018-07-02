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
    Board(std::string desc);

    const std::vector<Piece> &Pieces() const;

    bb Mask() const;
    bb HorzMask() const;
    bb VertMask() const;

    void AddPiece(const Piece &piece);

    void DoMove(const int piece, const int steps);
    void DoMove(const Move &move);
    void UndoMove(const Move &move);

    std::vector<Move> Moves() const;

    std::string String() const;
private:
    bb m_Mask;
    bb m_HorzMask;
    bb m_VertMask;
    std::vector<Piece> m_Pieces;
};

std::ostream& operator<<(std::ostream &stream, const Board &board);
