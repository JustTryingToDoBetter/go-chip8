# go-chip8

A CHIP-8 emulator written in Go. The emulator core lives in `internal/chip8`, and the runnable desktop frontend uses Ebiten from `cmd/chip8`.

## Features

- 4 KiB CHIP-8 memory with programs loaded at `0x200`
- Built-in CHIP-8 font loaded at `0x050`
- 64x32 monochrome display
- 16 8-bit general registers, `V0` through `VF`
- Index register `I`, program counter `PC`, stack, delay timer, and sound timer
- CHIP-8 keypad state with the common keyboard mapping
- Unit tests for core opcodes, drawing, timers, register storage, and key waits
- Bundled test ROMs in `roms/`

## Project Structure

```text
go-chip8/
├─ cmd/
│  └─ chip8/
│     └─ main.go          # Ebiten window, input mapping, emulator loop
├─ internal/
│  └─ chip8/
│     ├─ cpu.go           # CPU state, opcode fetch/decode/execute, display, timers
│     ├─ cpu_test.go      # Unit tests for implemented CHIP-8 behavior
│     └─ font.go          # CHIP-8 font sprites
├─ roms/                  # Local test ROMs
├─ go.mod
└─ go.sum
```

More implementation detail is in [docs/implementation.md](docs/implementation.md).

## Requirements

- Go installed
- A desktop environment supported by Ebiten

The project currently uses Ebiten for windowing, rendering, and keyboard input.

## Running

Run the emulator with a ROM path:

```powershell
go run ./cmd/chip8 ./roms/ibm.ch8
```

Other useful bundled ROMs:

```powershell
go run ./cmd/chip8 ./roms/chip8-logo.ch8
go run ./cmd/chip8 ./roms/corax.ch8
go run ./cmd/chip8 ./roms/flags.ch8
go run ./cmd/chip8 ./roms/quirks.ch8
```

The binary expects exactly one argument:

```text
chip8 <rom-path>
```

## Controls

The CHIP-8 keypad is mapped to the keyboard like this:

```text
CHIP-8 keypad       Keyboard
1 2 3 C             1 2 3 4
4 5 6 D             Q W E R
7 8 9 E             A S D F
A 0 B F             Z X C V
```

## Testing

Run all tests:

```powershell
go test ./...
```

Run only the CPU package tests:

```powershell
go test ./internal/chip8
```

The unit tests cover opcode execution at the CPU level. The ROMs in `roms/` are useful for manual compatibility checks in the Ebiten frontend.

## Current CHIP-8 Behavior

The core implements the standard baseline opcodes needed by the included logo, Corax, flags, and keypad-style test ROMs.

Notable behavior:

- `8XY1`, `8XY2`, and `8XY3` reset `VF` to `0`.
- `8XY6` and `8XYE` use `VY` as the shift source and store the result in `VX`.
- `BNNN` jumps to `NNN + V0`.
- `FX55` and `FX65` do not increment `I`.
- `FX0A` waits for a key press and then waits for that key to be released before storing it in `VX`.
- Sprites wrap around screen edges.

These choices match one common CHIP-8 baseline. Some ROMs expect different interpreter quirks, so quirk configuration is a good future improvement.

## Known Limitations

- No audio output is wired to the sound timer yet.
- Timers are decremented once per Ebiten update, not from a separate 60 Hz clock abstraction.
- No configurable compatibility profiles for CHIP-8, SUPER-CHIP, or XO-CHIP quirks.
- Rendering uses direct per-pixel `Set` calls, which is simple but not optimized.
- The emulator exits when it encounters an unknown opcode.

## Development Notes

When changing CPU behavior, prefer adding a small opcode-level test in `internal/chip8/cpu_test.go`. This keeps compatibility decisions explicit and makes regressions easier to diagnose before testing full ROMs.
