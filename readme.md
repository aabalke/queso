# <img width="36" height="36" alt="icon" src="https://github.com/user-attachments/assets/ce40228a-b93b-4dee-8522-ac490d58192f" /> queso: arm32 disassembler

queso is a arm32 disassembler written in golang. I originally wrote this to interop with my emulation projects, allowing for easier debugging. It allows for different arm32 versions, and when compared to binutils objdump and other disassemblers, tries to be historically accurate on older arm32 versions.

[Video Breakdown](https://youtu.be/YrqZTUkqksE)

# Features
- Supports specific arm32 versions (ARMv1, ARMv2, ARMv2a, ARMv3, ARMv3M, ARMv4xM, ARMv4, ARMv4TxM, ARMv4T, ARMv5xM, ARMv5, ARMv5TxM, ARMv5T, ARMv5TExP, and ARMv5TE)
- Supports Arm and Thumb instructions
- Supports Big and Little Order / Endianness
- Supports General and Convention Register Names (r15 vs pc) (and Custom!)
- Uses Unified Assembler Language (UAL) syntax

# Getting Started

queso is packaged in 4 ways: a web version, a cli version, a vendorable golang version and a vendorable c version.

## Webpage

See [guacemulator.com/queso](https://guacemulator.com/queso/).

## Cli

The command line version takes individual instructions as args or a file as input. Output can be pushed to a file, and all isa / architecture options can be altered with flags.

### file method
```
queso -input=filpath.bin
```

### arg method
```
queso -order=big 0EF0B0E1 010081E0
```

### Flags
-input=[filepath]: sets the input file for disassembly

-output=[filepath]: sets the output file for disassembly

-thumb: forces disassembly in thumb mode

-addr: outputs address of each instruction

-opcode: outputs opcode of each instruction

-offset=[int]: sets starting base of address

-order=[little,big,native]: sets order/endianness of disassembly. Default is little.

-regs=[convention,general,"comma separated list of register names"]: sets register names to convention (lr, pc), general (r14, r15), or custom. Custom starts with r0 and should be in the format "r0,r1,r2,r3,r4...". Default is Convention.

-undefined=[string]: sets output on undefined instruction

-arch=ArchVersion: sets Arm32 version. Default is ARMv5TE. All Versions ARMv1 through ARMv5TE are valid.

-version: prints queso version

-profile: uses pprof to profile queso

-h: help

## C

The C version requires libqueso.h, queso.a, and queso.h. libqueso.h provides types and enums for queso decoding options. The C version take in a byte array to decode, and a result type to write data to. See ./examples/c/ for a basic example.

## Golang

The golang version can be used within a go program though importing github.com/aabalke/queso. An io.Reader is used as instruction input, allowing for flexibility. See ./examples/go/ for a basic example.

# Building

Building from source requires go version 1.27+. Use build.ps1 or build.sh depending on shell.

## Webpage

After building, copy files in ./build/web/ to the hosting site. Files are 100% client side. All files are required.

## Cli

After building, use queso(.exe) available in ./build/cli/ .

## C

After building, copy files in ./build/c/ to c project needing queso disassembler. All files are required.

## Golang

Vendoring can be completed by including github.com/aabalke/queso.

# Todos
- Thumb2
- v6+ versions
- Arm Classic
- streamable version, web with large files
