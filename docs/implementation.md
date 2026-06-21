# Implementation Notes

This document describes how the emulator is currently structured and records the compatibility choices made in the CPU core.

## Runtime Flow

The executable entry point is `cmd/chip8/main.go`.

At startup:

1. `main` expects a ROM path argument.
2. `NewGame` creates a `chip8.CPU` with `chip8.New()`.
3. The ROM is read from disk and copied into memory starting at `0x200`.
4. Ebiten starts the game loop.

Each Ebiten update:

1. `syncInput` clears and rebuilds the CHIP-8 key state from the current keyboard state.
2. The CPU executes `cyclesPerTick` instructions.
3. Delay and sound timers are decremented once.

Each draw:

1. The screen is cleared to black.
2. The CPU display buffer is scanned.
3. Enabled pixels are drawn white.

## CPU State

The CPU state is stored in `internal/chip8/cpu.go`.

Important fields:

- `Memory [4096]byte`: CHIP-8 memory.
- `V [16]byte`: registers `V0` through `VF`.
- `I uint16`: index register.
- `PC uint16`: program counter.
- `Stack [16]uint16` and `SP byte`: subroutine stack.
- `Display [64 * 32]bool`: monochrome display buffer.
- `DelayTimer byte` and `SoundTimer byte`: timers.
- `Keys [16]bool`: CHIP-8 keypad state.

Programs are loaded at `ProgramStart`, currently `0x200`. The built-in font is copied to `FontStart`, currently `0x050`.

## Fetch, Decode, Execute

`Step` is the public instruction step:

1. `Fetch` reads two bytes at `PC`.
2. `Fetch` advances `PC` by two.
3. `Execute` decodes and applies the opcode.

Because `PC` is advanced during fetch, instructions that repeat the current opcode, such as `FX0A` while waiting for input, subtract two from `PC`.

## Display Behavior

`DXYN` draws an `N` byte sprite at coordinates `(VX, VY)`.

Current behavior:

- Coordinates are reduced with modulo screen dimensions before drawing.
- Sprite pixels wrap horizontally and vertically.
- Pixels are XORed into the display buffer.
- `VF` is set to `1` if any enabled pixel is turned off by collision.
- Sprite memory reads are range checked before drawing.

The display dimensions are exported as `ScreenWidth` and `ScreenHeight` for frontend code.

## Key Input

The frontend maps the keyboard to CHIP-8 keys:

```text
CHIP-8 keypad       Keyboard
1 2 3 C             1 2 3 4
4 5 6 D             Q W E R
7 8 9 E             A S D F
A 0 B F             Z X C V
```

`EX9E` skips if the key stored in `VX` is currently pressed.

`EXA1` skips if the key stored in `VX` is currently not pressed.

`FX0A` has stateful behavior:

1. If no key is pressed, `PC` is rewound and the instruction waits.
2. Once a key is pressed, the CPU records that key and keeps waiting.
3. After that same key is released, the key value is stored in `VX` and execution continues.

This matches keypad test ROMs that expect `FX0A` to halt until release.

## Compatibility Choices

CHIP-8 has several historical interpreter variants. This implementation currently hardcodes one baseline behavior.

Current choices:

| Area | Current behavior |
| --- | --- |
| Logical ops | `8XY1`, `8XY2`, and `8XY3` reset `VF` to `0` |
| Shift ops | `8XY6` and `8XYE` use `VY` as the source and store into `VX` |
| Offset jump | `BNNN` jumps to `NNN + V0` |
| Memory ops | `FX55` and `FX65` leave `I` unchanged |
| Index add | `FX1E` adds `VX` to `I` and does not alter `VF` |
| Drawing | Sprites wrap at screen edges |
| Key wait | `FX0A` waits for key release before continuing |

These choices should be made configurable if the emulator grows to support multiple compatibility profiles.

## Implemented Opcode Groups

The CPU currently implements:

- `00E0`, `00EE`
- `1NNN`, `2NNN`
- `3XNN`, `4XNN`, `5XY0`, `9XY0`
- `6XNN`, `7XNN`
- `8XY0`, `8XY1`, `8XY2`, `8XY3`, `8XY4`, `8XY5`, `8XY6`, `8XY7`, `8XYE`
- `ANNN`, `BNNN`, `CXNN`, `DXYN`
- `EX9E`, `EXA1`
- `FX07`, `FX0A`, `FX15`, `FX18`, `FX1E`, `FX29`, `FX33`, `FX55`, `FX65`

Unknown opcodes return an error from `Execute`, which bubbles up through `Step` and stops the frontend loop.

## Testing Strategy

Unit tests live in `internal/chip8/cpu_test.go`.

The tests execute small in-memory programs or call `Execute` directly. They cover:

- Program loading and fetch/step behavior
- Jumps, calls, returns, and skips
- Arithmetic and flag behavior
- Display drawing, wrapping, and collision
- Key skip and key wait behavior
- Timer reads/writes and timer decrementing
- BCD storage and register memory operations

Run tests with:

```powershell
go test ./...
```

Manual ROM checks can be run from the repository root:

```powershell
go run ./cmd/chip8 ./roms/chip8-logo.ch8
go run ./cmd/chip8 ./roms/ibm.ch8
go run ./cmd/chip8 ./roms/corax.ch8
go run ./cmd/chip8 ./roms/flags.ch8
go run ./cmd/chip8 ./roms/quirks.ch8
```

## Good Next Improvements

- Add quirk configuration for shift, memory, jump, logic flag, and clipping behavior.
- Drive delay and sound timers from a clear 60 Hz timing model.
- Add simple audio output for `SoundTimer`.
- Add automated screenshot or framebuffer assertions for known test ROMs.
- Split CPU compatibility settings from frontend concerns.
