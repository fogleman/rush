#pragma once

#include <vector>

#include "board.h"

class Cluster {
public:
    Cluster(const Board &input);

    bool Canonical() const {
        return m_Canonical;
    }

    bool Solvable() const {
        return m_Solvable;
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

    const Board &Input() const {
        return m_Input;
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
    bool m_Canonical;
    bool m_Solvable;
    int m_NumStates;
    Board m_Input;
    Board m_Solved;
    Board m_Unsolved;
    std::vector<int> m_Distances;
};
