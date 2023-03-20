package fcpu

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
	"unsafe"
)

const DataStackTop = 1 << 16
const ReturnStackTop = 1 << 15

type Addr uint32
type Word int32

const WordSize = Addr(unsafe.Sizeof(Word(0)))
const MemMask = int(WordSize - 1)

// BUS     = 0o004
// INVAL   = 0o010
// DEBUG   = 0o014
// IOT     = 0o020
// TTYIN   = 0o060
// TTYOUT  = 0o064
// FAULT   = 0o250
// CLOCK   = 0o100
// RK      = 0o220

const BinaryMagic uint32 = 0xc9f7a115
const MemoryLimit Addr = 0xfffffc00

type BinaryHeader struct {
	Magic    uint32
	TextSize Addr // text size in bytes
	DataSize Addr // initialized data size in bytes
	TextBase Addr // base of text
	DataBase Addr // base of data
}

type CPU struct {
	bus     *Bus   // Bus
	pc      Addr   // Program counter
	Ds      *Stack // Data Stack
	Rs      *Stack // Return Stack
	Verbose bool
	Time    uint64
	Limit   uint64
}

func NewCPU(filename string) (*CPU, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read header
	var header BinaryHeader
	err = binary.Read(file, binary.LittleEndian, &header)
	if err != nil {
		return nil, err
	}
	if header.Magic != BinaryMagic {
		return nil, new(ExecFormatError)
	}

	var cpu *CPU
	cpu = new(CPU)
	cpu.bus = NewBus()
	cpu.pc = header.TextBase
	cpu.Ds = NewStack(cpu.bus, DataStackTop)
	cpu.Rs = NewStack(cpu.bus, ReturnStackTop)

	// Load text segment
	var text = make([]byte, header.TextSize)
	_, err = file.Read(text)
	if err != nil {
		return nil, err
	}
	cpu.bus.WriteBytes(header.TextBase, text)

	// Load data segment
	if header.DataSize != 0 {
		var data = make([]byte, header.DataSize)
		_, err = file.Read(data)
		if err != nil {
			return nil, err
		}
		cpu.bus.WriteBytes(header.DataBase, data)
	}
	return cpu, nil
}

func (cpu *CPU) PrintRegisters() {
	var op Op
	op = Op(cpu.bus.ReadB(cpu.pc))
	// fmt.Printf("pc: %8x  sp: %4x  rsp: %4x  op: %-15s  stack: %s\n",
	// 	cpu.pc, cpu.Ds.pointer, cpu.Rs.pointer, op.String(), cpu.Ds,
	// )
	fmt.Printf("pc: %8x  sp: %8x  rsp: %4x  op: %-15s  stack: %-30.30s  rs: %s\n",
		cpu.pc, cpu.Ds.pointer, cpu.Rs.pointer, op.String(), cpu.Ds, cpu.Rs,
	)
}

func (cpu *CPU) PrintMemory() {
	cpu.bus.Mmu.PrintMemory()
}

func (cpu *CPU) clock() {
	for {
		time.Sleep(1000 * time.Millisecond)
		fmt.Println("X")
		cpu.bus.WriteW(MemoryLimit, Word(0))
	}
}

func (cpu *CPU) Loop() error {
	// go cpu.clock()
	for {
		err := cpu.Eval()
		// time.Sleep(1 * time.Millisecond)
		if err != nil {
			return err
		}
	}
}

func (cpu *CPU) Eval() error {
	var v1 Word
	var v2 Word
	op := Op(cpu.bus.ReadB(cpu.pc))
	if cpu.Verbose {
		cpu.PrintRegisters()
	}
	cpu.Time++
	if cpu.Limit != 0 && cpu.Time >= cpu.Limit {
		return new(Halt)
	}

	// Fetch operands
	if op&POP2 > 0 {
		v1, v2, _ = cpu.Ds.Pop2()
	} else if op&POP1 > 0 {
		v1, _ = cpu.Ds.Pop()
	}

	cpu.pc += OpSize
	switch op {
	case NOP:
		break
	case HLT:
		return new(Halt)
	case PUSH:
		cpu.Ds.Push(cpu.bus.ReadW(cpu.pc))
		cpu.pc += WordSize
	case PUSH_B:
		cpu.Ds.Push(Word(cpu.bus.ReadB(cpu.pc)))
		cpu.pc += 1
	case EMIT: // TODO
		fmt.Printf("%c", int(v1))
	case PERIOD: // TODO
		fmt.Printf(">>>> %d\n", int(v1))
	case DROP: /* Discards the top stack item */
		break
	case DUP: /* Duplicates the top stack item */
		cpu.Ds.Push(v1)
		cpu.Ds.Push(v1)
	case SWAP: /* Reverses the top two stack items */
		cpu.Ds.Push(v2)
		cpu.Ds.Push(v1)
	case OVER: /* Push a copy of the second element on the stack */
		cpu.Ds.Push(v1)
		cpu.Ds.Push(v2)
		cpu.Ds.Push(v1)
	case PICK: /* Remove u. Copy the x-u to the top of the stack. */
		cpu.Ds.Pick(v1)
	case ROLL: /* Remove u.  Rotate u+1 items on the top of the stack */
		cpu.Ds.Roll(v1)
	case DEPTH: /* Count number of items on stack */
		cpu.Ds.Push(Word(cpu.Ds.Size()))
	case TO_R: /* Move top item to the return stack. */
		cpu.Rs.Push(v1)
	case R_FROM: /* Retrieve item from the return stack. */
		v1, _ = cpu.Rs.Pop()
		cpu.Ds.Push(v1)
	case R_FETCH: /* Copy top of return stack onto stack */
		v1, _ = cpu.Rs.Get()
		cpu.Ds.Push(v1)
	case ADD:
		cpu.Ds.Push(v1 + v2)
	case SUB:
		cpu.Ds.Push(v1 - v2)
	case MUL:
		cpu.Ds.Push(v1 * v2)
	case DIV:
		cpu.Ds.Push(v1 / v2)
	case DIVMOD:
		quot := v1 / v2
		rem := v1 % v2
		if rem != 0 && v1*v2 < 0 {
			quot--
		}
		cpu.Ds.Push(rem)
		cpu.Ds.Push(quot)
	case MAX:
		if v1 > v2 {
			cpu.Ds.Push(v1)
		} else {
			cpu.Ds.Push(v2)
		}
	case MIN:
		if v1 < v2 {
			cpu.Ds.Push(v1)
		} else {
			cpu.Ds.Push(v2)
		}
	case ABS:
		if v1 < 0 {
			cpu.Ds.Push(-v1)
		} else {
			cpu.Ds.Push(v1)
		}
	case MOD:
		cpu.Ds.Push(v1 % v2)
	case LSHIFT:
		cpu.Ds.Push(v1 << v2)
	case RSHIFT:
		cpu.Ds.Push(v1 >> v2)
	case AND:
		cpu.Ds.Push(v1 & v2)
	case OR:
		cpu.Ds.Push(v1 | v2)
	case XOR:
		cpu.Ds.Push(v1 ^ v2)
	case NOT:
		cpu.Ds.PushBool(v1 == 0)
	case EQ: /* Compare Equal */
		cpu.Ds.PushBool(v1 == v2)
	case NE: /* Compare for Not Equal */
		cpu.Ds.PushBool(v1 != v2)
	case GE: /* Compare for Greater Or Equal */
		cpu.Ds.PushBool(v1 >= v2)
	case GT: /* Compare for Greater */
		cpu.Ds.PushBool(v1 > v2)
	case LE: /* Compare for Equal or Less */
		cpu.Ds.PushBool(v1 <= v2)
	case LT: /* Compare for Less */
		cpu.Ds.PushBool(v1 < v2)
	case STORE:
		cpu.bus.WriteW(Addr(v2), v1)
	case STORE_B:
		cpu.bus.WriteB(Addr(v2), byte(v1))
	case FETCH:
		value := cpu.bus.ReadW(Addr(v1))
		// fmt.Println("FETCH: ---", int(v1), int(value))
		cpu.Ds.Push(value)
	case FETCH_B:
		value := Word(cpu.bus.ReadB(Addr(v1)))
		// fmt.Println("FETCH_B: ---", int(v1), int(value))
		cpu.Ds.Push(value)
	case JNZ: // jump if not zero
		// fmt.Println("JNZ: ---", int(v1), int(v2))
		if v1 != 0 {
			cpu.pc = Addr(v2)
		}
	case JZ: // jump if zero
		// fmt.Println("JZ: ---", int(v1), int(v2))
		if v1 == 0 {
			cpu.pc = Addr(v2)
		}
	case JMP:
		cpu.pc = Addr(v1)
	case PUSHRSP:
		cpu.Ds.Push(Word(cpu.Rs.pointer))
	case POPRSP:
		cpu.Rs.pointer = Addr(v1)
	case PUSHRBP:
		cpu.Ds.Push(Word(cpu.Rs.origin))
	case POPRBP:
		cpu.Rs.origin = Addr(v1)
	case PUSHPC:
		cpu.Ds.Push(Word(cpu.pc))
	case CALL:
		cpu.Rs.Push(Word(cpu.pc))
		// cpu.bus.WriteW(cpu.rsp, Word(cpu.rsp))            // store rsp
		// cpu.bus.WriteW(cpu.rsp+1*WordSize, Word(cpu.rbp)) // store rbp
		// cpu.bus.WriteW(cpu.rsp+2*WordSize, Word(cpu.pc))  // store pc
		// cpu.rbp = cpu.rsp + 3*WordSize
		// cpu.rsp = cpu.rbp
		cpu.pc = Addr(v1)
	case RET:
		// cpu.pc = Addr(cpu.bus.ReadW(cpu.rsp - 1*WordSize))  // return
		// cpu.rbp = Addr(cpu.bus.ReadW(cpu.rsp - 2*WordSize)) // restore rbp
		// cpu.rsp = Addr(cpu.bus.ReadW(cpu.rsp - 3*WordSize)) // restore rsp
		v1, _ = cpu.Rs.Pop()
		cpu.pc = Addr(v1)
	}
	return nil
}

// int is_transmit_empty() {
//    return inb(PORT + 5) & 0x20;
// }
//
// void write_serial(char a) {
//    while (is_transmit_empty() == 0);
//
//    outb(PORT,a);
// }
