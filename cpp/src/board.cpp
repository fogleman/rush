#include "board.h"

#include <algorithm>
#include <map>

Board::Board() :
    m_Mask(0),
    m_HorzMask(0),
    m_VertMask(0)
{
}

Board::Board(std::string desc) :
    m_Mask(0),
    m_HorzMask(0),
    m_VertMask(0)
{
    std::map<char, std::vector<int>> positions;
    for (int i = 0; i < desc.length(); i++) {
        const char label = desc[i];
        if (label == '.') {
            continue;
        }
        positions[label].push_back(i);
    }

    std::vector<char> labels;
    labels.reserve(positions.size());
    for (const auto &pair : positions) {
        labels.push_back(pair.first);
    }
    std::sort(labels.begin(), labels.end());

    m_Pieces.reserve(labels.size());
    for (const char label : labels) {
        const auto &ps = positions[label];
        if (ps.size() < MinPieceSize) {
            throw "piece size < MinPieceSize";
        }
        if (ps.size() > MaxPieceSize) {
            throw "piece size > MaxPieceSize";
        }
        const int stride = ps[1] - ps[0];
        if (stride != H && stride != V) {
            throw "invalid piece shape";
        }
        for (int i = 2; i < ps.size(); i++) {
            if (ps[i] - ps[i-1] != stride) {
                throw "invalid piece shape";
            }
        }
        AddPiece(Piece(ps[0], ps.size(), stride));
    }
}

const std::vector<Piece> &Board::Pieces() const {
    return m_Pieces;
}

bb Board::Mask() const {
    return m_Mask;
}

bb Board::HorzMask() const {
    return m_HorzMask;
}

bb Board::VertMask() const {
    return m_VertMask;
}

void Board::AddPiece(const Piece &piece) {
    m_Pieces.push_back(piece);
    m_Mask |= piece.Mask();
    if (piece.Stride() == H) {
        m_HorzMask |= piece.Mask();
    } else {
        m_VertMask |= piece.Mask();
    }
}

void Board::DoMove(const int piece, const int steps) {
    auto &p = m_Pieces[piece];
    m_Mask &= ~p.Mask();
    if (p.Stride() == H) {
        m_HorzMask &= ~p.Mask();
        p.Move(steps);
        m_HorzMask |= p.Mask();
    } else {
        m_VertMask &= ~p.Mask();
        p.Move(steps);
        m_VertMask |= p.Mask();
    }
    m_Mask |= p.Mask();
}

void Board::DoMove(const Move &move) {
    DoMove(move.Piece(), move.Steps());
}

void Board::UndoMove(const Move &move) {
    DoMove(move.Piece(), -move.Steps());
}

std::vector<Move> Board::Moves() const {
    std::vector<Move> moves;
    for (int i = 0; i < m_Pieces.size(); i++) {
        const auto &piece = m_Pieces[i];
        const int position = piece.Position();
        const int size = piece.Size();
        const int stride = piece.Stride();
        // compute range
        int forwardSteps, reverseSteps;
        if (stride == H) {
            int x = position % BoardSize;
            reverseSteps = -x;
            forwardSteps = BoardSize - size - x;
        } else {
            int y = position / BoardSize;
            reverseSteps = -y;
            forwardSteps = BoardSize - size - y;
        }
        // reverse (negative steps)
        int p = position - stride;
        bb mask = (bb)1 << p;
        for (int steps = -1; steps >= reverseSteps; steps--) {
            if ((m_Mask & mask) != 0) {
                break;
            }
            moves.emplace_back(Move(i, steps));
            mask >>= stride;
        }
        // forward (positive steps)
        p = position + size * stride;
        mask = (bb)1 << p;
        for (int steps = 1; steps <= forwardSteps; steps++) {
            if ((m_Mask & mask) != 0) {
                break;
            }
            moves.emplace_back(Move(i, steps));
            mask <<= stride;
        }
    }
    return moves;
}

std::string Board::String() const {
    std::string s(BoardSize2, '.');
    for (int i = 0; i < m_Pieces.size(); i++) {
        const Piece &piece = m_Pieces[i];
        const char c = 'A' + i;
        int p = piece.Position();
        for (int i = 0; i < piece.Size(); i++) {
            s[p] = c;
            p += piece.Stride();
        }
    }
    return s;
}

std::ostream& operator<<(std::ostream &stream, const Board &board) {
    return stream << board.String();
}
