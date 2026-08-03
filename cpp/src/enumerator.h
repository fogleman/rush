#pragma once

#include <functional>
#include <vector>

#include "bb.h"
#include "board.h"
#include "piece.h"

class PositionEntry {
public:
    // clearRequire drops the "the cell behind each piece must be occupied"
    // constraint, which is what the primary row needs: its piece is pinned on
    // the target rather than pushed as far back as it will go.
    PositionEntry(
        const int group, const std::vector<Piece> &pieces,
        const bool clearRequire = false);

    int Group() const {
        return m_Group;
    }

    const std::vector<Piece> &Pieces() const {
        return m_Pieces;
    }

    bb Mask() const {
        return m_Mask;
    }

    bb Require() const {
        return m_Require;
    }

    int Walls() const {
        return m_Walls;
    }

private:
    int m_Group;
    std::vector<Piece> m_Pieces;
    bb m_Mask;
    bb m_Require;
    int m_Walls;
};

typedef std::function<void(const Board &)> EnumeratorFunc;

class Enumerator {
public:
    Enumerator();

    // Enumerates every position, in row-combination order. Instances are
    // immutable once constructed, so one can be shared by every worker.
    void Enumerate(const EnumeratorFunc &func) const;

    // Enumerates the positions of a single row combination.
    //
    // Rows never conflict with one another -- a row entry only occupies its own
    // row -- so choosing row entries needs no search at all: it is a mixed-radix
    // odometer over the per-row entry lists, and only the columns need a DFS.
    // That makes a combination index a self-contained unit of work: it can be
    // handed to a worker, assigned to a shard, or written down and resumed,
    // without enumerating anything to find where it starts.
    //
    // The callback is passed by reference all the way down: taking it by value
    // copies the std::function at every level of the recursion, and heap
    // allocates on each copy if the captures are too large to store inline.
    void EnumerateRowCombo(uint64_t combo, const EnumeratorFunc &func) const;

    // Size of the row-combination space: the product of the per-row entry
    // counts. The primary row contributes a radix of one, since goal
    // enumeration pins the primary piece on the target.
    uint64_t NumRowCombos() const {
        return m_NumRowCombos;
    }

    int NumRowEntries(const int y) const {
        return m_RowEntries[y].size();
    }

    int NumColumnEntries(const int x) const {
        return m_ColumnEntries[x].size();
    }

private:
    void PopulateColumn(
        const EnumeratorFunc &func, Board &board, int x,
        bb mask, bb require) const;

    void ComputeGroups(std::vector<int> &sizes, int sum);
    int GroupForPieces(const std::vector<Piece> &pieces);

    void ComputeRow(int y, int x, std::vector<Piece> &pieces);
    void ComputeColumn(int x, int y, std::vector<Piece> &pieces);
    void ComputePositionEntries();

    std::vector<std::vector<int>> m_Groups;
    std::vector<std::vector<PositionEntry>> m_RowEntries;
    std::vector<std::vector<PositionEntry>> m_ColumnEntries;
    uint64_t m_NumRowCombos;
};
