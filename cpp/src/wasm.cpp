// WebAssembly entry point for the Rush Hour solver.
//
// Exposes a single C function:
//   const char* solve(const char* boardString);
//
// The board string uses '.'/'o' for empty cells, 'x' for walls, and
// 'A'-'Z' for pieces ('A' is the primary horizontal piece). The return
// value is a static string with one of:
//   "ok A+1 C-2 ..."  - solvable; space-separated moves (piece letter +
//                       signed steps), in order
//   "ok"              - already solved
//   "unsolvable"      - no solution exists
//   "invalid: <reason>" - the board string could not be parsed/validated

#include <cstdlib>
#include <set>
#include <string>
#include <vector>

#include "board.h"
#include "solver.h"

extern "C" {

const char* solve(const char* boardString) {
    static std::string result;
    const std::string desc(boardString == nullptr ? "" : boardString);
    if (desc.length() != BoardSize2) {
        result = "invalid: length";
        return result.c_str();
    }
    // Piece indices assigned by Board(std::string) follow the sorted
    // distinct labels; keep the same list to map indices back to letters.
    std::set<char> labelSet;
    for (const char c : desc) {
        if (c == '.' || c == 'o' || c == 'x') {
            continue;
        }
        labelSet.insert(c);
    }
    const std::vector<char> labels(labelSet.begin(), labelSet.end());
    if (labels.empty()) {
        result = "invalid: no pieces";
        return result.c_str();
    }
    if (labels[0] != 'A') {
        result = "invalid: no primary piece";
        return result.c_str();
    }
    try {
        Board board(desc);
        if (board.Pieces()[0].Stride() != H) {
            result = "invalid: primary must be horizontal";
            return result.c_str();
        }
        Solver solver;
        const Solution solution = solver.Solve(board);
        if (!solution.Solvable()) {
            result = "unsolvable";
            return result.c_str();
        }
        result = "ok";
        for (const auto &move : solution.Moves()) {
            result += ' ';
            result += labels[move.Piece()];
            result += move.Steps() >= 0 ? '+' : '-';
            result += std::to_string(std::abs(move.Steps()));
        }
        return result.c_str();
    } catch (const char* reason) {
        result = std::string("invalid: ") + reason;
        return result.c_str();
    }
}

}
