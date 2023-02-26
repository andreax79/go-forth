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

//go:generate stringer -type=Word
const (
	HLT Word = iota + 0x10000
	NOP
	EMIT
	PERIOD

	/* Stack manipulation */
	PUSH  /* Push data onto stack */
	ZERO  /* Push 0 onto stack */
	DUP   /* Duplicates the top stack item */
	DROP  /* Discards the top stack item */
	SWAP  /* Reverses the top two stack items */
	OVER  /* Make copy of second item on top */
	ROT   /* Rotate third item to top */
	PICK  /* Copy n-th item to too */
	ROLL  /* Rotate n-th Item to top. */
	DEPTH /* Count number of items on stack */

	/* Return Stack manipulation */
	TO_R    /* Move top item to the return stack. */
	R_FROM  /* Retrieve item from the return stack. */
	R_FETCH /* Copy top of return stack onto stack */

	/* Arithmetic */
	ADD /* Add */
	SUB /* Subtract */
	MUL /* Multiply */
	DIV /* Divide */
	MAX /* Leave greater of two numbers */
	MIN /* Leave lesser of two numbers */
	ABS /* Absolute value */
	MOD /* Modulo */

	/* Logical */
	AND /* Bitwise and */
	OR  /* Bitwise or */
	XOR /* Bitwise xor */
	NOT /* Reverse true value */

	/* Comparison */
	EQ /* Compare Equal */
	NE /* Compare for Not Equal */
	GE /* Compare for Greater Or Equal */
	GT /* Compare for Greater */
	LE /* Compare for Equal or Less */
	LT /* Compare for Less */

	/* Control and subroutines */
	JNZ /* jump if not zero */
	JZ  /* jump if zero */
	JMP /* Jump */
	CALL
	RET

	//  PUSHPC /* Push PC */
	//  POPPC /* Pop -> PC */

	/* Memory */
	STORE
	FETCH
	FETCH_B

	/* Registers */
	GET_RSP
	SET_RSP
	GET_RBP
	SET_RBP
	GET_PC
	SET_PC
)

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
	var op Word
	op = cpu.mmu.ReadW(cpu.pc)
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
	op := cpu.mmu.ReadW(cpu.pc)
	if cpu.Verbose {
		cpu.PrintRegisters()
	}
	cpu.Time++
	if cpu.Limit != 0 && cpu.Time >= cpu.Limit {
		return new(Halt)
	}

	cpu.pc += WordSize
	switch op {
	case NOP:
		break
	case HLT:
		return new(Halt)
	case PUSH:
		cpu.Ds.Push(cpu.mmu.ReadW(cpu.pc))
		cpu.pc += WordSize
		break
	case EMIT: // TODO
		v1, _ = cpu.Ds.Pop()
		fmt.Printf(">>>> %d\n", int(v1))
		break
	case PERIOD: // TODO
		v1, _ = cpu.Ds.Pop()
		fmt.Printf(">>>> %d\n", int(v1))
		break
	case ZERO:
		cpu.Ds.Push(0)
		break
	case DROP: /* Discards the top stack item */
		cpu.Ds.Pop()
		break
	case DUP: /* Duplicates the top stack item */
		cpu.Ds.Dup()
		break
	case SWAP: /* Reverses the top two stack items */
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v2)
		cpu.Ds.Push(v1)
		break
	case OVER: /* Push a copy of the second element on the stack */
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1)
		cpu.Ds.Push(v2)
		cpu.Ds.Push(v1)
		break
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
		break
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
		break
	case SUB:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 - v2)
		break
	case MUL:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 * v2)
		break
	case DIV:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 / v2)
		break
	case MAX:
		v1, v2, _ = cpu.Ds.Pop2()
		if v1 > v2 {
			cpu.Ds.Push(v1)
		} else {
			cpu.Ds.Push(v2)
		}
		break
	case MIN:
		v1, v2, _ = cpu.Ds.Pop2()
		if v1 < v2 {
			cpu.Ds.Push(v1)
		} else {
			cpu.Ds.Push(v2)
		}
		break
	case ABS:
		v1, _ = cpu.Ds.Pop()
		if v1 < 0 {
			cpu.Ds.Push(-v1)
		} else {
			cpu.Ds.Push(v1)
		}
		break
	case MOD:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 % v2)
		break
	case AND:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 & v2)
		break
	case OR:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 | v2)
		break
	case XOR:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.Push(v1 ^ v2)
		break
	case NOT:
		v1, _ = cpu.Ds.Pop()
		cpu.Ds.PushBool(v1 == 0)
		break
	case EQ: /* Compare Equal */
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.PushBool(v1 == v2)
		break
	case NE: /* Compare for Not Equal */
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.PushBool(v1 != v2)
		break
	case GE: /* Compare for Greater Or Equal */
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.PushBool(v1 >= v2)
		break
	case GT: /* Compare for Greater */
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.PushBool(v1 > v2)
		break
	case LE: /* Compare for Equal or Less */
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.PushBool(v1 <= v2)
		break
	case LT: /* Compare for Less */
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.Ds.PushBool(v1 < v2)
		break
	case STORE:
		v1, v2, _ = cpu.Ds.Pop2()
		cpu.mmu.WriteW(Addr(v2), v1)
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
	case GET_RSP:
		cpu.Ds.Push(Word(cpu.Rs.pointer))
		break
	case SET_RSP:
		v1, _ = cpu.Ds.Pop()
		cpu.Rs.pointer = Addr(v1)
		break
	case GET_RBP:
		cpu.Ds.Push(Word(cpu.Rs.origin))
		break
	case SET_RBP:
		v1, _ = cpu.Ds.Pop()
		cpu.Rs.origin = Addr(v1)
		break
	case GET_PC:
		cpu.Ds.Push(Word(cpu.pc))
		break
	case SET_PC:
		v1, _ = cpu.Ds.Pop()
		cpu.pc = Addr(v1)
		break
	case CALL:
		v1, _ = cpu.Ds.Pop()
		cpu.Rs.Push(Word(cpu.pc))
		// cpu.mmu.WriteW(cpu.rsp, Word(cpu.rsp))            // store rsp
		// cpu.mmu.WriteW(cpu.rsp+1*WordSize, Word(cpu.rbp)) // store rbp
		// cpu.mmu.WriteW(cpu.rsp+2*WordSize, Word(cpu.pc))  // store pc
		// cpu.rbp = cpu.rsp + 3*WordSize
		// cpu.rsp = cpu.rbp
		cpu.pc = Addr(v1)
		break
	case RET:
		// cpu.pc = Addr(cpu.mmu.ReadW(cpu.rsp - 1*WordSize))  // return
		// cpu.rbp = Addr(cpu.mmu.ReadW(cpu.rsp - 2*WordSize)) // restore rbp
		// cpu.rsp = Addr(cpu.mmu.ReadW(cpu.rsp - 3*WordSize)) // restore rsp
		v1, _ = cpu.Rs.Pop()
		cpu.pc = Addr(v1)
		break
	}
	return nil
}
