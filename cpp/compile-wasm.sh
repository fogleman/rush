#!/bin/bash
# Builds the Rush Hour solver (src/wasm.cpp solve() entry point) as a
# WebAssembly ES6 module for the bubblyclouds unblockrace app.
#
# Regeneration workflow (run after ANY change to the C++ solver sources):
#   1. Activate emsdk:  source /path/to/emsdk/emsdk_env.sh   (emcc must be on PATH)
#   2. From cpp/:       ./compile-wasm.sh
#   3. Commit the regenerated artifacts in the monorepo:
#        build/wasm/solverWasm.js   (glue, generated - do not edit)
#        build/wasm/solver.wasm     (binary, fetched at runtime)
#      The TS loader passes locateFile to point the glue at /solver/solver.wasm,
#      so the default .wasm filename baked into the glue is irrelevant.
#
# Notes:
#   - RUSH_BOARD_SIZE=6 / RUSH_MAX_PIECE_SIZE=6 select the app's 6x6 boards
#     (the native default stays 5x5 for the enumeration tools).
#   - bb.cpp is excluded: the solver path never calls BitboardString /
#     RandomBitboard.
#   - -fwasm-exceptions is required because Board(std::string) throws on
#     invalid input and wasm.cpp catches it to return "invalid: <reason>".

set -euo pipefail
cd "$(dirname "$0")"

# GROWABLE_ARRAYBUFFERS=0: with ALLOW_MEMORY_GROWTH, emscripten 6 defaults to a
# resizable ArrayBuffer, which Chrome's TextDecoder rejects at runtime
# ("The provided ArrayBuffer value must not be resizable"); 0 restores the
# classic buffer-replacement growth that TextDecoder accepts.
emcc -O3 -flto -std=c++17 -DRUSH_BOARD_SIZE=6 -DRUSH_MAX_PIECE_SIZE=6 \
  -fwasm-exceptions \
  src/wasm.cpp src/board.cpp src/piece.cpp src/move.cpp src/solver.cpp \
  -sEXPORTED_FUNCTIONS=_solve -sEXPORTED_RUNTIME_METHODS=ccall \
  -sMODULARIZE=1 -sEXPORT_ES6=1 -sENVIRONMENT=web -sALLOW_MEMORY_GROWTH=1 \
  -sGROWABLE_ARRAYBUFFERS=0 \
  -o build/wasm/solverWasm.js

# Bundlers (Next.js webpack/turbopack) statically resolve `new URL(..., import.meta.url)`
# and fail because the .wasm is served from public/, not alongside the glue. The consumer
# always passes Module.locateFile, so this fallback branch is dead code — replace it with
# a plain string to keep the bundler out of it.
sed -i '' 's|new URL("solverWasm.wasm",import.meta.url).href|"solverWasm.wasm"|' build/wasm/solverWasm.js

