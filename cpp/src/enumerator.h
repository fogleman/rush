#pragma once

#include <functional>
#include <vector>

#include "bb.h"
#include "board.h"
#include "piece.h"

class PositionEntry {
public:
    PositionEntry(const int group, const std::vector<Piece> &pieces);

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

private:
    int m_Group;
    std::vector<Piece> m_Pieces;
    bb m_Mask;
    bb m_Require;
};

typedef std::function<void(uint64_t id, uint64_t group, const Board &)> EnumeratorFunc;

class Enumerator {
public:
    Enumerator();

    void Enumerate(EnumeratorFunc func);
    void EnumerateGroup(const uint64_t group, EnumeratorFunc func);

    uint64_t NumGroups() const;

private:
    void PopulatePrimaryRow(
        EnumeratorFunc func, Board &board, uint64_t &id) const;
    void PopulateRow(
        EnumeratorFunc func, Board &board, uint64_t &id, int y,
        bb mask, bb require, uint64_t group) const;
    void PopulateCol(
        EnumeratorFunc func, Board &board, uint64_t &id, int x,
        bb mask, bb require, uint64_t group) const;

    void PopulateGroupPrimaryRow(
        const uint64_t group, EnumeratorFunc func, Board &board, uint64_t &id) const;
    void PopulateGroupRow(
        const uint64_t group, const int digit,
        EnumeratorFunc func, Board &board, uint64_t &id,
        int y, bb mask, bb require) const;
    void PopulateGroupCol(
        const uint64_t group, const int digit,
        EnumeratorFunc func, Board &board, uint64_t &id,
        int x, bb mask, bb require) const;

    void ComputeGroups(std::vector<int> &sizes, int sum);
    int GroupForPieces(const std::vector<Piece> &pieces);

    void ComputeRow(int y, int x, std::vector<Piece> &pieces);
    void ComputeCol(int x, int y, std::vector<Piece> &pieces);
    void ComputePositionEntries();

    std::vector<std::vector<int>> m_Groups;
    std::vector<std::vector<PositionEntry>> m_RowEntries;
    std::vector<std::vector<PositionEntry>> m_ColEntries;
};
