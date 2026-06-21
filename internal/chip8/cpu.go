package chip8

import (
	"crypto/rand"
	"fmt"
)

const (
	MemorySize   = 4096  // Total memory size of CHIP-8
	ProgramStart = 0x200 // Start of most CHIP-8 programs
	screenWidth  = 64
	screenHeight = 32
)

type CPU struct {
	Memory [MemorySize]byte // 4K memory

	V [16]byte // 16 registers (V0 to VF)

	I  uint16 // Index register
	PC uint16 // Program counter

	Stack [16]uint16 // Stack for subroutine calls
	SP    byte       // Stack pointer

	Display      [screenWidth * screenHeight]bool //
	DisplayDirty bool

	DelayTimer byte
	SoundTimer byte

	Keys [16]bool
}

func New() *CPU {
	cpu := &CPU{
		PC: ProgramStart,
	}

	copy(cpu.Memory[FontStart:], fontSet[:])

	return cpu
}

func (c *CPU) LoadProgram(program []byte) error {
	if len(program) > MemorySize-ProgramStart {
		return fmt.Errorf("program size exceeds available memory: %d", len(program))
	}

	copy(c.Memory[ProgramStart:], program) // Load program into memory starting at 0x200
	return nil
}

func (c *CPU) Fetch() (uint16, error) {
	if int(c.PC)+1 >= len(c.Memory) {
		return 0, fmt.Errorf("program counter out of bounds: %d", c.PC)
	}

	opcode := uint16(c.Memory[c.PC])<<8 | uint16(c.Memory[c.PC+1]) // Fetch 2 bytes as opcode
	c.PC += 2                                                      // Move to the next instruction

	return opcode, nil
}

func (c *CPU) Step() (uint16, error) {
	opcode, err := c.Fetch()
	if err != nil {
		return 0, err
	}

	if err := c.Execute(opcode); err != nil {
		return 0, err
	}

	return opcode, nil
}

func (c *CPU) Execute(opcode uint16) error {
	nnn := opcode & 0x0FFF            // last 12 bits
	nn := byte(opcode & 0x00FF)       // last 8 bits
	n := byte(opcode & 0x000F)        // last 4 bits
	x := byte((opcode & 0x0F00) >> 8) // bits 8-11
	y := byte((opcode & 0x00F0) >> 4) //

	// Decode and execute the opcode
	switch opcode & 0xF000 {
	case 0x0000: // 0x00E0: Clear the display
		switch opcode {
		case 0x00E0:
			c.clearScreen()
		case 0x00EE: // Return from subroutine
			return c.ret() // 0x00E0: Clear the display

		default:
			return fmt.Errorf("unknown 0x0000 opcode: 0x%04X", opcode)
		}
	case 0x1000: // 0x1NNN: Jump to address NNN
		c.PC = nnn

	case 0x2000: // 0x2NNN: Call subroutine at NNN
		return c.call(nnn) // 0x2NNN: Call subroutine at NNN

	case 0x3000: // 0x3XNN: Skip next instruction if VX == NN
		if c.V[x] == nn {
			c.PC += 2
		}

	case 0x4000: // 0x4XNN: Skip next instruction if VX != NN
		if c.V[x] != nn {
			c.PC += 2
		}

	case 0x5000:
		if opcode&0x000F != 0 {
			return fmt.Errorf("unknown 0x5000: %04X", opcode)
		}

		// 5XY0 - skip next intruscttions if VX == VY
		if c.V[x] == c.V[y] {
			c.PC += 2
		}

	case 0x6000: // 0x6XNN: Set VX to NN
		c.V[x] = nn

	case 0x7000: // 0x7XNN: Add NN to VX
		c.V[x] += nn

	case 0x8000:
		switch opcode & 0x00F {
		case 0x0:
			c.V[x] = c.V[y]

		case 0x1:
			c.V[x] |= c.V[y]

		case 0x2:
			c.V[x] &= c.V[y]

		case 0x3:
			c.V[x] ^= c.V[y]

		case 0x4:
			sum := uint16(c.V[x]) + uint16(c.V[y])

			if sum > 0xFF {
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}

			c.V[x] = byte(sum)

		case 0x5:
			// 8XY5 - VX = VX - VY, VF = NOT borrow
			vx := c.V[x]
			vy := c.V[y]

			c.V[x] = vx - vy

			if vx >= vy {
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}

		case 0x6:
			// 8XY6 - VX = VY >> 1, VF = dropped least-significant bit
			vy := c.V[y]

			c.V[x] = vy >> 1
			c.V[0xF] = vy & 0x01

		case 0x7:
			// 8XY7 - VX = VY - VX, VF = NOT borrow
			vx := c.V[x]
			vy := c.V[y]

			c.V[x] = vy - vx

			if vy >= vx {
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}

		case 0xE:
			// 8XYE - VX = VY << 1, VF = dropped most-significant bit
			vy := c.V[y]

			c.V[x] = vy << 1
			c.V[0xF] = (vy & 0x80) >> 7

		default:
			return fmt.Errorf("unknown 0x8000 opcode: %04X", opcode)
		}

		return nil

	case 0x9000:
		if opcode&0x000F != 0 {
			return fmt.Errorf("unknown 0x9000 opcode: %04X", opcode)
		}

		// 9XY0 - skip next instruction if VX != VY
		if c.V[x] != c.V[y] {
			c.PC += 2
		}

	case 0xB000:
		// BNNN - jump to V0 + NNN
		c.PC = uint16(c.V[0]) + nnn

	case 0xC000:
		// CXNN - VX = random byte AND NN
		c.V[x] = randomByte() & nn
	case 0xA000: // 0xANNN: Set I to address NNN
		c.I = nnn
	case 0xD000:
		// DXYN - draw N-byte sprite at (VX, VY)
		c.drawSprite(x, y, n)

	case 0xF000:
		switch opcode & 0x00FF {
		case 0x07:
			// FX07 - VX = delay timer read
			c.V[x] = c.DelayTimer

		// stop here until a key is pressed
		case 0x0A:
			// wait for key presses, then store key in VX
			key, pressed := c.pressedKey()
			if !pressed {
				c.PC -= 2
				return nil
			}

		case 0x1E:
			// I = I + VX
			c.I += uint16(c.V[x])
		case 0x15:
			// FX15 - delay timer = VX set
			c.DelayTimer = c.V[x]
		case 0x18:
			// FX18 - sound timer = VX set
			c.SoundTimer = c.V[x]
		case 0x29:
			// FX29 - set I to location of sprite for digit VX
			digit := c.V[x] & 0x0F
			c.I = FontStart + uint16(digit)*5

		case 0x33:
			// store BCD reprsentatoins of VX at I, I+1, I+2
			if err := c.ensureMemoryRange(c.I, 3); err != nil {
				return err
			}

			value := c.V[x]

			c.Memory[c.I] = value / 100
			c.Memory[c.I+1] = (value / 10) % 10
			c.Memory[c.I+2] = value % 10

		case 0x55:
			// store V0 through VX into memory starting at I
			count := int(x) + 1
			if err := c.ensureMemoryRange(c.I, count); err != nil {
				return err
			}

			for i := 0; i < count; i++ {
				c.Memory[c.I+uint16(i)] = c.V[i]
			}

		case 0x65:
			// laod V0 through VX from memory starting at I
			count := int(x) + 1
			if err := c.ensureMemoryRange(c.I, count); err != nil {
				return err
			}

			for i := 0; i < count; i++ {
				c.V[i] = c.Memory[c.I+uint16(i)]
			}

		default:
			return fmt.Errorf("unknown 0xF000 opcode: 0x%04X", opcode)
		}

	case 0xE000:
		switch opcode & 0x00FF {
		case 0x9E:
			// skip next instruction if key stored in VX is pressed
			key := c.V[x]

			if c.isKeyPressed(key) {
				c.PC += 2
			}
		case 0xA1:
			// skip next instruction if key stored in VX is not pressed

			key := c.V[x]

			if !c.isKeyPressed(key) {
				c.PC += 2
			}

		default:
			return fmt.Errorf("unknown 0xE000 opcode: 0x%04X", opcode)
		}
	default:
		return fmt.Errorf("unknown opcode: 0x%04X", opcode)
	}

	return nil
}

func (c *CPU) call(address uint16) error {
	if int(c.SP) >= len(c.Stack) {
		return fmt.Errorf("stack overflow")
	}

	c.Stack[c.SP] = c.PC // Store current PC on stack
	c.SP++               // Increment stack pointer
	c.PC = address       // Jump to subroutine address

	return nil
}

func (c *CPU) ret() error {
	if c.SP == 0 {
		return fmt.Errorf("stack underflow")
	}

	c.SP--               // Decrement stack pointer
	c.PC = c.Stack[c.SP] // Jump to the address stored on the stack
	return nil
}

func (c *CPU) clearScreen() {
	for i := range c.Display {
		c.Display[i] = false
	}

	c.DisplayDirty = true
}

func randomByte() byte {
	var b [1]byte

	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}

	return b[0]
}

func (c *CPU) drawSprite(xReg, yReg, height byte) {
	xPos := int(c.V[xReg]) % screenWidth
	yPos := int(c.V[yReg]) % screenHeight

	c.V[0xF] = 0

	for row := 0; row < int(height); row++ {
		spriteByte := c.Memory[(c.I + uint16(row))] //

		for col := 0; col < 8; col++ {
			if spriteByte == 0 {
				continue
			}

			x := (xPos + col) % screenWidth
			y := (yPos + row) % screenHeight
			index := y*screenWidth + x

			if c.Display[index] {
				c.V[0xF] = 1
			}

			c.Display[index] = !c.Display[index]
		}

		c.DisplayDirty = true

	}
}

func (c *CPU) UpdateTimers() {
	if c.DelayTimer > 0 {
		c.DelayTimer--
	}

	if c.SoundTimer > 0 {
		c.SoundTimer--
	}
}

func (c *CPU) isKeyPressed(key byte) bool {
	if key > 0xF {
		return false
	}

	return c.Keys[key]
}

func (c *CPU) pressedKey() (byte, bool) {
	for key, pressed := range c.Keys {
		if pressed {
			return byte(key), true
		}
	}

	return 0, false
}

func (c *CPU) ensureMemoryRange(start uint16, length int) error {
	if length < 0 {
		return fmt.Errorf("invalid memory range length: %d", length)
	}

	end := int(start) + length

	if end > len(c.Memory) {
		return fmt.Errorf("memory range out of bounds: start=0x%03X length=%d", start, length)
	}

	return nil
}
