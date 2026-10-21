# <img width="36" height="36" alt="icon" src="https://github.com/user-attachments/assets/ce40228a-b93b-4dee-8522-ac490d58192f" /> queso: arm32 disassembler

queso is a arm32 disassembler written in golang. I originally wrote this to interop with my emulation projects, allowing for easier debugging. It allows for different arm32 versions, and when compared to binutils objdump and other disassemblers, tries to be historically accurate on older arm32 versions.

# Features
- Supports specific arm32 versions (ARMv1, ARMv2, ARMv2a, ARMv3, ARMv3M, ARMv4xM, ARMv4, ARMv4TxM, ARMv4T, ARMv5xM, ARMv5, ARMv5TxM, ARMv5T, ARMv5TExP, and ARMv5TE)
- Supports Arm and Thumb instructions
- Supports Big and Little Order / Endianness
- Supports General and Convention Register Names (r15 vs pc) (and Custom!)
- Uses Unified Assembler Language (UAL) syntax

# Getting Started

queso is packaged in 4 ways: a web version, a cli version, a vendorable golang version and a vendorable c version.

# How to Use / Building

building requires 

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
- v6+ versions
- Arm Classic
- streamable version, web with large files
