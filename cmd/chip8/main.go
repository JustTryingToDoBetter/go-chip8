// cmd/chip8/main.go
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/JustTryingToDoBetter/go-chip8/internal/chip8"
)

const defaultSteps = 20

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		log.Fatal("usage: chip8 <rom-path> [steps]")
	}

	romPath := os.Args[1]
	steps := defaultSteps

	if len(os.Args) == 3 {
		parsed, err := strconv.Atoi(os.Args[2])
		if err != nil || parsed <= 0 {
			log.Fatalf("invalid step count :%q", os.Args[2])
		}

		steps = parsed
	}

	cpu := chip8.New()
	if err := cpu.LoadROM(romPath); err != nil {
		log.Fatal(err)
	}
	program := []byte{
		0x60, 0x0A, // V0 = 10
		0x70, 0x01, // V0 += 1
		0x61, 0x05, // V1 = 5
		0xA3, 0x00, // I = 0x300
	}

	if err := cpu.LoadProgram(program); err != nil {
		log.Fatal(err)
	}

	for i := 0; i < steps; i++ {
		pcBefore := cpu.PC

		opcode, err := cpu.Step()
		if err != nil {
			log.Fatalf("step %d failed: %v", i+1, err)
		}

		fmt.Printf(
			"STEP=%03d PC=0x%03X OPCODE=0x%04X V0=%02X V1=%02X V2=%02X I=0x%03X SP=%d DT=%d ST=%d\n",
			i+1,
			pcBefore,
			opcode,
			cpu.V[0],
			cpu.V[1],
			cpu.V[2],
			cpu.I,
			cpu.SP,
			cpu.DelayTimer,
			cpu.SoundTimer,
		)
	}
}
