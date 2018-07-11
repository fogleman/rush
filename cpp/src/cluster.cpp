#include "cluster.h"

#include <deque>
#include <limits>
#include <boost/unordered_map.hpp>

#include "solver.h"

Cluster::Cluster(const uint64_t id, const Board &input) :
    m_ID(id),
    m_Canonical(false),
    m_Solvable(false),
    m_Minimal(false),
    m_NumStates(0)
{
    // move generation buffer
    std::vector<Move> moves;

    // exploration queue
    std::deque<Board> queue;
    queue.push_back(input);

    // unsolve queue
    std::deque<Board> unsolveQueue;

    // large sentinel distance when distance is not yet known
    const int sentinel = std::numeric_limits<int>::max();

    // maps keys to distance from nearest goal state
    boost::unordered_map<BoardKey, int> distance;
    distance[input.Key()] = sentinel;

    // explore reachable nodes
    while (!queue.empty()) {
        Board &board = queue.front();
        if (board.Solved()) {
            m_Solvable = true;
            distance[board.Key()] = 0;
            unsolveQueue.push_back(board);
        }
        board.Moves(moves);
        for (const auto &move : moves) {
            board.DoMove(move);
            if (board < input) {
                // not canonical, exit early
                // and don't count non-canonical solvable boards
                m_Solvable = false;
                return;
            }
            if (distance.emplace(board.Key(), sentinel).second) {
                queue.push_back(board);
            }
            board.UndoMove(move);
        }
        queue.pop_front();
    }

    m_Canonical = true;
    m_NumStates = distance.size();

    if (!m_Solvable) {
        // nothing else to do if it's not solvable
        return;
    }

    // determine how far each state is from a goal state
    int maxDistance = 0;
    m_Unsolved = input;
    while (!unsolveQueue.empty()) {
        Board &board = unsolveQueue.front();
        const int d = distance[board.Key()] + 1;
        board.Moves(moves);
        for (const auto &move : moves) {
            board.DoMove(move);
            const auto item = distance.find(board.Key());
            if (item->second > d) {
                item->second = d;
                unsolveQueue.push_back(board);
                if (d > maxDistance) {
                    maxDistance = d;
                    m_Unsolved = board;
                } else if (d == maxDistance) {
                    if (board < m_Unsolved) {
                        m_Unsolved = board;
                    }
                }
            }
            board.UndoMove(move);
        }
        unsolveQueue.pop_front();
    }

    // determine if unsolved board is minimal
    Solver solver;
    const auto solution = solver.Solve(m_Unsolved);
    const int numPieces = input.Pieces().size();
    std::vector<bool> pieceMoved(numPieces, false);
    for (const auto &move : solution.Moves()) {
        pieceMoved[move.Piece()] = true;
    }
    for (int i = 1; i < pieceMoved.size(); i++) {
        if (pieceMoved[i]) {
            continue;
        }
        Board board(m_Unsolved);
        board.RemovePiece(i);
        if (solver.CountMoves(board) == maxDistance) {
            return;
        }
    }
    m_Minimal = true;

    // record number of states by distance to goal
    m_Distances.resize(maxDistance + 1);
    for (const auto &item : distance) {
        m_Distances[item.second]++;
    }
}
