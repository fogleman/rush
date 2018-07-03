#include "cluster.h"

#include <limits>
#include <list>
#include <unordered_map>
#include <unordered_set>

Cluster::Cluster(const uint64_t id, const uint64_t group, const Board &input) :
    m_ID(id),
    m_Group(group),
    m_Canonical(true),
    m_Solvable(false),
    m_NumStates(0)
{
    // move generation buffer
    std::vector<Move> moves;

    // exploration queue
    std::list<Board> queue;
    queue.push_back(input);

    // unsolve queue
    std::list<Board> unsolveQueue;

    // large sentinel distance when distance is not yet known
    const int sentinel = std::numeric_limits<int>::max();

    // maps keys to distance from nearest goal state
    std::unordered_map<BoardKey, int> distance;
    distance[input.Key()] = sentinel;

    // explore reachable nodes
    while (!queue.empty()) {
        Board &board = queue.front();
        if (board < input) {
            // not canonical, exit early
            m_Canonical = false;
            // return;
        }
        if (board.Solved()) {
            if (!m_Solvable || board < m_Solved) {
                m_Solved = board;
            }
            m_Solvable = true;
            distance[board.Key()] = 0;
            unsolveQueue.push_back(board);
        }
        board.Moves(moves);
        for (const auto &move : moves) {
            board.DoMove(move);
            if (distance.emplace(std::make_pair(board.Key(), sentinel)).second) {
                queue.push_back(board);
            }
            board.UndoMove(move);
        }
        queue.pop_front();
    }

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

    // record number of states by distance to goal
    m_Distances.resize(maxDistance + 1);
    for (const auto &item : distance) {
        m_Distances[item.second]++;
    }
}
