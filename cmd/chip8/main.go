// cmd/chip8/main.go
package main

import (
	"image/color"
	"log"
	"os"

	"github.com/JustTryingToDoBetter/go-chip8/internal/chip8"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	windowScale   = 12
	cyclesPerTick = 10
)

type Game struct {
	cpu *chip8.CPU
}

func NewGame(romPath string) (*Game, error) {
	cpu := chip8.New()
	if err := cpu.LoadROM(romPath); err != nil {
		return nil, err
	}

	return &Game{
		cpu: cpu,
	}, nil
}

func (g *Game) Update() error {
	g.syncInput()

	for i := 0; i < cyclesPerTick; i++ {
		if _, err := g.cpu.Step(); err != nil {
			return err
		}
	}

	g.cpu.UpdateTimers()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	for y := 0; y < chip8.ScreenHeight; y++ {
		for x := 0; x < chip8.ScreenWidth; x++ {
			index := y*chip8.ScreenWidth + x

			if g.cpu.Display[index] {
				screen.Set(x, y, color.White)
			}
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return chip8.ScreenWidth, chip8.ScreenHeight
}

func (g *Game) syncInput() {
	for i := range g.cpu.Keys {
		g.cpu.Keys[i] = false
	}

	keyMap := map[ebiten.Key]byte{
		ebiten.KeyX: 0x0,
		ebiten.Key1: 0x1,
		ebiten.Key2: 0x2,
		ebiten.Key3: 0x3,
		ebiten.KeyQ: 0x4,
		ebiten.KeyW: 0x5,
		ebiten.KeyE: 0x6,
		ebiten.KeyA: 0x7,
		ebiten.KeyS: 0x8,
		ebiten.KeyD: 0x9,
		ebiten.KeyZ: 0xA,
		ebiten.KeyC: 0xB,
		ebiten.Key4: 0xC,
		ebiten.KeyR: 0xD,
		ebiten.KeyF: 0xE,
		ebiten.KeyV: 0xF,
	}

	for physicalKey, chip8Key := range keyMap {
		if ebiten.IsKeyPressed(physicalKey) {
			g.cpu.Keys[chip8Key] = true
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: chip8 <rom-path>")
	}

	game, err := NewGame(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowSize(
		chip8.ScreenWidth*windowScale,
		chip8.ScreenHeight*windowScale,
	)
	ebiten.SetWindowTitle("go-chip8")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
