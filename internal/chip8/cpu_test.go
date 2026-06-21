// internal/chip8/cpu_test.go
package chip8

import "testing"

func stepN(t *testing.T, cpu *CPU, n int) {
	t.Helper()

	for i := 0; i < n; i++ {
		if _, err := cpu.Step(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLoadProgram(t *testing.T) {
	cpu := New()

	program := []byte{0x60, 0x0A, 0x61, 0x05}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	if cpu.Memory[0x200] != 0x60 {
		t.Fatalf("expected memory[0x200] to be 0x60, got 0x%02X", cpu.Memory[0x200])
	}

	if cpu.Memory[0x201] != 0x0A {
		t.Fatalf("expected memory[0x201] to be 0x0A, got 0x%02X", cpu.Memory[0x201])
	}

	if cpu.PC != 0x200 {
		t.Fatalf("expected PC to start at 0x200, got 0x%03X", cpu.PC)
	}
}

func TestSetRegisterAndAdd(t *testing.T) {
	cpu := New()

	program := []byte{
		0x60, 0x0A, // V0 = 10
		0x70, 0x01, // V0 += 1
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 2)

	if cpu.V[0] != 11 {
		t.Fatalf("expected V0 to be 11, got %d", cpu.V[0])
	}
}

func TestJump(t *testing.T) {
	cpu := New()

	program := []byte{
		0x12, 0x06, // jump to 0x206
		0x60, 0xFF, // skipped
		0x60, 0xEE, // skipped
		0x61, 0x0A, // V1 = 10
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 1)

	if cpu.PC != 0x206 {
		t.Fatalf("expected PC to be 0x206, got 0x%03X", cpu.PC)
	}

	stepN(t, cpu, 1)

	if cpu.V[1] != 10 {
		t.Fatalf("expected V1 to be 10, got %d", cpu.V[1])
	}
}

func TestCallAndReturn(t *testing.T) {
	cpu := New()

	program := []byte{
		0x22, 0x06, // 0x200: call 0x206
		0x60, 0x01, // 0x202: V0 = 1
		0x00, 0x00, // 0x204: unused
		0x61, 0x02, // 0x206: V1 = 2
		0x00, 0xEE, // 0x208: return
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 1)

	if cpu.PC != 0x206 {
		t.Fatalf("expected PC to be 0x206 after CALL, got 0x%03X", cpu.PC)
	}

	if cpu.SP != 1 {
		t.Fatalf("expected SP to be 1 after CALL, got %d", cpu.SP)
	}

	if cpu.Stack[0] != 0x202 {
		t.Fatalf("expected return address 0x202, got 0x%03X", cpu.Stack[0])
	}

	stepN(t, cpu, 1)

	if cpu.V[1] != 2 {
		t.Fatalf("expected V1 to be 2, got %d", cpu.V[1])
	}

	stepN(t, cpu, 1)

	if cpu.PC != 0x202 {
		t.Fatalf("expected PC to return to 0x202, got 0x%03X", cpu.PC)
	}

	if cpu.SP != 0 {
		t.Fatalf("expected SP to be 0 after return, got %d", cpu.SP)
	}

	stepN(t, cpu, 1)

	if cpu.V[0] != 1 {
		t.Fatalf("expected V0 to be 1, got %d", cpu.V[0])
	}
}

func TestClearScreen(t *testing.T) {
	cpu := New()

	cpu.Display[0] = true
	cpu.Display[10] = true
	cpu.DisplayDirty = false

	program := []byte{
		0x00, 0xE0, // clear screen
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 1)

	for i, pixel := range cpu.Display {
		if pixel {
			t.Fatalf("expected display[%d] to be false", i)
		}
	}

	if !cpu.DisplayDirty {
		t.Fatal("expected DisplayDirty to be true after clearing screen")
	}
}

func Test3XNNSkipIfVXEqualsNN(t *testing.T) {
	cpu := New()

	program := []byte{
		0x60, 0x0A, // V0 = 10
		0x30, 0x0A, // skip next if V0 == 10
		0x61, 0xFF, // skipped
		0x61, 0x05, // V1 = 5
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 3)

	if cpu.V[1] != 5 {
		t.Fatalf("expected V1 to be 5, got %d", cpu.V[1])
	}
}

func Test4XNNSkipIfVXNotEqualsNN(t *testing.T) {
	cpu := New()

	program := []byte{
		0x60, 0x0A, // V0 = 10
		0x40, 0x05, // skip next if V0 != 5
		0x61, 0xFF, // skipped
		0x61, 0x05, // V1 = 5
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 3)

	if cpu.V[1] != 5 {
		t.Fatalf("expected V1 to be 5, got %d", cpu.V[1])
	}
}

func Test5XY0SkipIfVXEqualsVY(t *testing.T) {
	cpu := New()

	program := []byte{
		0x60, 0x0A, // V0 = 10
		0x61, 0x0A, // V1 = 10
		0x50, 0x10, // skip next if V0 == V1
		0x62, 0xFF, // skipped
		0x62, 0x05, // V2 = 5
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 4)

	if cpu.V[2] != 5 {
		t.Fatalf("expected V2 to be 5, got %d", cpu.V[2])
	}
}

func Test8XY0SetVXToVY(t *testing.T) {
	cpu := New()

	program := []byte{
		0x61, 0x0A, // V1 = 10
		0x80, 0x10, // V0 = V1
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 2)

	if cpu.V[0] != 10 {
		t.Fatalf("expected V0 to be 10, got %d", cpu.V[0])
	}
}

func Test8XY1Or(t *testing.T) {
	cpu := New()

	program := []byte{
		0x60, 0b1010, // V0 = 10
		0x61, 0b1100, // V1 = 12
		0x80, 0x11, // V0 = V0 OR V1
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 3)

	if cpu.V[0] != 0b1110 {
		t.Fatalf("expected V0 to be 14, got %d", cpu.V[0])
	}
}

func Test8XY2And(t *testing.T) {
	cpu := New()

	program := []byte{
		0x60, 0b1010, // V0 = 10
		0x61, 0b1100, // V1 = 12
		0x80, 0x12, // V0 = V0 AND V1
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 3)

	if cpu.V[0] != 0b1000 {
		t.Fatalf("expected V0 to be 8, got %d", cpu.V[0])
	}
}

func Test8XY3Xor(t *testing.T) {
	cpu := New()

	program := []byte{
		0x60, 0b1010, // V0 = 10
		0x61, 0b1100, // V1 = 12
		0x80, 0x13, // V0 = V0 XOR V1
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 3)

	if cpu.V[0] != 0b0110 {
		t.Fatalf("expected V0 to be 6, got %d", cpu.V[0])
	}
}

func Test8XY4AddNoCarry(t *testing.T) {
	cpu := New()

	program := []byte{
		0x60, 10, // V0 = 10
		0x61, 20, // V1 = 20
		0x80, 0x14, // V0 = V0 + V1
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 3)

	if cpu.V[0] != 30 {
		t.Fatalf("expected V0 to be 30, got %d", cpu.V[0])
	}

	if cpu.V[0xF] != 0 {
		t.Fatalf("expected VF to be 0, got %d", cpu.V[0xF])
	}
}

func Test8XY4AddWithCarry(t *testing.T) {
	cpu := New()

	program := []byte{
		0x60, 250, // V0 = 250
		0x61, 10, // V1 = 10
		0x80, 0x14, // V0 = V0 + V1
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 3)

	if cpu.V[0] != 4 {
		t.Fatalf("expected V0 to wrap to 4, got %d", cpu.V[0])
	}

	if cpu.V[0xF] != 1 {
		t.Fatalf("expected VF to be 1, got %d", cpu.V[0xF])
	}
}

func Test8XY5SubtractNoBorrow(t *testing.T) {
	cpu := New()

	program := []byte{
		0x60, 10, // V0 = 10
		0x61, 3, // V1 = 3
		0x80, 0x15, // V0 = V0 - V1
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 3)

	if cpu.V[0] != 7 {
		t.Fatalf("expected V0 to be 7, got %d", cpu.V[0])
	}

	if cpu.V[0xF] != 1 {
		t.Fatalf("expected VF to be 1 because no borrow happened, got %d", cpu.V[0xF])
	}
}

func Test8XY5SubtractWithBorrow(t *testing.T) {
	cpu := New()

	program := []byte{
		0x60, 3, // V0 = 3
		0x61, 10, // V1 = 10
		0x80, 0x15, // V0 = V0 - V1
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 3)

	if cpu.V[0] != 249 {
		t.Fatalf("expected V0 to wrap to 249, got %d", cpu.V[0])
	}

	if cpu.V[0xF] != 0 {
		t.Fatalf("expected VF to be 0 because borrow happened, got %d", cpu.V[0xF])
	}
}

func Test8XY6ShiftRight(t *testing.T) {
	cpu := New()

	program := []byte{
		0x61, 0b00000101, // V1 = 5
		0x80, 0x16, // V0 = V1 >> 1, VF = dropped LSB
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 2)

	if cpu.V[0] != 0b00000010 {
		t.Fatalf("expected V0 to be 2, got %d", cpu.V[0])
	}

	if cpu.V[0xF] != 1 {
		t.Fatalf("expected VF to be 1, got %d", cpu.V[0xF])
	}
}

func Test8XY7ReverseSubtract(t *testing.T) {
	cpu := New()

	program := []byte{
		0x60, 3, // V0 = 3
		0x61, 10, // V1 = 10
		0x80, 0x17, // V0 = V1 - V0
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 3)

	if cpu.V[0] != 7 {
		t.Fatalf("expected V0 to be 7, got %d", cpu.V[0])
	}

	if cpu.V[0xF] != 1 {
		t.Fatalf("expected VF to be 1 because no borrow happened, got %d", cpu.V[0xF])
	}
}

func Test8XYEShiftLeft(t *testing.T) {
	cpu := New()

	program := []byte{
		0x61, 0b10000001, // V1 = 129
		0x80, 0x1E, // V0 = V1 << 1, VF = dropped MSB
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 2)

	if cpu.V[0] != 0b00000010 {
		t.Fatalf("expected V0 to be 2, got %d", cpu.V[0])
	}

	if cpu.V[0xF] != 1 {
		t.Fatalf("expected VF to be 1, got %d", cpu.V[0xF])
	}
}

func Test9XY0SkipIfVXNotEqualsVY(t *testing.T) {
	cpu := New()

	program := []byte{
		0x60, 1, // V0 = 1
		0x61, 2, // V1 = 2
		0x90, 0x10, // skip next if V0 != V1
		0x62, 99, // skipped
		0x62, 42, // V2 = 42
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 4)

	if cpu.V[2] != 42 {
		t.Fatalf("expected V2 to be 42, got %d", cpu.V[2])
	}
}

func TestANNNSetIndexRegister(t *testing.T) {
	cpu := New()

	program := []byte{
		0xA3, 0x00, // I = 0x300
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 1)

	if cpu.I != 0x300 {
		t.Fatalf("expected I to be 0x300, got 0x%03X", cpu.I)
	}
}

func TestBNNNJumpToV0PlusNNN(t *testing.T) {
	cpu := New()

	program := []byte{
		0x60, 0x04, // V0 = 4
		0xB2, 0x08, // PC = 0x208 + V0 = 0x20C
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 2)

	if cpu.PC != 0x20C {
		t.Fatalf("expected PC to be 0x20C, got 0x%03X", cpu.PC)
	}
}

func TestCXNNRandomAndMask(t *testing.T) {
	cpu := New()

	program := []byte{
		0xC0, 0x0F, // V0 = random byte AND 0x0F
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	stepN(t, cpu, 1)

	if cpu.V[0] > 0x0F {
		t.Fatalf("expected V0 to be masked to <= 0x0F, got 0x%02X", cpu.V[0])
	}
}

// internal/chip8/cpu_test.go

func TestDXYNDrawSinglePixelSprite(t *testing.T) {
	cpu := New()

	program := []byte{
		0xA3, 0x00, // I = 0x300
		0x60, 10, // V0 = x = 10
		0x61, 5, // V1 = y = 5
		0xD0, 0x11, // draw sprite at (V0, V1), height 1
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	cpu.Memory[0x300] = 0b10000000

	stepN(t, cpu, 4)

	index := 5*screenWidth + 10

	if !cpu.Display[index] {
		t.Fatalf("expected pixel at (10, 5) to be on")
	}

	if cpu.V[0xF] != 0 {
		t.Fatalf("expected VF to be 0 when no collision happens, got %d", cpu.V[0xF])
	}

	if !cpu.DisplayDirty {
		t.Fatal("expected DisplayDirty to be true after drawing")
	}
}

func TestDXYNDrawFullByteSprite(t *testing.T) {
	cpu := New()

	program := []byte{
		0xA3, 0x00, // I = 0x300
		0x60, 4, // V0 = x = 4
		0x61, 3, // V1 = y = 3
		0xD0, 0x11, // draw sprite at (V0, V1), height 1
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	cpu.Memory[0x300] = 0b11110000

	stepN(t, cpu, 4)

	onPixels := []int{
		3*screenWidth + 4,
		3*screenWidth + 5,
		3*screenWidth + 6,
		3*screenWidth + 7,
	}

	offPixels := []int{
		3*screenWidth + 8,
		3*screenWidth + 9,
		3*screenWidth + 10,
		3*screenWidth + 11,
	}

	for _, index := range onPixels {
		if !cpu.Display[index] {
			t.Fatalf("expected display[%d] to be on", index)
		}
	}

	for _, index := range offPixels {
		if cpu.Display[index] {
			t.Fatalf("expected display[%d] to be off", index)
		}
	}
}

func TestDXYNDrawMultiRowSprite(t *testing.T) {
	cpu := New()

	program := []byte{
		0xA3, 0x00, // I = 0x300
		0x60, 2, // V0 = x = 2
		0x61, 4, // V1 = y = 4
		0xD0, 0x13, // draw sprite at (V0, V1), height 3
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	cpu.Memory[0x300] = 0b10000000
	cpu.Memory[0x301] = 0b01000000
	cpu.Memory[0x302] = 0b00100000

	stepN(t, cpu, 4)

	expectedPixels := []int{
		4*screenWidth + 2,
		5*screenWidth + 3,
		6*screenWidth + 4,
	}

	for _, index := range expectedPixels {
		if !cpu.Display[index] {
			t.Fatalf("expected display[%d] to be on", index)
		}
	}

	if cpu.V[0xF] != 0 {
		t.Fatalf("expected VF to be 0 when no collision happens, got %d", cpu.V[0xF])
	}
}

func TestDXYNDrawSpriteCollision(t *testing.T) {
	cpu := New()

	program := []byte{
		0xA3, 0x00, // I = 0x300
		0x60, 10, // V0 = x = 10
		0x61, 5, // V1 = y = 5
		0xD0, 0x11, // draw once
		0xD0, 0x11, // draw same sprite again
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	cpu.Memory[0x300] = 0b10000000

	stepN(t, cpu, 4)

	index := 5*screenWidth + 10

	if !cpu.Display[index] {
		t.Fatalf("expected pixel to be on after first draw")
	}

	if cpu.V[0xF] != 0 {
		t.Fatalf("expected VF to be 0 after first draw, got %d", cpu.V[0xF])
	}

	stepN(t, cpu, 1)

	if cpu.Display[index] {
		t.Fatalf("expected pixel to be off after second draw")
	}

	if cpu.V[0xF] != 1 {
		t.Fatalf("expected VF to be 1 after collision, got %d", cpu.V[0xF])
	}
}

func TestDXYNDrawSpriteWrapsAroundScreen(t *testing.T) {
	cpu := New()

	program := []byte{
		0xA3, 0x00, // I = 0x300
		0x60, 63, // V0 = x = 63
		0x61, 31, // V1 = y = 31
		0xD0, 0x12, // draw sprite at bottom-right, height 2
	}

	if err := cpu.LoadProgram(program); err != nil {
		t.Fatal(err)
	}

	cpu.Memory[0x300] = 0b11000000
	cpu.Memory[0x301] = 0b10000000

	stepN(t, cpu, 4)

	expectedPixels := []int{
		31*screenWidth + 63, // first row, x = 63
		31*screenWidth + 0,  // first row wraps to x = 0
		0*screenWidth + 63,  // second row wraps to y = 0
	}

	for _, index := range expectedPixels {
		if !cpu.Display[index] {
			t.Fatalf("expected display[%d] to be on", index)
		}
	}
}
