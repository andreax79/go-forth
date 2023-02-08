package main

import (
    "os"
    "fmt"
    "unsafe"
)

const MemorySize = 128

type Addr uint32

type Word int
//go:generate stringer -type=Word

const (
    HLT Word = iota
    NOP
    EMIT

    /* Stack manipulation */
    PUSH    /* Push data onto stack */
    ZERO    /* Push 0 onto stack */
    DUP     /* Duplicates the top stack item */
    CDUP    /* ?DUP - Duplicate only if non-zero */
    DROP    /* Discards the top stack item */
    SWAP    /* Reverses the top two stack items */
    // OVER /* Make copy of second item on top */
    // ROT  /* Rotate third item to top */
    // PICK /* Copy n-th item to too */
    // ROLL
    // DEPTH /* Count number of items on stack */

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
    AND     /* Bitwise and */
    OR      /* Bitwise or */
    XOR     /* Bitwise xor */
    NOT     /* Reverse true value */

    /* Comparison */
    EQ        /* Compare Equal */
    NOT_EQ    /* Compare for Not Equal */
    EQ_GREAT  /* Compare for Greater Or Equal */
    GREAT     /* Compare for Greater */
    EQ_LESS   /* Compare for Equal or Less */
    LESS      /* Compare for Less */

    /* Control and subroutines */
    JMPC      /* Jump if condition */
    CALL
    RET

    /* Memory */
    STORE
    STORE_ABS
    LOAD
    LOAD_ABS

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
    memory []Word // [MemorySize]int
    pc Addr  // Program counter
    sp Addr  // Stack pointer (top of the stack)
    rsp Addr // Return stack pointer
    rbp Addr // Return stack base
}

func NewCPU(prog []Word) (cpu *CPU) {
    cpu = new(CPU)
    cpu.memory = make([]Word, MemorySize)
    cpu.pc = 0
    cpu.sp = Addr(cpu.Size() - 1)
    cpu.rbp = Addr(len(prog))
    cpu.rsp = cpu.rbp
    copy(cpu.memory, prog)
    return cpu
}

func (cpu *CPU) Size() int {
    return len(cpu.memory)
}

func (cpu *CPU) Push(value Word) (error) {
    cpu.memory[cpu.sp] = value
    cpu.sp-- // TODO out of stack
    return nil
}

func (cpu *CPU) PushBool(value bool) (error) {
    if value {
        return cpu.Push(1)
    } else {
        return cpu.Push(0)
    }
}

func (cpu *CPU) Get() (Word, error) {
    value := cpu.memory[cpu.sp + 1] // TODO out of stack
    return value, nil
}

func (cpu *CPU) Pop() (Word, error) {
    cpu.sp++ // TODO out of stack
    value := cpu.memory[cpu.sp]
    cpu.memory[cpu.sp] = Word(0) // TODO - TEMP for debugging
    return value, nil
}

func (cpu *CPU) Dup() (error) {
    var value Word
    var err error
    if value, err = cpu.Get(); err != nil {
        return err
    }
    return cpu.Push(value)
}

func (cpu *CPU) Pop2() (Word, Word, error) {
    var v1 Word
    var v2 Word
    var err error
    if v2, err = cpu.Pop(); err != nil {
        return 0, 0, err
    }
    if v1, err = cpu.Pop(); err != nil {
        return 0, 0, err
    }
    return v1, v2, nil
}

func (cpu *CPU) PrintMemory() {
    memory := unsafe.Slice((*int)(unsafe.Pointer(&cpu.memory[0])), len(cpu.memory))
    fmt.Println(memory)
}

func (cpu *CPU) PrintRegisters() {
    var op Word
    op = cpu.memory[cpu.pc]
    fmt.Printf("pc: %4d  sp: %4d  rbp: %4d  rsp: %4d  op: %s\n", cpu.pc, cpu.sp, cpu.rbp, cpu.rsp, op.String())
}

func (cpu *CPU) PrintStack() {
    if cpu.sp+1 < Addr(len(cpu.memory)) {
        stack := unsafe.Slice((*int)(unsafe.Pointer(&cpu.memory[cpu.sp+1])), Addr(len(cpu.memory))-cpu.sp-1)
        fmt.Println("stack: ", stack)
    } else {
        fmt.Println("stack:  []")
    }
}

func (cpu *CPU) Eval() (error) {
    var v1 Word
    var v2 Word
    op := cpu.memory[cpu.pc]
    cpu.PrintRegisters()
    cpu.PrintStack()

    cpu.pc++
    switch op {
    case NOP:
        break
    case HLT:
        return new(Halt)
    case PUSH:
        cpu.Push(cpu.memory[cpu.pc])
        cpu.pc++
        break
    case EMIT:
        v1, _ = cpu.Pop()
        fmt.Printf(">>>> %d\n", int(v1))
        break
    case ZERO:
        cpu.Push(0)
        break
    case DROP: /* Discards the top stack item */
        cpu.Pop()
        break
    case DUP: /* Duplicates the top stack item */
        cpu.Dup()
        break
    case CDUP: /* Duplicates the top stack item */
        v1, _ = cpu.Get()
        if v1 != 0 {
            cpu.Dup()
        }
        break
    case SWAP: /* Reverses the top two stack items */
        v1, v2, _ = cpu.Pop2()
        cpu.Push(v2)
        cpu.Push(v1)
        break
    case ADD:
        v1, v2, _ = cpu.Pop2()
        cpu.Push(v1 + v2)
        break
    case SUB:
        v1, v2, _ = cpu.Pop2()
        cpu.Push(v1 - v2)
        break
    case MUL:
        v1, v2, _ = cpu.Pop2()
        cpu.Push(v1 * v2)
        break
    case DIV:
        v1, v2, _ = cpu.Pop2()
        cpu.Push(v1 / v2)
        break
    case ADD_ONE: /* Increment by 1 */
        v1, _ = cpu.Pop()
        cpu.Push(v1 + 1)
        break
    case SUB_ONE: /* Decrement by 1 */
        v1, _ = cpu.Pop()
        cpu.Push(v1 - 1)
        break
    case MAX:
        v1, v2, _ = cpu.Pop2()
        if v1 > v2 {
            cpu.Push(v1)
        } else {
            cpu.Push(v2)
        }
        break
    case MIN:
        v1, v2, _ = cpu.Pop2()
        if v1 < v2 {
            cpu.Push(v1)
        } else {
            cpu.Push(v2)
        }
        break
    case ABS:
        v1, _ = cpu.Pop()
        if v1 < 0 {
            cpu.Push(-v1)
        } else {
            cpu.Push(v1)
        }
        break
    case MOD:
        v1, v2, _ = cpu.Pop2()
        cpu.Push(v1 % v2)
        break
    case AND:
        v1, v2, _ = cpu.Pop2()
        cpu.Push(v1 & v2)
        break
    case OR:
        v1, v2, _ = cpu.Pop2()
        cpu.Push(v1 | v2)
        break
    case XOR:
        v1, v2, _ = cpu.Pop2()
        cpu.Push(v1 ^ v2)
        break
    case NOT:
        v1, _ = cpu.Pop()
        cpu.PushBool(v1 == 0)
        break
    case EQ: /* Compare Equal */
        v1, v2, _ = cpu.Pop2()
        cpu.PushBool(v1 == v2)
        break
    case NOT_EQ: /* Compare for Not Equal */
        v1, v2, _ = cpu.Pop2()
        cpu.PushBool(v1 != v2)
        break
    case EQ_GREAT: /* Compare for Greater Or Equal */
        v1, v2, _ = cpu.Pop2()
        cpu.PushBool(v1 >= v2)
        break
    case GREAT: /* Compare for Greater */
        v1, v2, _ = cpu.Pop2()
        cpu.PushBool(v1 > v2)
        break
    case EQ_LESS: /* Compare for Equal or Less */
        v1, v2, _ = cpu.Pop2()
        cpu.PushBool(v1 <= v2)
        break
    case LESS: /* Compare for Less */
        v1, v2, _ = cpu.Pop2()
        cpu.PushBool(v1 < v2)
        break
    case STORE:
        v1, v2, _ = cpu.Pop2()
        cpu.memory[Addr(v2) + cpu.rbp] = v1
        break
    case STORE_ABS:
        v1, v2, _ = cpu.Pop2()
        cpu.memory[int(v2)] = v1
    case LOAD:
        v1, _ := cpu.Pop()
        value := cpu.memory[Addr(v1) + cpu.rbp]
        cpu.Push(value)
        break
    case LOAD_ABS:
        v1, _ := cpu.Pop()
        value := cpu.memory[int(v1)]
        fmt.Println("LOAD_ABS: ---", int(v1), int(value))
        cpu.Push(value)
    case JMPC:
        v1, v2, _ := cpu.Pop2()
        if v1 != 0 {
            cpu.pc = Addr(v2)
        }
    case GET_RSP:
        cpu.Push(Word(cpu.rsp))
        break
    case INC_RSP:
        cpu.rsp++
        break
    case SET_RSP:
        v1, _ = cpu.Pop()
        cpu.rsp = Addr(v1)
        break
    case GET_RBP:
        cpu.Push(Word(cpu.rbp))
        break
    case INC_RBP:
        cpu.rbp++
        break
    case SET_RBP:
        v1, _ = cpu.Pop()
        cpu.rbp = Addr(v1)
        break
    case GET_PC:
        cpu.Push(Word(cpu.pc))
        break
    case SET_PC:
        v1, _ = cpu.Pop()
        cpu.pc = Addr(v1)
        break
    case CALL:
        v1, _ = cpu.Pop()
        cpu.memory[cpu.rsp] = Word(cpu.rsp) // store rsp
        cpu.memory[cpu.rsp + 1] = Word(cpu.rbp) // store rbp
        cpu.memory[cpu.rsp + 2] = Word(cpu.pc) // store pc
        cpu.rbp = cpu.rsp + 3
        cpu.rsp = cpu.rbp
        cpu.pc = Addr(v1)
        break
    case RET:
        cpu.pc = Addr(cpu.memory[cpu.rsp - 1]) // return
        cpu.rbp = Addr(cpu.memory[cpu.rsp - 2]) // restore rbp
        cpu.rsp = Addr(cpu.memory[cpu.rsp - 3]) // restore rsp
        break
    }
    return nil
}

func main() {
    prog, err := Compile(os.Args[1])
    if err != nil {
        fmt.Println(err)
        return
    }
    cpu := NewCPU(prog)
    for {
        err := cpu.Eval()
        if err != nil {
            break
        }
    }
    cpu.PrintMemory()
}
