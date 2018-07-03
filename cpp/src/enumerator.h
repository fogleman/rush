#pragma once

#include <vector>

#include "bb.h"
#include "piece.h"

class PositionEntry {
public:
    PositionEntry(const int group, const std::vector<Piece> &pieces);

    int Group() const {
        return m_Group;
    }

private:
    int m_Group;
    std::vector<Piece> m_Pieces;
    bb m_Mask;
    bb m_Require;
};

class Enumerator {
public:
    Enumerator();

private:
    void ComputeGroups(std::vector<int> &sizes, int sum);
    int GroupForPieces(const std::vector<Piece> &pieces);

    void ComputeRow(int y, int x, std::vector<Piece> &pieces);
    void ComputeCol(int x, int y, std::vector<Piece> &pieces);
    void ComputePositionEntries();

    std::vector<std::vector<int>> m_Groups;
    std::vector<std::vector<PositionEntry>> m_RowEntries;
    std::vector<std::vector<PositionEntry>> m_ColEntries;
};
