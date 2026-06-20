// internal/chip8/cpu_test.go
package chip8

import "testing"

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

	if _, err := cpu.Step(); err != nil {
		t.Fatal(err)
	}

	if cpu.V[0] != 10 {
		t.Fatalf("expected V0 to be 10, got %d", cpu.V[0])
	}

	if _, err := cpu.Step(); err != nil {
		t.Fatal(err)
	}

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

	if _, err := cpu.Step(); err != nil {
		t.Fatal(err)
	}

	if cpu.PC != 0x206 {
		t.Fatalf("expected PC to be 0x206, got 0x%03X", cpu.PC)
	}

	if _, err := cpu.Step(); err != nil {
		t.Fatal(err)
	}

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

	if _, err := cpu.Step(); err != nil {
		t.Fatal(err)
	}

	if cpu.PC != 0x206 {
		t.Fatalf("expected PC to be 0x206 after CALL, got 0x%03X", cpu.PC)
	}

	if cpu.SP != 1 {
		t.Fatalf("expected SP to be 1 after CALL, got %d", cpu.SP)
	}

	if cpu.Stack[0] != 0x202 {
		t.Fatalf("expected return address 0x202, got 0x%03X", cpu.Stack[0])
	}

	if _, err := cpu.Step(); err != nil {
		t.Fatal(err)
	}

	if cpu.V[1] != 2 {
		t.Fatalf("expected V1 to be 2, got %d", cpu.V[1])
	}

	if _, err := cpu.Step(); err != nil {
		t.Fatal(err)
	}

	if cpu.PC != 0x202 {
		t.Fatalf("expected PC to return to 0x202, got 0x%03X", cpu.PC)
	}

	if cpu.SP != 0 {
		t.Fatalf("expected SP to be 0 after return, got %d", cpu.SP)
	}

	if _, err := cpu.Step(); err != nil {
		t.Fatal(err)
	}

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

	if _, err := cpu.Step(); err != nil {
		t.Fatal(err)
	}

	for i, pixel := range cpu.Display {
		if pixel {
			t.Fatalf("expected display[%d] to be false", i)
		}
	}

	if !cpu.DisplayDirty {
		t.Fatal("expected DisplayDirty to be true after clearing screen")
	}
}
