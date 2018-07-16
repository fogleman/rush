#include "board.h"

#include <algorithm>
#include <map>

Board::Board() :
    m_HorzMask(0),
    m_VertMask(0)
{
}

Board::Board(std::string desc) :
    m_HorzMask(0),
    m_VertMask(0)
{
    if (desc.length() != BoardSize2) {
        throw "board string is wrong length";
    }

    std::map<char, std::vector<int>> positions;
    for (int i = 0; i < desc.length(); i++) {
        const char label = desc[i];
        if (label == '.' || label == 'o') {
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
    if (piece.Stride() == H) {
        m_HorzMask |= piece.Mask();
    } else {
        m_VertMask |= piece.Mask();
    }
}

void Board::PopPiece() {
    const auto &piece = m_Pieces.back();
    if (piece.Stride() == H) {
        m_HorzMask &= ~piece.Mask();
    } else {
        m_VertMask &= ~piece.Mask();
    }
    m_Pieces.pop_back();
}

void Board::RemovePiece(const int i) {
    const auto &piece = m_Pieces[i];
    if (piece.Stride() == H) {
        m_HorzMask &= ~piece.Mask();
    } else {
        m_VertMask &= ~piece.Mask();
    }
    m_Pieces.erase(m_Pieces.begin() + i);
}

void Board::DoMove(const int i, const int steps) {
    auto &piece = m_Pieces[i];
    if (piece.Stride() == H) {
        m_HorzMask &= ~piece.Mask();
        piece.Move(steps);
        m_HorzMask |= piece.Mask();
    } else {
        m_VertMask &= ~piece.Mask();
        piece.Move(steps);
        m_VertMask |= piece.Mask();
    }
}

void Board::DoMove(const Move &move) {
    DoMove(move.Piece(), move.Steps());
}

void Board::UndoMove(const Move &move) {
    DoMove(move.Piece(), -move.Steps());
}

void Board::Moves(std::vector<Move> &moves) const {
    moves.clear();
    const bb boardMask = Mask();
    for (int i = 0; i < m_Pieces.size(); i++) {
        const auto &piece = m_Pieces[i];
        if (piece.Fixed()) {
            continue;
        }
        if (piece.Stride() == H) {
            // reverse / left (negative steps)
            if ((piece.Mask() & LeftColumn) == 0) {
                bb mask = (piece.Mask() >> H) & ~piece.Mask();
                int steps = -1;
                while ((boardMask & mask) == 0) {
                    moves.emplace_back(Move(i, steps));
                    if ((mask & LeftColumn) != 0) {
                        break;
                    }
                    mask >>= H;
                    steps--;
                }
            }
            // forward / right (positive steps)
            if ((piece.Mask() & RightColumn) == 0) {
                bb mask = (piece.Mask() << H) & ~piece.Mask();
                int steps = 1;
                while ((boardMask & mask) == 0) {
                    moves.emplace_back(Move(i, steps));
                    if ((mask & RightColumn) != 0) {
                        break;
                    }
                    mask <<= H;
                    steps++;
                }
            }
        } else {
            // reverse / up (negative steps)
            if ((piece.Mask() & TopRow) == 0) {
                bb mask = (piece.Mask() >> V) & ~piece.Mask();
                int steps = -1;
                while ((boardMask & mask) == 0) {
                    moves.emplace_back(Move(i, steps));
                    if ((mask & TopRow) != 0) {
                        break;
                    }
                    mask >>= V;
                    steps--;
                }
            }
            // forward / down (positive steps)
            if ((piece.Mask() & BottomRow) == 0) {
                bb mask = (piece.Mask() << V) & ~piece.Mask();
                int steps = 1;
                while ((boardMask & mask) == 0) {
                    moves.emplace_back(Move(i, steps));
                    if ((mask & BottomRow) != 0) {
                        break;
                    }
                    mask <<= V;
                    steps++;
                }
            }
        }
    }
}

std::string Board::String() const {
    std::string s(BoardSize2, '.');
    for (int i = 0; i < m_Pieces.size(); i++) {
        const Piece &piece = m_Pieces[i];
        const char c = piece.Fixed() ? 'x' : 'A' + i;
        int p = piece.Position();
        for (int i = 0; i < piece.Size(); i++) {
            s[p] = c;
            p += piece.Stride();
        }
    }
    return s;
}

std::string Board::String2D() const {
    std::string s(BoardSize * (BoardSize + 1), '.');
    for (int y = 0; y < BoardSize; y++) {
        const int p = y * (BoardSize + 1) + BoardSize;
        s[p] = '\n';
    }
    for (int i = 0; i < m_Pieces.size(); i++) {
        const Piece &piece = m_Pieces[i];
        const char c = piece.Fixed() ? 'x' : 'A' + i;
        int stride = piece.Stride();
        if (stride == V) {
            stride++;
        }
        const int y = piece.Position() / BoardSize;
        const int x = piece.Position() % BoardSize;
        int p = y * (BoardSize + 1) + x;
        for (int i = 0; i < piece.Size(); i++) {
            s[p] = c;
            p += stride;
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
