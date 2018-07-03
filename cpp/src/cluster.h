#pragma once

#include <vector>

#include "board.h"

class Cluster {
public:
    Cluster(const uint64_t id, const uint64_t group, const Board &input);

    uint64_t ID() const {
        return m_ID;
    }

    uint64_t Group() const {
        return m_Group;
    }

    bool Canonical() const {
        return m_Canonical;
    }

    bool Solvable() const {
        return m_Solvable;
    }

    bool Minimal() const {
        return m_Minimal;
    }

    int NumStates() const {
        return m_NumStates;
    }

    int NumMoves() const {
        return m_Distances.size() - 1;
    }

    int NumSolvedStates() const {
        return m_Distances.front();
    }

    int NumUnsolvedStates() const {
        return m_Distances.back();
    }

    const Board &Solved() const {
        return m_Solved;
    }

    const Board &Unsolved() const {
        return m_Unsolved;
    }

    const std::vector<int> &DistanceCounts() const {
        return m_Distances;
    }

private:
    uint64_t m_ID;
    uint64_t m_Group;
    bool m_Canonical;
    bool m_Solvable;
    bool m_Minimal;
    int m_NumStates;
    Board m_Solved;
    Board m_Unsolved;
    std::vector<int> m_Distances;
};
