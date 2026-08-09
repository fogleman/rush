#pragma once

#include <array>

#include "bb.h"

const int BoardSize = 7;
const int PrimaryRow = 3;
const int PrimarySize = 2;
const int MinPieceSize = 2;
const int MaxPieceSize = 3;
const int MinWalls = 0;
const int MaxWalls = 0;
const int NumWorkers = 8;

// Only puzzles needing at least this many moves are reported, and a cluster is
// abandoned as early as it can be proven not to qualify. Zero builds the full
// database; a tail threshold turns the backward pass, the minimality test and
// the histogram into rounding errors, which is what makes a top-N run at 6x6
// or 7x7 affordable. See Cluster::Explore.
const int MinMoves = 30;

// Positions in the old packed enumeration, kept as a record of how big each
// board is. Progress is no longer measured against them: goal enumeration emits
// far fewer positions (124,886 at 5x5, 88,914,655 at 6x6), and work is addressed
// by row combination now -- see Enumerator::NumRowCombos, which is 11^4 at 5x5,
// 22^5 at 6x6 and 41^6 at 7x7.
//
//        1,348 // 4x4
//        9,803 // 4x4, 0-1 walls
//       33,952 // 4x4, 0-2 walls
//       76,837 // 4x4, 0-3 walls
//      268,109 // 5x5
//    2,988,670 // 5x5, 0-1 walls
//   16,330,430 // 5x5, 0-2 walls
//  243,502,786 // 6x6
// 3,670,622,352 // 6x6, 0-1 walls
// 27,403,231,255 // 6x6, 0-2 walls
// 561,276,504,437 // 7x7

const int BoardSize2 = BoardSize * BoardSize;
const bb BoardMask =
    BoardSize2 >= 64 ? ~(bb)0 : (((bb)1 << BoardSize2) - 1);
const int Target = PrimaryRow * BoardSize + BoardSize - PrimarySize;
const int H = 1; // horizontal stride
const int V = BoardSize; // vertical stride
const bool DoWalls = MinPieceSize == 1;

const std::array<bb, BoardSize> RowMasks = []() {
    std::array<bb, BoardSize> rowMasks;
    for (int y = 0; y < BoardSize; y++) {
        bb mask = 0;
        for (int x = 0; x < BoardSize; x++) {
            const int i = y * BoardSize + x;
            mask |= (bb)1 << i;
        }
        rowMasks[y] = mask;
    }
    return rowMasks;
}();

const std::array<bb, BoardSize> ColumnMasks = []() {
    std::array<bb, BoardSize> columnMasks;
    for (int x = 0; x < BoardSize; x++) {
        bb mask = 0;
        for (int y = 0; y < BoardSize; y++) {
            const int i = y * BoardSize + x;
            mask |= (bb)1 << i;
        }
        columnMasks[x] = mask;
    }
    return columnMasks;
}();

const bb TopRow = RowMasks.front();
const bb BottomRow = RowMasks.back();
const bb LeftColumn = ColumnMasks.front();
const bb RightColumn = ColumnMasks.back();

// Reflecting a board top to bottom maps the enumerated set onto itself when the
// primary row is its own mirror image: the primary piece and the target stay put,
// every other piece keeps its size and orientation, and wall counts are
// unchanged. So the mirror of a legal position is another legal position with
// the same solution length, and half of the space is redundant. This only holds
// on odd boards whose primary row is the middle one -- at 6x6 with primary row 2
// the reflection moves the primary row, and there is nothing to exploit.
//
// Left-to-right reflection is never a symmetry: it would move the target off the
// right edge.
//
// See Enumerator::EnumerateRowCombo, which retires half the space on this, and
// Cluster::Explore, which picks one representative per mirror pair.
const bool DoVertSymmetry = BoardSize % 2 == 1 && PrimaryRow == BoardSize / 2;

// The board reflected top to bottom: row y becomes row BoardSize-1-y. Applied to
// a horizontal or a vertical mask separately, it gives the corresponding mask of
// the mirrored board -- a reflection cannot turn one orientation into the other.
// The shifts are all compile-time constants once the loop is unrolled.
inline bb ReflectV(const bb mask) {
    const bb row = ((bb)1 << BoardSize) - 1;
    bb result = 0;
    for (int y = 0; y < BoardSize; y++) {
        result |= ((mask >> (y * BoardSize)) & row)
            << ((BoardSize - 1 - y) * BoardSize);
    }
    return result;
}
