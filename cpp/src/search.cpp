#include "search.h"

#include <limits>
#include <list>
#include <unordered_map>
#include <unordered_set>

/*
canonical?
solvable?
canonical unsolved board
max distance
distance counts
# of reachable states
*/

int ReachableStates(const Board &input) {
    bool canonical = true;
    std::vector<Move> moves;
    std::list<Board> queue;
    queue.push_back(input);
    std::list<Board> unsolveQueue;
    std::unordered_map<BoardKey, int> distance;
    distance[input.Key()] = -1;
    const int sentinel = std::numeric_limits<int>::max();
    while (!queue.empty()) {
        Board &board = queue.front();
        if (canonical && board < input) {
            canonical = false;
            // break;
        }
        if (board.Solved()) {
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

    const int solvedCount = distance.size();
    if (solvedCount == 0) {
        // not solvable
    }

    int maxDistance = 0;
    Board maxDistanceBoard(input);
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
                    maxDistanceBoard = board;
                } else if (d == maxDistance) {
                    if (board < maxDistanceBoard) {
                        maxDistanceBoard = board;
                    }
                }
            }
            board.UndoMove(move);
        }
        unsolveQueue.pop_front();
    }

    std::vector<int> distanceCounts(maxDistance + 1);
    for (const auto &item : distance) {
        distanceCounts[item.second]++;
    }

    std::cout << maxDistance << std::endl;

    for (int i = 0; i <= maxDistance; i++) {
        std::cout << i << " " << distanceCounts[i] << std::endl;
    }

    return distance.size();
}
