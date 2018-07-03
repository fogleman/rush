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
    if (desc.length() != BoardSize2) {
        throw "board string is wrong length";
    }

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

void Board::AddPiece(const Piece &piece) {
    m_Pieces.push_back(piece);
    m_Mask |= piece.Mask();
    if (piece.Stride() == H) {
        m_HorzMask |= piece.Mask();
    } else {
        m_VertMask |= piece.Mask();
    }
}

void Board::PopPiece() {
    const auto &piece = m_Pieces.back();
    m_Mask &= ~piece.Mask();
    if (piece.Stride() == H) {
        m_HorzMask &= ~piece.Mask();
    } else {
        m_VertMask &= ~piece.Mask();
    }
    m_Pieces.pop_back();
}

void Board::DoMove(const int index, const int steps) {
    auto &piece = m_Pieces[index];
    m_Mask &= ~piece.Mask();
    if (piece.Stride() == H) {
        m_HorzMask &= ~piece.Mask();
        piece.Move(steps);
        m_HorzMask |= piece.Mask();
    } else {
        m_VertMask &= ~piece.Mask();
        piece.Move(steps);
        m_VertMask |= piece.Mask();
    }
    m_Mask |= piece.Mask();
}

void Board::DoMove(const Move &move) {
    DoMove(move.Piece(), move.Steps());
}

void Board::UndoMove(const Move &move) {
    DoMove(move.Piece(), -move.Steps());
}

void Board::Moves(std::vector<Move> &moves) const {
    moves.clear();
    for (int i = 0; i < m_Pieces.size(); i++) {
        const auto &piece = m_Pieces[i];
        // compute range
        int forwardSteps, reverseSteps;
        if (piece.Stride() == H) {
            int x = piece.Position() % BoardSize;
            reverseSteps = -x;
            forwardSteps = BoardSize - piece.Size() - x;
        } else {
            int y = piece.Position() / BoardSize;
            reverseSteps = -y;
            forwardSteps = BoardSize - piece.Size() - y;
        }
        // reverse (negative steps)
        int p = piece.Position() - piece.Stride();
        bb mask = (bb)1 << p;
        for (int steps = -1; steps >= reverseSteps; steps--) {
            if ((m_Mask & mask) != 0) {
                break;
            }
            moves.emplace_back(Move(i, steps));
            mask >>= piece.Stride();
        }
        // forward (positive steps)
        p = piece.Position() + piece.Size() * piece.Stride();
        mask = (bb)1 << p;
        for (int steps = 1; steps <= forwardSteps; steps++) {
            if ((m_Mask & mask) != 0) {
                break;
            }
            moves.emplace_back(Move(i, steps));
            mask <<= piece.Stride();
        }
    }
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

bool operator<(const Board &b1, const Board &b2) {
    if (b1.HorzMask() == b2.HorzMask()) {
        return b1.VertMask() < b2.VertMask();
    }
    return b1.HorzMask() < b2.HorzMask();
}
