#pragma once

#include <cstdint>
#include <type_traits>
#include <vector>

#include "board.h"
#include "solver.h"

// Cluster explores one cluster (one connected component of the state graph)
// and reports what the puzzle database needs to know about it.
//
// Reuse a single instance per thread and call Explore() repeatedly: the hash
// table, the node arena and the solver are all retained across calls, so once
// they are warm a cluster costs no heap allocation at all. Instances are not
// thread safe.
class Cluster {
public:
    void Explore(const Board &input);

    // True if the board passed to Explore() was the lexicographically smallest
    // state in its cluster. When false, exploration stopped early and every
    // other accessor is meaningless: some other state will report this cluster.
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

    // Moves needed to solve Unsolved(), i.e. the cluster's eccentricity with
    // respect to its goal states. Zero unless Solvable().
    int NumMoves() const {
        return m_MaxDistance;
    }

    // The hardest state in the cluster: the one furthest from a goal state,
    // breaking ties in favor of the lexicographically smallest. Empty unless
    // Solvable().
    const Board &Unsolved() const {
        return m_Unsolved;
    }

    // Number of states at each distance from a goal state. Only populated for
    // minimal clusters, since that is the only case the database reports.
    const std::vector<int> &DistanceCounts() const {
        return m_Distances;
    }

private:
    // A state is the tuple of every piece's offset along its own row or column.
    // Sizes, strides and lines are fixed for the whole cluster, so only the
    // offsets vary, and the whole tuple packs into a single integer that serves
    // as the hash key.
    //
    // This is equivalent to keying on (HorzMask, VertMask): pieces cannot pass
    // each other, so their order along a line is fixed and a line's bitmask
    // determines its pieces' offsets uniquely.
    static const int OffsetBits = 3;

    // Every piece covers at least two cells, except walls, of which a board
    // holds at most MaxWalls.
    static const int MaxPieces = MaxWalls + (BoardSize2 - MaxWalls) / 2;

    // A wall has BoardSize offsets; every other piece has fewer.
    static const int MaxOffsets = BoardSize;

    typedef typename std::conditional<
        (MaxPieces * OffsetBits <= 64), uint64_t, unsigned __int128>::type State;

    static_assert(BoardSize - MinPieceSize < (1 << OffsetBits),
        "OffsetBits is too small for this BoardSize");
    static_assert(MaxPieces * OffsetBits <= 8 * sizeof(State),
        "packed state does not fit");

    struct PieceInfo {
        bb mask[MaxOffsets];       // cells covered at each offset
        bb firstCell[MaxOffsets];  // leading cell, must be free to shift down
        bb lastCell[MaxOffsets];   // trailing cell, must be free to shift up
        int base;                  // board position at offset zero
        int stride;
        int size;
        int maxOffset;
        bool horz;
    };

    struct Node {
        State state;
        bb horz;
        bb vert;
        int32_t distance;
    };

    static int Offset(const State state, const int i) {
        return (int)((state >> (i * OffsetBits)) & (((State)1 << OffsetBits) - 1));
    }

    static State WithOffset(const State state, const int i, const int offset) {
        const int shift = i * OffsetBits;
        const State mask = (((State)1 << OffsetBits) - 1) << shift;
        return (state & ~mask) | ((State)offset << shift);
    }

    // The primary piece is always piece zero, and it sits on the target exactly
    // at its largest offset.
    bool IsSolved(const State state) const {
        return Offset(state, 0) == m_Pieces[0].maxOffset;
    }

    void BuildPieceInfo(const Board &input);
    State StartState(const Board &input) const;
    Board ToBoard(const State state) const;

    // Calls fn(state, horz, vert) once per legal move from the given node. fn
    // returns false to abort, in which case ForEachMove returns false too.
    template <class F>
    bool ForEachMove(const Node &node, F fn) const;

    void BeginCluster();
    void NewGeneration();
    void GrowTable();
    // Returns the state's node index, inserting it if it is new.
    uint32_t Insert(const State state, const bb horz, const bb vert);
    uint32_t Find(const State state) const;

    PieceInfo m_Pieces[MaxPieces];
    int m_NumPieces = 0;

    std::vector<Node> m_Nodes;
    std::vector<uint64_t> m_Table;  // (generation << 32) | (node index + 1)
    std::vector<uint32_t> m_Queue;  // backward pass queue of node indexes
    std::vector<bool> m_PieceMoved;
    uint64_t m_TableMask = 0;
    uint32_t m_Generation = 0;
    Solver m_Solver;

    // Adjacency recorded during the forward pass, so the backward pass can walk
    // the graph without regenerating moves or hashing anything. Nodes are
    // expanded in index order, so each node's neighbors form one contiguous run
    // and m_EdgeStart[i] .. m_EdgeStart[i + 1] delimits node i's.
    std::vector<uint32_t> m_Edges;
    std::vector<uint32_t> m_EdgeStart;
    bool m_HaveEdges = false;

    bool m_Canonical = false;
    bool m_Solvable = false;
    bool m_Minimal = false;
    int m_NumStates = 0;
    int m_MaxDistance = 0;
    Board m_Unsolved;
    std::vector<int> m_Distances;
};
