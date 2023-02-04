package main

import (
    "fmt"
    "unsafe"
)

const MemorySize = 64

type Word int
//go:generate stringer -type=Word

const (
    NOP Word = iota
    HLT
    PUSH    /* Push data onto stack */
    ZERO    /* Push 0 onto stack */
    DROP    /* Discards the top stack item */
    DUP     /* Duplicates the top stack item */
    SWAP    /* Reverses the top two stack items */
    ADD
    SUB     /* Subtraction */
    MUL
    DIV
    INC     /* Increment by 1*/
    DEC     /* Decrement by 1 */
    AND
    OR
    XOR
    EQ        /* Compare Equal */
    NOT_EQ    /* Compare for Not Equal */
    EQ_GREAT  /* Compare for Greater Or Equal */
    GREAT     /* Compare for Greater */
    EQ_LESS   /* Compare for Equal or Less */
    LESS      /* Compare for Less */
    JUMP_CON  /* Jump if condition */
    STORE
    LOAD
)


// var prog = []int{ PUSH, 79, PUSH, 1, ADD, PUSH, 40, DIV, HLT}
var prog = []Word{
// PUSH, 79, PUSH, 1, ADD, PUSH, 40, DIV, INC, PUSH, 3, EQ,
PUSH, 10,
PUSH, 31, STORE,

PUSH, 31, LOAD, // 5
DEC,
DUP,
PUSH, 31, STORE,

PUSH, 30, LOAD,
INC,
PUSH, 30, STORE,

PUSH, 0, GREAT,
PUSH, 5, JUMP_CON,

PUSH, 30, LOAD,
PUSH, 5,
SWAP,
}

type CPU struct {
    memory []Word // [MemorySize]int
    pc int
    sp int
    ds int
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

func (cpu *CPU) Pop() (Word, error) {
    cpu.sp++ // TODO out of stack
    value := cpu.memory[cpu.sp]
    return value, nil
}

func (cpu *CPU) Dup() (error) {
    value := cpu.memory[cpu.sp + 1]
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

func main() {
    cpu := new(CPU)
    cpu.memory = make([]Word, MemorySize)
    cpu.pc = 0
    cpu.sp = cpu.Size() - 1
    cpu.ds = len(prog)
    copy(cpu.memory, prog)
    // fmt.Println(prog)
    for {
        op := prog[cpu.pc]
        fmt.Printf("pc: %4d sp: %4d op: %s\n", cpu.pc, cpu.sp, op.String())

        cpu.pc++
        switch op {
        case NOP:
            break
        case HLT:
            fmt.Println(cpu.memory)
            return
        case PUSH:
            cpu.Push(prog[cpu.pc])
            cpu.pc++
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
        case SWAP: /* Reverses the top two stack items */
            v1, v2, _ := cpu.Pop2()
            cpu.Push(v2)
            cpu.Push(v1)
            break
        case ADD:
            v1, v2, _ := cpu.Pop2()
            cpu.Push(v1 + v2)
            break
        case SUB:
            v1, v2, _ := cpu.Pop2()
            cpu.Push(v1 - v2)
            break
        case MUL:
            v1, v2, _ := cpu.Pop2()
            cpu.Push(v1 * v2)
            break
        case DIV:
            v1, v2, _ := cpu.Pop2()
            cpu.Push(v1 / v2)
            break
        case INC: /* Increment by 1 */
            v1, _ := cpu.Pop()
            cpu.Push(v1 + 1)
            break
        case DEC: /* Decrement by 1 */
            v1, _ := cpu.Pop()
            cpu.Push(v1 - 1)
            break
        case AND:
            v1, v2, _ := cpu.Pop2()
            cpu.Push(v1 & v2)
            break
        case OR:
            v1, v2, _ := cpu.Pop2()
            cpu.Push(v1 | v2)
            break
        case XOR:
            v1, v2, _ := cpu.Pop2()
            cpu.Push(v1 ^ v2)
            break
        case EQ: /* Compare Equal */
            v1, v2, _ := cpu.Pop2()
            cpu.PushBool(v1 == v2)
            break
        case NOT_EQ: /* Compare for Not Equal */
            v1, v2, _ := cpu.Pop2()
            cpu.PushBool(v1 != v2)
            break
        case EQ_GREAT: /* Compare for Greater Or Equal */
            v1, v2, _ := cpu.Pop2()
            cpu.PushBool(v1 >= v2)
            break
        case GREAT: /* Compare for Greater */
            v1, v2, _ := cpu.Pop2()
            cpu.PushBool(v1 > v2)
            break
        case EQ_LESS: /* Compare for Equal or Less */
            v1, v2, _ := cpu.Pop2()
            cpu.PushBool(v1 <= v2)
            break
        case LESS: /* Compare for Less */
            v1, v2, _ := cpu.Pop2()
            cpu.PushBool(v1 < v2)
            break
        case STORE:
            value, addr, _ := cpu.Pop2()
            cpu.memory[addr] = value
            // cpu.Push(value)
            break
        case LOAD:
            addr, _ := cpu.Pop()
            value := cpu.memory[addr]
            cpu.Push(value)
            break
        case JUMP_CON:
            cond, addr, _ := cpu.Pop2()
            if cond != 0 {
                cpu.pc = int(addr)
            }
        }
        if cpu.pc >= len(prog) {
            break
        }
    }
    memory := unsafe.Slice((*int)(unsafe.Pointer(&cpu.memory[0])), len(cpu.memory))
    fmt.Println(memory)
    // fmt.Println(cpu.memory)
    // fmt.Println(cpu.memory[cpu.sp+1])
}
