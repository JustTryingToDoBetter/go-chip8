// cmd/chip8/main.go
package main

import (
	"fmt"
	"log"

	"github.com/JustTryingToDoBetter/go-chip8/internal/chip8"
)

func main() {
	cpu := chip8.New()

	program := []byte{
		0x60, 0x0A, // V0 = 10
		0x70, 0x01, // V0 += 1
		0x61, 0x05, // V1 = 5
		0xA3, 0x00, // I = 0x300
	}

	if err := cpu.LoadProgram(program); err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 4; i++ {
		pcBefore := cpu.PC

		opcode, err := cpu.Step()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf(
			"PC=0x%03X OPCODE=0x%04X V0=%d V1=%d I=0x%03X SP=%d\n",
			pcBefore,
			opcode,
			cpu.V[0],
			cpu.V[1],
			cpu.I,
			cpu.SP,
		)
	}
}
