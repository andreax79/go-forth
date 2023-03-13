package fcpu

import (
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"
)

const DataStackTop = 1 << 16
const ReturnStackTop = 1 << 15

type Addr uint32
type Word int32

const WordSize = Addr(unsafe.Sizeof(Word(0)))

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

type BinaryHeader struct {
	Magic    uint32
	TextSize Addr // text size in bytes
	DataSize Addr // initialized data size in bytes
	TextBase Addr // base of text
	DataBase Addr // base of data
}

type CPU struct {
	mmu     *MMU   // Memory Management Unit
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
	cpu.mmu = NewMMU()
	cpu.pc = header.TextBase
	cpu.Ds = NewStack(cpu.mmu, DataStackTop)
	cpu.Rs = NewStack(cpu.mmu, ReturnStackTop)

	// Load text segment
	var text = make([]byte, header.TextSize)
	_, err = file.Read(text)
	if err != nil {
		return nil, err
	}
	cpu.mmu.WriteBytes(header.TextBase, text)

	// Load data segment
	if header.DataSize != 0 {
		var data = make([]byte, header.DataSize)
		_, err = file.Read(data)
		if err != nil {
			return nil, err
		}
		cpu.mmu.WriteBytes(header.DataBase, data)
	}
	return cpu, nil
}

func (cpu *CPU) PrintRegisters() {
	var op Op
	op = Op(cpu.mmu.ReadB(cpu.pc))
	// fmt.Printf("pc: %8x  sp: %4x  rsp: %4x  op: %-15s  stack: %s\n",
	// 	cpu.pc, cpu.Ds.pointer, cpu.Rs.pointer, op.String(), cpu.Ds,
	// )
	fmt.Printf("pc: %8x  sp: %8x  rsp: %4x  op: %-15s  stack: %-30.30s  rs: %s\n",
		cpu.pc, cpu.Ds.pointer, cpu.Rs.pointer, op.String(), cpu.Ds, cpu.Rs,
	)
}

func (cpu *CPU) PrintMemory() {
	cpu.mmu.PrintMemory()
}

func (cpu *CPU) Eval() error {
	var v1 Word
	var v2 Word
	var v3 Word
	op := Op(cpu.mmu.ReadB(cpu.pc))
	if cpu.Verbose {
		cpu.PrintRegisters()
	}
	cpu.Time++
	if cpu.Limit != 0 && cpu.Time >= cpu.Limit {
		return new(Halt)
	}

	cpu.pc += OpSize
	switch op {
	case NOP:
		break
	case HLT:
		return new(Halt)
	case PUSH:
		cpu.Ds.Push(cpu.mmu.ReadW(cpu.pc))
		cpu.pc += WordSize
	case PUSH_B:
		cpu.Ds.Push(Word(cpu.mmu.ReadB(cpu.pc)))
		cpu.pc += 1
	case EMIT: // TODO
		v1, _ = cpu.Ds.Pop()
		fmt.Printf("%c", int(v1))
	case PERIOD: // TODO
		v1, _ = cpu.Ds.Pop()
		fmt.Printf(">>>> %d\n", int(v1))
	case ZERO:
		cpu.Ds.Push(0)
	case DROP: /* Discards the top stack item */
		cpu.Ds.Pop()
	case DUP: /* Duplicates the top stack item */
		cpu.Ds.Dup()
	case SWAP: /* Reverses the top two stack items */
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v2)
		cpu.Ds.Push(v1)
	case OVER: /* Push a copy of the second element on the stack */
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1)
		cpu.Ds.Push(v2)
		cpu.Ds.Push(v1)
	case ROT: /* Rotate the third item to top */
		v1, v2, _ = cpu.Ds.Pop2()
		v3, _ = cpu.Ds.Pop()
		cpu.Ds.Push(v1)
		cpu.Ds.Push(v2)
		cpu.Ds.Push(v3)
	case PICK: /* Remove u. Copy the x-u to the top of the stack. */
		v1, _ = cpu.Ds.Pop()
		cpu.Ds.Pick(v1)
	case ROLL: /* Remove u.  Rotate u+1 items on the top of the stack */
		v1, _ = cpu.Ds.Pop()
		cpu.Ds.Roll(v1)
	case DEPTH: /* Count number of items on stack */
		cpu.Ds.Push(Word(cpu.Ds.Size()))
	case TO_R: /* Move top item to the return stack. */
		v1, _ = cpu.Ds.Pop()
		cpu.Rs.Push(v1)
	case R_FROM: /* Retrieve item from the return stack. */
		v1, _ = cpu.Rs.Pop()
		cpu.Ds.Push(v1)
	case R_FETCH: /* Copy top of return stack onto stack */
		v1, _ = cpu.Rs.Get()
		cpu.Ds.Push(v1)
	case ADD:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 + v2)
	case SUB:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 - v2)
	case MUL:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 * v2)
	case DIV:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 / v2)
	case DIVMOD:
		v1, v2, _ = cpu.Ds.Pop2()
		quot := v1 / v2
		rem := v1 % v2
		if rem != 0 && v1*v2 < 0 {
			quot--
		}
		cpu.Ds.Push(rem)
		cpu.Ds.Push(quot)
	case MAX:
		v1, v2, _ = cpu.Ds.Pop2()
		if v1 > v2 {
			cpu.Ds.Push(v1)
		} else {
			cpu.Ds.Push(v2)
		}
	case MIN:
		v1, v2, _ = cpu.Ds.Pop2()
		if v1 < v2 {
			cpu.Ds.Push(v1)
		} else {
			cpu.Ds.Push(v2)
		}
	case ABS:
		v1, _ = cpu.Ds.Pop()
		if v1 < 0 {
			cpu.Ds.Push(-v1)
		} else {
			cpu.Ds.Push(v1)
		}
	case MOD:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 % v2)
	case LSHIFT:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 << v2)
	case RSHIFT:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 >> v2)
	case AND:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 & v2)
	case OR:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 | v2)
	case XOR:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 ^ v2)
	case NOT:
		v1, _ = cpu.Ds.Pop()
		cpu.Ds.PushBool(v1 == 0)
	case EQ: /* Compare Equal */
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.PushBool(v1 == v2)
	case NE: /* Compare for Not Equal */
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.PushBool(v1 != v2)
	case GE: /* Compare for Greater Or Equal */
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.PushBool(v1 >= v2)
	case GT: /* Compare for Greater */
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.PushBool(v1 > v2)
	case LE: /* Compare for Equal or Less */
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.PushBool(v1 <= v2)
	case LT: /* Compare for Less */
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.PushBool(v1 < v2)
	case STORE:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.mmu.WriteW(Addr(v2), v1)
	case STORE_B:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.mmu.WriteB(Addr(v2), byte(v1))
	case FETCH:
		v1, _ := cpu.Ds.Pop()
		value := cpu.mmu.ReadW(Addr(v1))
		// fmt.Println("FETCH: ---", int(v1), int(value))
		cpu.Ds.Push(value)
	case FETCH_B:
		v1, _ := cpu.Ds.Pop()
		value := Word(cpu.mmu.ReadB(Addr(v1)))
		// fmt.Println("FETCH_B: ---", int(v1), int(value))
		cpu.Ds.Push(value)
	case JNZ: // jump if not zero
		v1, v2, _ := cpu.Ds.Pop2()
		// fmt.Println("JNZ: ---", int(v1), int(v2))
		if v1 != 0 {
			cpu.pc = Addr(v2)
		}
	case JZ: // jump if zero
		v1, v2, _ := cpu.Ds.Pop2()
		// fmt.Println("JZ: ---", int(v1), int(v2))
		if v1 == 0 {
			cpu.pc = Addr(v2)
		}
	case JMP:
		v1, _ := cpu.Ds.Pop()
		cpu.pc = Addr(v1)
	case PUSHRSP:
		cpu.Ds.Push(Word(cpu.Rs.pointer))
	case POPRSP:
		v1, _ = cpu.Ds.Pop()
		cpu.Rs.pointer = Addr(v1)
	case PUSHRBP:
		cpu.Ds.Push(Word(cpu.Rs.origin))
	case POPRBP:
		v1, _ = cpu.Ds.Pop()
		cpu.Rs.origin = Addr(v1)
	case PUSHPC:
		cpu.Ds.Push(Word(cpu.pc))
	case CALL:
		v1, _ = cpu.Ds.Pop()
		cpu.Rs.Push(Word(cpu.pc))
		// cpu.mmu.WriteW(cpu.rsp, Word(cpu.rsp))            // store rsp
		// cpu.mmu.WriteW(cpu.rsp+1*WordSize, Word(cpu.rbp)) // store rbp
		// cpu.mmu.WriteW(cpu.rsp+2*WordSize, Word(cpu.pc))  // store pc
		// cpu.rbp = cpu.rsp + 3*WordSize
		// cpu.rsp = cpu.rbp
		cpu.pc = Addr(v1)
	case RET:
		// cpu.pc = Addr(cpu.mmu.ReadW(cpu.rsp - 1*WordSize))  // return
		// cpu.rbp = Addr(cpu.mmu.ReadW(cpu.rsp - 2*WordSize)) // restore rbp
		// cpu.rsp = Addr(cpu.mmu.ReadW(cpu.rsp - 3*WordSize)) // restore rsp
		v1, _ = cpu.Rs.Pop()
		cpu.pc = Addr(v1)
	}
	return nil
}
