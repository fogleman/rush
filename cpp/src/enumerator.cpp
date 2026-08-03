#include "enumerator.h"

#include <algorithm>
#include <cmath>

#include "config.h"

PositionEntry::PositionEntry(
    const int group, const std::vector<Piece> &pieces, const bool clearRequire) :
    m_Group(group),
    m_Pieces(pieces),
    m_Mask(0),
    m_Require(0),
    m_Walls(0)
{
    bb movableMask = 0;
    for (const auto &piece : pieces) {
        m_Mask |= piece.Mask();
        if (piece.Fixed()) {
            m_Walls++;
        } else {
            movableMask |= piece.Mask();
        }
    }
    if (!pieces.empty() && !clearRequire) {
        const int stride = pieces[0].Stride();
        if (stride == H) {
            m_Require = (movableMask >> stride) & ~m_Mask & ~RightColumn;
        } else {
            m_Require = (movableMask >> stride) & ~m_Mask;
        }
    }
}

Enumerator::Enumerator() :
    m_NumRowCombos(1)
{
    std::vector<int> sizes;
    ComputeGroups(sizes, 0);
    ComputePositionEntries();
    for (int y = 0; y < BoardSize; y++) {
        m_NumRowCombos *= m_RowEntries[y].size();
    }
}

void Enumerator::Enumerate(const EnumeratorFunc &func) const {
    for (uint64_t combo = 0; combo < m_NumRowCombos; combo++) {
        EnumerateRowCombo(combo, func);
    }
}

void Enumerator::EnumerateRowCombo(
    uint64_t combo, const EnumeratorFunc &func) const
{
    const PositionEntry *rows[BoardSize];
    bb mask = 0;
    bb require = 0;
    int walls = 0;
    for (int y = 0; y < BoardSize; y++) {
        const auto &entries = m_RowEntries[y];
        const PositionEntry &pe = entries[combo % entries.size()];
        combo /= entries.size();
        rows[y] = &pe;
        mask |= pe.Mask();
        require |= pe.Require();
        walls += pe.Walls();
    }
    if (DoWalls && (walls > MaxWalls || walls < MinWalls)) {
        // there are no vertical walls, so the count is already final
        return;
    }

    // A necessary condition, checked in front of the whole column DFS: a
    // required cell can only be covered by a vertical piece, and a vertical
    // piece covering a cell has to extend to the cell above or below it, which
    // must in turn be free of row pieces. A few instructions against a DFS.
    const bb free = ~mask & BoardMask;
    const bb reachable = ((free << V) | (free >> V)) & BoardMask;
    if ((require & ~reachable) != 0) {
        return;
    }

    // The primary row goes on first, then the other rows top to bottom, then the
    // columns left to right. Piece labels in a reported board follow this order,
    // so it is part of the output format.
    Board board;
    for (const auto &piece : rows[PrimaryRow]->Pieces()) {
        board.AddPiece(piece);
    }
    for (int y = 0; y < BoardSize; y++) {
        if (y == PrimaryRow) {
            continue;
        }
        for (const auto &piece : rows[y]->Pieces()) {
            board.AddPiece(piece);
        }
    }
    PopulateColumn(func, board, 0, mask, require);
}

void Enumerator::PopulateColumn(
    const EnumeratorFunc &func, Board &board, int x,
    bb mask, bb require) const
{
    if (x >= BoardSize) {
        func(board);
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
            func, board, x + 1,
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
            // Goal enumeration: the primary piece sits on the target, so this
            // row admits exactly one placement of it (plus whatever walls fit
            // behind it). Every position emitted is then solvable by
            // construction, and canonicality is decided among a cluster's goal
            // states only -- see Cluster::Explore.
            if (ps[0].Position() != Target) {
                return;
            }
        }
        const int group = GroupForPieces(ps);
        // The primary's position is pinned, not minimized, so requiring the
        // cell behind it to be occupied would wrongly discard positions.
        m_RowEntries[y].emplace_back(
            PositionEntry(group, ps, y == PrimaryRow));
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
