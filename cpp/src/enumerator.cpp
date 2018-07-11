#include "enumerator.h"

#include <algorithm>
#include <cmath>

#include "config.h"

PositionEntry::PositionEntry(const int group, const std::vector<Piece> &pieces) :
    m_Group(group),
    m_Pieces(pieces),
    m_Mask(0),
    m_Require(0)
{
    bb movableMask = 0;
    for (const auto &piece : pieces) {
        m_Mask |= piece.Mask();
        if (!piece.Fixed()) {
            movableMask |= piece.Mask();
        }
    }
    if (!pieces.empty()) {
        const int stride = pieces[0].Stride();
        if (stride == H) {
            m_Require = (movableMask >> stride) & ~m_Mask & ~RightColumn;
        } else {
            m_Require = (movableMask >> stride) & ~m_Mask;
        }
    }
}

Enumerator::Enumerator() {
    std::vector<int> sizes;
    ComputeGroups(sizes, 0);
    ComputePositionEntries();
}

void Enumerator::Enumerate(EnumeratorFunc func) {
    Board board;
    uint64_t id = 0;
    PopulatePrimaryRow(func, board, id);
}

void Enumerator::PopulatePrimaryRow(
    EnumeratorFunc func, Board &board, uint64_t &id) const
{
    for (const auto &pe : m_RowEntries[PrimaryRow]) {
        for (const auto &piece : pe.Pieces()) {
            board.AddPiece(piece);
        }
        PopulateRow(func, board, id, 0, pe.Mask(), pe.Require());
        for (int i = 0; i < pe.Pieces().size(); i++) {
            board.PopPiece();
        }
    }
}

void Enumerator::PopulateRow(
    EnumeratorFunc func, Board &board, uint64_t &id, int y,
    bb mask, bb require) const
{
    if (DoWalls) {
        int walls = 0;
        for (const auto &piece : board.Pieces()) {
            if (piece.Fixed()) {
                walls++;
            }
        }
        if (walls > MaxWalls) {
            return;
        }
        if (y >= BoardSize && walls < MinWalls) {
            return;
        }
    }
    if (y >= BoardSize) {
        PopulateColumn(func, board, id, 0, mask, require);
        return;
    }
    if (y == PrimaryRow) {
        PopulateRow(func, board, id, y + 1, mask, require);
        return;
    }
    for (const auto &pe : m_RowEntries[y]) {
        if ((mask & pe.Mask()) != 0) {
            continue;
        }
        for (const auto &piece : pe.Pieces()) {
            board.AddPiece(piece);
        }
        PopulateRow(
            func, board, id, y + 1,
            mask | pe.Mask(), require | pe.Require());
        for (int i = 0; i < pe.Pieces().size(); i++) {
            board.PopPiece();
        }
    }
}

void Enumerator::PopulateColumn(
    EnumeratorFunc func, Board &board, uint64_t &id, int x,
    bb mask, bb require) const
{
    if (x >= BoardSize) {
        func(id, board);
        id++;
        return;
    }
    for (const auto &pe : m_ColumnEntries[x]) {
        if ((mask & pe.Mask()) != 0) {
            continue;
        }
        if ((mask & pe.Require()) != pe.Require()) {
            continue;
        }
        const bb columnRequire = require & ColumnMasks[x];
        if ((pe.Mask() & columnRequire) != columnRequire) {
            continue;
        }
        for (const auto &piece : pe.Pieces()) {
            board.AddPiece(piece);
        }
        PopulateColumn(
            func, board, id, x + 1,
            mask | pe.Mask(), require | pe.Require());
        for (int i = 0; i < pe.Pieces().size(); i++) {
            board.PopPiece();
        }
    }
}

void Enumerator::ComputeGroups(std::vector<int> &sizes, int sum) {
    if (sum >= BoardSize) {
        return;
    }
    int walls = 0;
    for (const int size : sizes) {
        if (size == 1) {
            walls++;
        }
    }
    if (walls > MaxWalls) {
        return;
    }
    m_Groups.push_back(sizes);
    for (int s = MinPieceSize; s <= MaxPieceSize; s++) {
        sizes.push_back(s);
        ComputeGroups(sizes, sum + s);
        sizes.pop_back();
    }
}

int Enumerator::GroupForPieces(const std::vector<Piece> &pieces) {
    for (int i = 0; i < m_Groups.size(); i++) {
        const auto &group = m_Groups[i];
        if (group.size() != pieces.size()) {
            continue;
        }
        bool ok = true;
        for (int j = 0; j < group.size(); j++) {
            if (group[j] != pieces[j].Size()) {
                ok = false;
                break;
            }
        }
        if (ok) {
            return i;
        }
    }
    throw "GroupForPieces failed";
}

void Enumerator::ComputeRow(int y, int x, std::vector<Piece> &pieces) {
    if (x >= BoardSize) {
        int n = 0;
        int walls = 0;
        for (const auto &piece : pieces) {
            n += piece.Size();
            if (piece.Fixed()) {
                walls++;
            }
        }
        if (walls > MaxWalls) {
            return;
        }
        if (n >= BoardSize) {
            return;
        }
        std::vector<Piece> ps = pieces;
        // special constraints for the primary row
        if (y == PrimaryRow) {
            // can only have one non-wall (the primary piece itself)
            const int nonWalls = ps.size() - walls;
            if (nonWalls != 1) {
                return;
            }
            // find the non-wall
            int primaryIndex = -1;
            for (int i = 0; i < ps.size(); i++) {
                if (!ps[i].Fixed()) {
                    primaryIndex = i;
                    break;
                }
            }
            if (primaryIndex < 0) {
                return;
            }
            // swap it to position zero
            std::swap(ps[0], ps[primaryIndex]);
            // check its size
            if (ps[0].Size() != PrimarySize) {
                return;
            }
            // no walls can appear to the right of the primary piece
            for (int i = 1; i < ps.size(); i++) {
                if (ps[i].Position() > ps[0].Position()) {
                    return;
                }
            }
        }
        const int group = GroupForPieces(ps);
        m_RowEntries[y].emplace_back(PositionEntry(group, ps));
        return;
    }
    for (int s = MinPieceSize; s <= MaxPieceSize; s++) {
        if (x + s > BoardSize) {
            continue;
        }
        const int p = y * BoardSize + x;
        pieces.emplace_back(Piece(p, s, H));
        ComputeRow(y, x + s, pieces);
        pieces.pop_back();
    }
    ComputeRow(y, x + 1, pieces);
}

void Enumerator::ComputeColumn(int x, int y, std::vector<Piece> &pieces) {
    if (y >= BoardSize) {
        int n = 0;
        for (const auto &piece : pieces) {
            n += piece.Size();
        }
        if (n >= BoardSize) {
            return;
        }
        const int group = GroupForPieces(pieces);
        m_ColumnEntries[x].emplace_back(PositionEntry(group, pieces));
        return;
    }
    for (int s = MinPieceSize; s <= MaxPieceSize; s++) {
        if (s == 1) {
            // no "vertical" walls
            continue;
        }
        if (y + s > BoardSize) {
            continue;
        }
        const int p = y * BoardSize + x;
        pieces.emplace_back(Piece(p, s, V));
        ComputeColumn(x, y + s, pieces);
        pieces.pop_back();
    }
    ComputeColumn(x, y + 1, pieces);
}

void Enumerator::ComputePositionEntries() {
    m_RowEntries.resize(BoardSize);
    m_ColumnEntries.resize(BoardSize);
    std::vector<Piece> pieces;
    for (int i = 0; i < BoardSize; i++) {
        ComputeRow(i, 0, pieces);
        ComputeColumn(i, 0, pieces);
    }
    for (int i = 0; i < BoardSize; i++) {
        std::stable_sort(m_RowEntries[i].begin(), m_RowEntries[i].end(),
            [](const PositionEntry &a, const PositionEntry &b)
        {
            return a.Group() < b.Group();
        });
        std::stable_sort(m_ColumnEntries[i].begin(), m_ColumnEntries[i].end(),
            [](const PositionEntry &a, const PositionEntry &b)
        {
            return a.Group() < b.Group();
        });
    }
}
