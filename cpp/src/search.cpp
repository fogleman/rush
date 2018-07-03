#include "search.h"

#include <list>
#include <unordered_map>
#include <unordered_set>

int ReachableStates(const Board &input) {
    bool canonical = true;

    std::unordered_set<BoardKey> seen;
    seen.emplace(input.Key());

    std::list<Board> queue;
    queue.push_back(input);

    std::unordered_map<BoardKey, int> distance;
    std::list<Board> unsolveQueue;

    std::vector<Move> moves;
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
            if (seen.emplace(board.Key()).second) {
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
    int maxDistanceCount = 1;
    Board maxDistanceBoard(input);
    while (!unsolveQueue.empty()) {
        Board &board = unsolveQueue.front();
        const int d = distance[board.Key()] + 1;
        board.Moves(moves);
        for (const auto &move : moves) {
            board.DoMove(move);
            const auto item = distance.find(board.Key());
            if (item == distance.end() || item->second > d) {
                distance[board.Key()] = d;
                unsolveQueue.push_back(board);
                if (d > maxDistance) {
                    maxDistance = d;
                    maxDistanceCount = 1;
                    maxDistanceBoard = board;
                } else if (d == maxDistance) {
                    maxDistanceCount++;
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

    std::cout << maxDistance << " " << maxDistanceCount << std::endl;

    for (int i = 0; i <= maxDistance; i++) {
        std::cout << i << " " << distanceCounts[i] << std::endl;
    }

    return seen.size();
}
