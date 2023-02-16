package fcpu

import (
	"fmt"
	"unsafe"
)

const DataStackTop = 1 << 16

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
	CDUP  /* ?DUP - Duplicate only if non-zero */
	DROP  /* Discards the top stack item */
	SWAP  /* Reverses the top two stack items */
	OVER  /* Make copy of second item on top */
	ROT   /* Rotate third item to top */
	PICK  /* Copy n-th item to too */
	ROLL  /* Rotate n-th Item to top. */
	DEPTH /* Count number of items on stack */

	/* Arithmetic */
	ADD     /* Add */
	SUB     /* Subtract */
	MUL     /* Multiply */
	DIV     /* Divide */
	ADD_ONE /* Increment by 1*/
	SUB_ONE /* Decrement by 1 */
	MAX     /* Leave greater of two numbers */
	MIN     /* Leave lesser of two numbers */
	ABS     /* Absolute value */
	MOD     /* Modulo */

	/* Logical */
	AND /* Bitwise and */
	OR  /* Bitwise or */
	XOR /* Bitwise xor */
	NOT /* Reverse true value */

	/* Comparison */
	EQ       /* Compare Equal */
	NOT_EQ   /* Compare for Not Equal */
	EQ_GREAT /* Compare for Greater Or Equal */
	GREAT    /* Compare for Greater */
	EQ_LESS  /* Compare for Equal or Less */
	LESS     /* Compare for Less */

	/* Control and subroutines */
	JCC /* Jump if condition is met */
	JMP /* Jump */
	CALL
	RET

	/* Memory */
	STORE
	// STORE_ABS
	LOAD
	// LOAD_ABS

	/* Registers */
	GET_RSP
	INC_RSP
	SET_RSP
	GET_RBP
	INC_RBP
	SET_RBP
	GET_PC
	SET_PC
)

type Halt struct {
}

func (e *Halt) Error() string {
	return "Halt"
}

type CPU struct {
	mmu *MMU   // Memory Management Unit
	pc  Addr   // Program counter
	ds  *Stack // Data Stack
	rsp Addr   // Return stack pointer
	rbp Addr   // Return stack base
}

func NewCPU(prog []byte) (cpu *CPU) {
	cpu = new(CPU)
	cpu.mmu = NewMMU()
	cpu.pc = 0
	cpu.ds = NewStack(cpu.mmu, DataStackTop)
	cpu.rbp = Addr(len(prog))
	cpu.rsp = cpu.rbp
	cpu.mmu.WriteBytes(0, prog)
	return cpu
}

func (cpu *CPU) PrintRegisters() {
	var op Word
	op = cpu.mmu.ReadW(cpu.pc)
	fmt.Printf("pc: %4d  sp: %4d  rbp: %4d  rsp: %4d  op: %-15s  stack: %s\n",
		cpu.pc, cpu.ds.pointer, cpu.rbp, cpu.rsp, op.String(), cpu.ds,
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
	cpu.PrintRegisters()

	cpu.pc += WordSize
	switch op {
	case NOP:
		break
	case HLT:
		return new(Halt)
	case PUSH:
		cpu.ds.Push(cpu.mmu.ReadW(cpu.pc))
		cpu.pc += WordSize
		break
	case EMIT:
		v1, _ = cpu.ds.Pop()
		fmt.Printf(">>>> %d\n", int(v1))
		break
	case ZERO:
		cpu.ds.Push(0)
		break
	case DROP: /* Discards the top stack item */
		cpu.ds.Pop()
		break
	case DUP: /* Duplicates the top stack item */
		cpu.ds.Dup()
		break
	case CDUP: /* Duplicates the top stack item */
		v1, _ = cpu.ds.Get()
		if v1 != 0 {
			cpu.ds.Dup()
		}
		break
	case SWAP: /* Reverses the top two stack items */
		v1, v2, _ = cpu.ds.Pop2()
		cpu.ds.Push(v2)
		cpu.ds.Push(v1)
		break
	case OVER: /* Push a copy of the second element on the stack */
		v1, v2, _ = cpu.ds.Pop2()
		cpu.ds.Push(v1)
		cpu.ds.Push(v2)
		cpu.ds.Push(v1)
		break
	case ROT: /* Rotate the third item to top */
		v1, v2, _ = cpu.ds.Pop2()
		v3, _ = cpu.ds.Pop()
		cpu.ds.Push(v1)
		cpu.ds.Push(v2)
		cpu.ds.Push(v3)
	case PICK: /* Remove u. Copy the x-u to the top of the stack. */
		v1, _ = cpu.ds.Pop()
		cpu.ds.Pick(v1)
	case ROLL: /* Remove u.  Rotate u+1 items on the top of the stack */
		v1, _ = cpu.ds.Pop()
		cpu.ds.Roll(v1)
	case DEPTH: /* Count number of items on stack */
		cpu.ds.Push(Word(cpu.ds.Size()))
		break
	case ADD:
		v1, v2, _ = cpu.ds.Pop2()
		cpu.ds.Push(v1 + v2)
		break
	case SUB:
		v1, v2, _ = cpu.ds.Pop2()
		cpu.ds.Push(v1 - v2)
		break
	case MUL:
		v1, v2, _ = cpu.ds.Pop2()
		cpu.ds.Push(v1 * v2)
		break
	case DIV:
		v1, v2, _ = cpu.ds.Pop2()
		cpu.ds.Push(v1 / v2)
		break
	case ADD_ONE: /* Increment by 1 */
		v1, _ = cpu.ds.Pop()
		cpu.ds.Push(v1 + 1)
		break
	case SUB_ONE: /* Decrement by 1 */
		v1, _ = cpu.ds.Pop()
		cpu.ds.Push(v1 - 1)
		break
	case MAX:
		v1, v2, _ = cpu.ds.Pop2()
		if v1 > v2 {
			cpu.ds.Push(v1)
		} else {
			cpu.ds.Push(v2)
		}
		break
	case MIN:
		v1, v2, _ = cpu.ds.Pop2()
		if v1 < v2 {
			cpu.ds.Push(v1)
		} else {
			cpu.ds.Push(v2)
		}
		break
	case ABS:
		v1, _ = cpu.ds.Pop()
		if v1 < 0 {
			cpu.ds.Push(-v1)
		} else {
			cpu.ds.Push(v1)
		}
		break
	case MOD:
		v1, v2, _ = cpu.ds.Pop2()
		cpu.ds.Push(v1 % v2)
		break
	case AND:
		v1, v2, _ = cpu.ds.Pop2()
		cpu.ds.Push(v1 & v2)
		break
	case OR:
		v1, v2, _ = cpu.ds.Pop2()
		cpu.ds.Push(v1 | v2)
		break
	case XOR:
		v1, v2, _ = cpu.ds.Pop2()
		cpu.ds.Push(v1 ^ v2)
		break
	case NOT:
		v1, _ = cpu.ds.Pop()
		cpu.ds.PushBool(v1 == 0)
		break
	case EQ: /* Compare Equal */
		v1, v2, _ = cpu.ds.Pop2()
		cpu.ds.PushBool(v1 == v2)
		break
	case NOT_EQ: /* Compare for Not Equal */
		v1, v2, _ = cpu.ds.Pop2()
		cpu.ds.PushBool(v1 != v2)
		break
	case EQ_GREAT: /* Compare for Greater Or Equal */
		v1, v2, _ = cpu.ds.Pop2()
		cpu.ds.PushBool(v1 >= v2)
		break
	case GREAT: /* Compare for Greater */
		v1, v2, _ = cpu.ds.Pop2()
		cpu.ds.PushBool(v1 > v2)
		break
	case EQ_LESS: /* Compare for Equal or Less */
		v1, v2, _ = cpu.ds.Pop2()
		cpu.ds.PushBool(v1 <= v2)
		break
	case LESS: /* Compare for Less */
		v1, v2, _ = cpu.ds.Pop2()
		cpu.ds.PushBool(v1 < v2)
		break
	// case STORE:
	// 	v1, v2, _ = cpu.ds.Pop2()
	// 	cpu.mmu.WriteW(Addr(v2)+cpu.rbp, v1)
	// 	break
	case STORE:
		v1, v2, _ = cpu.ds.Pop2()
		cpu.mmu.WriteW(Addr(v2), v1)
	// case LOAD:
	// 	v1, _ := cpu.ds.Pop()
	// 	value := cpu.mmu.ReadW(Addr(v1) + cpu.rbp)
	// 	// fmt.Println("LOAD: ---", int(v1), int(value))
	// 	cpu.ds.Push(value)
	// 	break
	case LOAD:
		v1, _ := cpu.ds.Pop()
		value := cpu.mmu.ReadW(Addr(v1))
		// fmt.Println("LOAD_ABS: ---", int(v1), int(value))
		cpu.ds.Push(value)
	case JCC:
		v1, v2, _ := cpu.ds.Pop2()
		// fmt.Println("JCC: ---", int(v1), int(v2))
		if v1 != 0 {
			cpu.pc = Addr(v2)
		}
	case JMP:
		v1, _ := cpu.ds.Pop()
		cpu.pc = Addr(v1)
	case GET_RSP:
		cpu.ds.Push(Word(cpu.rsp))
		break
	case INC_RSP:
		cpu.rsp += WordSize
		break
	case SET_RSP:
		v1, _ = cpu.ds.Pop()
		cpu.rsp = Addr(v1)
		break
	case GET_RBP:
		cpu.ds.Push(Word(cpu.rbp))
		break
	case INC_RBP:
		cpu.rbp += WordSize
		break
	case SET_RBP:
		v1, _ = cpu.ds.Pop()
		cpu.rbp = Addr(v1)
		break
	case GET_PC:
		cpu.ds.Push(Word(cpu.pc))
		break
	case SET_PC:
		v1, _ = cpu.ds.Pop()
		cpu.pc = Addr(v1)
		break
	case CALL:
		v1, _ = cpu.ds.Pop()
		cpu.mmu.WriteW(cpu.rsp, Word(cpu.rsp))            // store rsp
		cpu.mmu.WriteW(cpu.rsp+1*WordSize, Word(cpu.rbp)) // store rbp
		cpu.mmu.WriteW(cpu.rsp+2*WordSize, Word(cpu.pc))  // store pc
		cpu.rbp = cpu.rsp + 3*WordSize
		cpu.rsp = cpu.rbp
		cpu.pc = Addr(v1)
		break
	case RET:
		cpu.pc = Addr(cpu.mmu.ReadW(cpu.rsp - 1*WordSize))  // return
		cpu.rbp = Addr(cpu.mmu.ReadW(cpu.rsp - 2*WordSize)) // restore rbp
		cpu.rsp = Addr(cpu.mmu.ReadW(cpu.rsp - 3*WordSize)) // restore rsp
		break
	}
	return nil
}
