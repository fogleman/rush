#pragma once

#include <cstdint>
#include <random>
#include <string>

typedef uint64_t bb;

std::string BitboardString(const bb b);
bb RandomBitboard(std::mt19937 &gen);
