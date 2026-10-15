# 🧀 queso: arm32 disassembler

queso is a arm32 disassembler written in golang. I originally wrote this to interop with my emulation projects, allowing for easier debugging. It allows for different arm32 versions, and when compared to binutils objdump and other disassemblers, tries to be historically accurate on older arm32 versions.

# Features
- Supports specific arm32 versions (armv2, armv3, armv4t, and armv5te)
- Supports Arm and Thumb instructions
- Supports Big and Little Order / Endianness
- Supports General and Convention Register Names (r15 vs pc) (and Custom!)
- Uses Unified Assembler Language (UAL) syntax

# How to Use

## Webpage

Use [guacemulator.com/queso](https://guacemulator.com/queso/). Or see building for hosting your own.

## Cli

See releases for cli version.

[TODO]

## Vendoring C

[TODO]

## Vendoring Golang

[TODO]

# Building

## Webpage


# Todos
- Thumb2
- Modern architecture versions
- Arm Classic

# AI disclosure
This project was written by a human; however test cases were originally written by AI.
