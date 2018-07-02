#include "bb.h"

#include "config.h"

std::string BitboardString(const bb b) {
    std::string s(BoardSize2, '.');
    for (int i = 0; i < BoardSize2; i++) {
        const bb mask = (bb)1 << i;
        if ((b & mask) != 0) {
            s[i] = 'X';
        }
    }
    return s;
}
