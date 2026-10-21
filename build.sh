#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/build"

# Parse flags (accepts -cli or --cli, etc.)
ALL=0
CLI=0
WEB=0
C=0
for arg in "$@"; do
    case "$arg" in
    -all | --all) ALL=1 ;;
    -cli | --cli) CLI=1 ;;
    -web | --web) WEB=1 ;;
    -c | --c) C=1 ;;
    *)
        echo "Unknown option: $arg" >&2
        exit 1
        ;;
    esac
done

if ((! (ALL || CLI || WEB || C))); then
    echo "Usage: ./build.sh [-cli] [-web] [-c] [-all]"
    exit 1
fi

# Windows shells that run bash (Git Bash, MSYS2, Cygwin) need .exe
case "$(uname -s)" in
MINGW* | MSYS* | CYGWIN*) EXE_NAME="queso.exe" ;;
*) EXE_NAME="queso" ;;
esac

new_out_dir() {
    mkdir -p "$BUILD_DIR/$1"
    echo "$BUILD_DIR/$1"
}

build_cli() {
    local out_dir
    out_dir="$(new_out_dir cli)"

    (cd "$SCRIPT_DIR/cmd/cli" && go build -o "$out_dir/$EXE_NAME" .)

    echo "Queso - Cli build completed."
}

build_web() {
    local out_dir
    out_dir="$(new_out_dir web)"

    (
        cd "$SCRIPT_DIR/cmd/web"
        GOOS=js GOARCH=wasm go build -o "$out_dir/main.wasm"
        cp -f index.html queso.js "$out_dir/"
        cp -f "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$out_dir/"
    )

    echo "Queso - Web build completed."
}

build_c() {
    local out_dir
    out_dir="$(new_out_dir c)"

    (
        cd "$SCRIPT_DIR/cmd/c"
        CGO_ENABLED=1 go build -buildmode=c-archive -o queso.a c.go

        mv -f queso.a queso.h "$out_dir/"
        cp -f libqueso.h "$out_dir/"
    )

    echo "Queso - C build completed."
}

((CLI || ALL)) && build_cli
((WEB || ALL)) && build_web
((C || ALL)) && build_c

echo "Queso completed."
