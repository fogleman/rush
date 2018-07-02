#pragma once

const int BoardSize = 6;
const int PrimaryRow = 2;
const int PrimarySize = 2;
const int MinPieceSize = 2;
const int MaxPieceSize = 3;

const int BoardSize2 = BoardSize * BoardSize;
const int Target = PrimaryRow * BoardSize + BoardSize - PrimarySize;
const int H = 1; // horizontal stride
const int V = BoardSize; // vertical stride
