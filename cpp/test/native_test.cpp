// Native verification harness for the wasm solver path.
//
// Build (from cpp/):
//   mkdir -p build && clang++ -std=c++17 -O2 \
//     -DRUSH_BOARD_SIZE=6 -DRUSH_MAX_PIECE_SIZE=6 \
//     test/native_test.cpp src/wasm.cpp src/board.cpp src/piece.cpp \
//     src/move.cpp src/solver.cpp -o build/native_test
//   ./build/native_test
//
// Fixtures are sampled from the app's puzzles.json (boardString +
// movesRequired) across the full difficulty range.

#include <chrono>
#include <cstdio>
#include <cstdlib>
#include <set>
#include <string>
#include <vector>

#include "../src/board.h"
#include "../src/solver.h"

extern "C" const char* solve(const char* boardString);

namespace {

int failures = 0;

void check(bool ok, const std::string &what) {
    if (!ok) {
        failures++;
        std::printf("FAIL: %s\n", what.c_str());
    }
}

double solveTimed(const std::string &desc, std::string &out) {
    const auto t0 = std::chrono::steady_clock::now();
    out = solve(desc.c_str());
    const auto t1 = std::chrono::steady_clock::now();
    return std::chrono::duration<double, std::milli>(t1 - t0).count();
}

std::vector<char> pieceLabels(const std::string &desc) {
    std::set<char> labelSet;
    for (const char c : desc) {
        if (c == '.' || c == 'o' || c == 'x') {
            continue;
        }
        labelSet.insert(c);
    }
    return std::vector<char>(labelSet.begin(), labelSet.end());
}

struct ParsedMove {
    char label;
    int steps;
};

// Parses "ok A+1 C-2 ..." into moves; returns false if not an ok result.
bool parseOk(const std::string &result, std::vector<ParsedMove> &moves) {
    moves.clear();
    if (result != "ok" && result.rfind("ok ", 0) != 0) {
        return false;
    }
    size_t i = 2;
    while (i < result.length()) {
        if (result[i] == ' ') {
            i++;
            continue;
        }
        const char label = result[i++];
        const char sign = result[i++];
        int steps = 0;
        while (i < result.length() && result[i] != ' ') {
            steps = steps * 10 + (result[i++] - '0');
        }
        moves.push_back({label, sign == '-' ? -steps : steps});
    }
    return true;
}

// Replays the moves, asserting each is legal (present in Board::Moves())
// and that the final position is solved.
bool replay(const std::string &desc, const std::vector<ParsedMove> &moves) {
    Board board(desc);
    const auto labels = pieceLabels(desc);
    std::vector<Move> buf;
    for (const auto &pm : moves) {
        int index = -1;
        for (size_t i = 0; i < labels.size(); i++) {
            if (labels[i] == pm.label) {
                index = (int)i;
                break;
            }
        }
        if (index < 0) {
            return false;
        }
        board.Moves(buf);
        bool legal = false;
        for (const auto &m : buf) {
            if (m.Piece() == index && m.Steps() == pm.steps) {
                legal = true;
                break;
            }
        }
        if (!legal) {
            return false;
        }
        board.DoMove(index, pm.steps);
    }
    return board.Solved();
}

struct Fixture {
    const char* desc;
    int movesRequired;
};

const Fixture Fixtures[] = {
    {"FBBCCoFoGoooAAGooooDDooooEEooooooooo", 4},
    {"oBBBooooGHCCAAGHoIFoDDoIFoooooooooEE", 6},
    {"FBBBooFoHICCAAHIJKoGDDJKoGooooEEooox", 8},
    {"oGBBoKoGCCJKAAooJKFoHIooFoHIDDooooEE", 11},
    {"oBBCCooxDDKoAAIJKoHoIJEEHoFFoLoGGooL", 13},
    {"BBHoooFoHCCKFAAoJKoGDDJooGoIxoEEoIoo", 16},
    {"BBHoooFGHCCKFGAAJKDDoIJooooIEEoxoooo", 18},
    {"FoxooKFoHBBKAAHIoooGoICCoGDDJoEEooJo", 21},
    {"BBBoxoooHCCKAAHIoKGooIDDGoEEJoooFFJo", 23},
    {"BBooxoGooJCCGAAJKLGHDDKLoHIEEoFFIooo", 26},
    {"FooxoxFoBBoKAAHIoKoGHICCoGDDJoEEooJo", 28},
    {"oBBxoLooICCLAAIoKoGHDDKoGHoJEExFFJoo", 31},
    {"BBIoxLCCIooLGHAAKMGHDDKMGooJEEFFoJoo", 33},
    {"ooIBBBGoIJCCGAAJoLoHDDKLoHEEKoFFooKx", 36},
    {"xBBCCLooIDDLAAIoKoxHEEKooHoJFFoooJGG", 38},
    {"BBHooLoGHCCLoGAAKoDDIoKoFoIJxoFEEJox", 41},
    {"FBBoKLFGHoKLFGHAAMCCIJoMooIJDDoEEoxo", 43},
    {"xoIBBBGoIJCCGAAJKLoHDDKLoHEEKoFFooox", 46},
    {"BBBKLMHCCKLMHoAALoDDJooooIJEEooIFFGG", 48},
    {"BBBICCGooIoJGAAIoJooHDDKEEHooKooxoFF", 50},
};

} // namespace

int main() {
    std::string result;
    std::vector<ParsedMove> moves;

    std::printf("== Fixtures from puzzles.json ==\n");
    for (const auto &fixture : Fixtures) {
        const std::string desc(fixture.desc);
        const double ms = solveTimed(desc, result);
        const bool ok = parseOk(result, moves);
        check(ok, desc + " expected ok, got: " + result);
        if (!ok) {
            continue;
        }
        check((int)moves.size() == fixture.movesRequired,
              desc + " expected " + std::to_string(fixture.movesRequired) +
              " moves, got " + std::to_string(moves.size()));
        check(replay(desc, moves), desc + " replay failed");
        std::printf("%s moves=%2d expected=%2d time=%8.2fms %s\n",
                    desc.c_str(), (int)moves.size(), fixture.movesRequired, ms,
                    fixture.movesRequired >= 45 ? "<- >=45 mover" : "");
    }

    std::printf("== Special cases ==\n");

    // wall directly between primary and exit -> unsolvable
    {
        const std::string desc = "oooooo" "oooooo" "AAxooo" "oooooo" "oooooo" "oooooo";
        const double ms = solveTimed(desc, result);
        check(result == "unsolvable", "wall-blocked expected unsolvable, got: " + result);
        check(ms < 2000, "wall-blocked took too long: " + std::to_string(ms) + "ms");
        std::printf("wall-blocked: %s time=%.2fms\n", result.c_str(), ms);
    }

    // exit column fully packed with immovable vertical pieces (no walls) -> unsolvable
    {
        const std::string desc = "oooooB" "oooooB" "AAoooC" "oooooC" "oooooD" "oooooD";
        const double ms = solveTimed(desc, result);
        check(result == "unsolvable", "self-blocked expected unsolvable, got: " + result);
        check(ms < 2000, "self-blocked took too long: " + std::to_string(ms) + "ms");
        std::printf("self-blocked: %s time=%.2fms\n", result.c_str(), ms);
    }

    // vertical piece across the exit row pinned by walls on both ends -> unsolvable
    {
        const std::string desc = "ooooox" "oooooB" "AAoooB" "oooooB" "ooooox" "oooooo";
        const double ms = solveTimed(desc, result);
        check(result == "unsolvable", "wall-pinned expected unsolvable, got: " + result);
        check(ms < 2000, "wall-pinned took too long: " + std::to_string(ms) + "ms");
        std::printf("wall-pinned: %s time=%.2fms\n", result.c_str(), ms);
    }

    // already solved -> "ok" with no moves
    {
        const std::string desc = "oooooo" "oooooo" "ooooAA" "oooooo" "oooooo" "oooooo";
        solveTimed(desc, result);
        check(result == "ok", "already-solved expected \"ok\", got: " + result);
        std::printf("already-solved: %s\n", result.c_str());
    }

    // primary on row 0 (runtime target) -> 2 moves
    {
        const std::string desc = "AAoBoo" "oooBoo" "oooooo" "oooooo" "oooooo" "oooooo";
        solveTimed(desc, result);
        const bool ok = parseOk(result, moves);
        check(ok, "row-0 primary expected ok, got: " + result);
        if (ok) {
            check(moves.size() == 2, "row-0 primary expected 2 moves, got " +
                  std::to_string(moves.size()));
            check(replay(desc, moves), "row-0 primary replay failed");
        }
        std::printf("row-0 primary: %s\n", result.c_str());
    }

    // walled but solvable -> 1 move (wall away from the exit row)
    {
        const std::string desc = "ooxooo" "oooooo" "ooAAoo" "oooooo" "oooooo" "oooooo";
        solveTimed(desc, result);
        const bool ok = parseOk(result, moves);
        check(ok, "walled-solvable expected ok, got: " + result);
        if (ok) {
            check(moves.size() == 1, "walled-solvable expected 1 move, got " +
                  std::to_string(moves.size()));
            check(replay(desc, moves), "walled-solvable replay failed");
        }
        std::printf("walled-solvable: %s\n", result.c_str());
    }

    // invalid inputs
    {
        solveTimed("AAoo", result);
        check(result.rfind("invalid:", 0) == 0, "short string expected invalid, got: " + result);
        solveTimed(std::string(36, 'o'), result);
        check(result.rfind("invalid:", 0) == 0, "empty board expected invalid, got: " + result);
        const std::string vertical = "Aooooo" "Aooooo" "oooooo" "oooooo" "oooooo" "oooooo";
        solveTimed(vertical, result);
        check(result == "invalid: primary must be horizontal",
              "vertical primary expected invalid, got: " + result);
    }

    if (failures == 0) {
        std::printf("ALL TESTS PASSED\n");
        return 0;
    }
    std::printf("%d FAILURE(S)\n", failures);
    return 1;
}
