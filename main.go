package main

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
)

type chip8 struct {
	memory     [memory]byte
	reg        [regCount]uint8
	instPtr    uint16
	delayTimer uint8
	soundTimer uint8
	pc         uint16
	sp         uint8
	stack      [stackSize]uint16
	display    [displayCol * displayRow]bool
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

func getLast3Nibbles(opcode uint16) uint16 {
	return opcode & 0x0FFF
}

func getLastByte(opcode uint16) uint8 {
	return uint8(opcode & 0x00FF)
}

func cls(chip *chip8, opcode uint16) {
	if opcode == 0x00E0 {
		for i := 0; i < displayCol*displayRow; i++ {
			chip.display[i] = false
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
		// TODO: Continue from here.
		// http://devernay.free.fr/hacks/chip8/C8TECH10.HTM
	}
}

func main() {

}
