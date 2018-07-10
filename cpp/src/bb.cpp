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

bb RandomBitboard(std::mt19937 &gen) {
    std::uniform_int_distribution<int> dis(0, 0xffff);
    const bb a = dis(gen);
    const bb b = dis(gen);
    const bb c = dis(gen);
    const bb d = dis(gen);
    return a << 48 | b << 32 | c << 16 | d;
}
