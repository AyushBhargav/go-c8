package main

import (
	"math/rand"
	"time"
)

const (
	memory         = 4096
	programLoc     = 0x200
	regCount       = 0xF + 1
	timerFrequence = 60 //Hz
	stackSize      = 16
	displayRow     = 32
	displayCol     = 64
	nibble         = 4
	byt            = 8
	bytNib         = 12
	carry          = 0xF
	fontNum        = 0xF
)

type chip8 struct {
	memory     [memory]byte
	fontMemory [fontNum]uint16
	reg        [regCount]uint8
	instPtr    uint16
	delayTimer uint8
	soundTimer uint8
	pc         uint16
	sp         uint8
	stack      [stackSize]uint16
	display    [displayRow][displayCol]uint8
	keyboard   [regCount]bool
}

type operation interface {
	do(chip *chip8, opcode uint16)
}

func getNibble(opcode uint16, place uint8) uint8 {
	if place < 1 || place > 4 {
		return 0
	}
	for i := uint8(1); i < place; i++ {
		opcode = opcode << 4
	}
	opcode = (opcode & 0xF000) >> 12
	return uint8(opcode)
}

func getLast2Nibbles(opcode uint16) uint8 {
	return uint8(opcode & 0x00FF)
}

func getLast3Nibbles(opcode uint16) uint16 {
	return opcode & 0x0FFF
}

func getLastByte(opcode uint16) uint8 {
	return uint8(opcode & 0x00FF)
}

func cls(chip *chip8, opcode uint16) {
	if opcode == 0x00E0 {
		for i := 0; i < displayRow; i++ {
			for j := 0; i < displayCol; j++ {
				chip.display[i][j] = 0
			}
		}
	}
}

func jpAddr(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x1 {
		chip.pc = getLast3Nibbles(opcode)
	}
}

func callAddr(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x2 {
		chip.sp++
		if chip.sp == stackSize {
			panic("Stack overflown")
		}
		chip.stack[chip.sp] = chip.pc
		chip.pc = getLast3Nibbles(opcode)
	}
}

func skipInst(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x3 {
		if chip.reg[getLastByte(opcode)] == getNibble(opcode, 2) {
			chip.pc += 2
		}
	}
}

func skipNotInst(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x4 {
		if chip.reg[getLastByte(opcode)] != getNibble(opcode, 2) {
			chip.pc += 2
		}
	}
}

func skipInstVxVy(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x5 && getNibble(opcode, 4) == 0 {
		if chip.reg[getNibble(opcode, 2)] == chip.reg[getNibble(opcode, 3)] {
			chip.pc += 2
		}
	}
}

func loadVx(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x6 {
		chip.reg[getNibble(opcode, 2)] = getLastByte(opcode)
	}
}

func addVx(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x7 {
		chip.reg[getNibble(opcode, 2)] += getLastByte(opcode)
	}
}

func loadVxVy(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x8 && getNibble(opcode, 4) == 0 {
		chip.reg[getNibble(opcode, 2)] = chip.reg[getNibble(opcode, 3)]
	}
}

func orVxVy(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x8 && getNibble(opcode, 4) == 0x1 {
		chip.reg[getNibble(opcode, 2)] |= chip.reg[getNibble(opcode, 3)]
	}
}

func andVxVy(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x8 && getNibble(opcode, 4) == 0x2 {
		chip.reg[getNibble(opcode, 2)] &= chip.reg[getNibble(opcode, 3)]
	}
}

func xorVxVy(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x8 && getNibble(opcode, 4) == 0x3 {
		chip.reg[getNibble(opcode, 2)] ^= chip.reg[getNibble(opcode, 3)]
	}
}

func addVxVy(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x8 && getNibble(opcode, 4) == 0x4 {
		temp := uint16(chip.reg[getNibble(opcode, 2)]) + uint16(chip.reg[getNibble(opcode, 3)])
		if temp > 0xFF {
			chip.reg[0xF] = 1
		}
		chip.reg[getNibble(opcode, 2)] = uint8(temp)
	}
}

func subVxVy(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x8 && getNibble(opcode, 4) == 0x5 {
		temp := uint16(chip.reg[getNibble(opcode, 2)]) - uint16(chip.reg[getNibble(opcode, 3)])
		if temp < 0 {
			chip.reg[0xF] = 1
		}
		chip.reg[getNibble(opcode, 2)] = uint8(temp)
	}
}

func shrVx(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x8 && getNibble(opcode, 4) == 0x6 {
		chip.reg[carry] = chip.reg[getNibble(opcode, 2)] & 0x1
		chip.reg[getNibble(opcode, 2)] /= 2
	}
}

func subNotVxVy(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x8 && getNibble(opcode, 4) == 0x7 {
		vx := getNibble(opcode, 2)
		vy := getNibble(opcode, 3)
		if chip.reg[vx] > chip.reg[vy] {
			chip.reg[carry] = 1
		} else {
			chip.reg[carry] = 0
		}
		chip.reg[vx] = chip.reg[vy] - chip.reg[vx]
	}
}

func shiftLeftVx(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x8 && getNibble(opcode, 4) == 0xE {
		vx := getNibble(opcode, 2)
		if chip.reg[vx]&1<<7 != 0 {
			chip.reg[carry] = 1
		} else {
			chip.reg[carry] = 0
		}
		chip.reg[vx] = 2 * chip.reg[vx]
	}
}

func skipIfSameVxVy(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0x9 && getNibble(opcode, 4) == 0x0 {
		vx := getNibble(opcode, 2)
		vy := getNibble(opcode, 3)
		if chip.reg[vx] == chip.reg[vy] {
			chip.pc++
		}
	}
}

func setI(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0xA {
		chip.instPtr = getLast3Nibbles(opcode)
	}
}

func jumpI(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0xB {
		chip.instPtr = uint16(chip.reg[0]) + getLast3Nibbles(opcode)
	}
}

func randomVx(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0xC {
		vx := getNibble(opcode, 2)
		k := getLast2Nibbles(opcode)
		rand.Seed(time.Now().UnixNano())
		chip.reg[vx] = uint8(rand.Intn(256)) & k
	}
}

func drawVxVy(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0xD {
		vx := chip.reg[getNibble(opcode, 2)]
		vy := chip.reg[getNibble(opcode, 3)]
		n := getNibble(opcode, 4)
		pixelErased := false
		for i := uint8(0); i < n; i++ {
			b := chip.memory[chip.instPtr+uint16(i)]
			for j := uint8(7); j >= 0; j-- {
				x := (vx + j) % displayCol
				y := (vy + i) % displayRow
				bit := b & (0x1 << j)
				oldPixel := chip.display[y][x]
				chip.display[y][x] = chip.display[y][x] ^ bit
				pixelErased = pixelErased || (oldPixel == 1 && chip.display[y][x] == 0)
			}
		}
		if pixelErased {
			chip.reg[carry] = 1
		} else {
			chip.reg[carry] = 0
		}
	}
}

func skipIfPressedVx(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0xE && getNibble(opcode, 3) == 0x9 &&
		getNibble(opcode, 4) == 0xE {
		if chip.keyboard[chip.reg[getNibble(opcode, 2)]] {
			chip.pc++
		}
	}
}

func skipIfNotPressedVx(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0xE && getNibble(opcode, 3) == 0xA &&
		getNibble(opcode, 4) == 0x1 {
		if !chip.keyboard[chip.reg[getNibble(opcode, 2)]] {
			chip.pc++
		}
	}
}

func setDelayTimerToVx(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0xF && getNibble(opcode, 3) == 0x0 &&
		getNibble(opcode, 4) == 0x7 {
		chip.reg[getNibble(opcode, 2)] = chip.delayTimer
	}
}

func setInputToVx(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0xF && getNibble(opcode, 3) == 0x0 &&
		getNibble(opcode, 4) == 0xA {
		vx := getNibble(opcode, 2)
		for isKeyPressed := false; !isKeyPressed; {
			for key, keyPress := range chip.keyboard {
				if keyPress {
					chip.reg[vx] = uint8(key)
					isKeyPressed = true
					break
				}
			}
		}
	}
}

func setDelayTimerVx(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0xF && getNibble(opcode, 3) == 0x1 &&
		getNibble(opcode, 4) == 0x5 {
		chip.delayTimer = chip.reg[getNibble(opcode, 2)]
	}
}

func setSoundTimerVx(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0xF && getNibble(opcode, 3) == 0x1 &&
		getNibble(opcode, 4) == 0x8 {
		chip.soundTimer = chip.reg[getNibble(opcode, 2)]
	}
}

func setInstrPointerFontLocationVx(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0xF && getNibble(opcode, 3) == 0x2 &&
		getNibble(opcode, 4) == 0x9 {
		vx := chip.reg[getNibble(opcode, 2)]
		chip.instPtr = chip.fontMemory[vx]
	}
}

func setInstrPointerBCDVx(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0xF && getNibble(opcode, 3) == 0x3 &&
		getNibble(opcode, 4) == 0x3 {
		vx := chip.reg[getNibble(opcode, 2)]
		for i := uint16(2); i >= 0; i-- {
			chip.memory[chip.instPtr+i] = vx % 10
			vx = vx / 10
		}
	}
}

func setRegistersInMemoryVx(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0xF && getNibble(opcode, 3) == 0x5 &&
		getNibble(opcode, 4) == 0x5 {
		vx := chip.reg[getNibble(opcode, 2)]
		for i := uint8(0); i <= vx; i++ {
			chip.memory[chip.instPtr+uint16(i)] = chip.reg[i]
		}
	}
}

func loadRegistersFromMemoryVx(chip *chip8, opcode uint16) {
	if getNibble(opcode, 1) == 0xF && getNibble(opcode, 3) == 0x6 &&
		getNibble(opcode, 4) == 0x5 {
		vx := chip.reg[getNibble(opcode, 2)]
		for i := uint8(0); i <= vx; i++ {
			chip.reg[i] = chip.memory[chip.instPtr+uint16(i)]
		}
	}
}

func main() {
}
