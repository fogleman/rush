#include "search.h"

#include <iostream>
#include <list>
#include <unordered_set>

int ReachableStates(const Board &input, uint64_t counter) {
    bool canonical = true;
    int solvedStates = 0;

    std::unordered_set<BoardKey> seen;
    seen.emplace(input.Key());

    std::list<Board> queue;
    queue.push_back(input);

    std::vector<Move> moves;
    while (!queue.empty()) {
        Board &board = queue.front();
        if (canonical && board < input) {
            canonical = false;
            // break;
        }
        if (board.Solved()) {
            solvedStates++;
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

    // if (canonical && solvedStates > 0) {
        std::cout << counter << " " << canonical << " " << input << " " << seen.size() << " " << solvedStates << std::endl;
    // }

    return seen.size();
}
