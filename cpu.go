package main

import (
    "os"
    "fmt"
    "unsafe"
    "bufio"
    "strings"
    "strconv"
)

const MemorySize = 128

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
    MAX     /* Leave greater of two numbers */
    MIN     /* Leave lesser of two numbers */
    ABS     /* Absolute value */
    AND
    OR
    XOR
    EQ        /* Compare Equal */
    NOT_EQ    /* Compare for Not Equal */
    EQ_GREAT  /* Compare for Greater Or Equal */
    GREAT     /* Compare for Greater */
    EQ_LESS   /* Compare for Equal or Less */
    LESS      /* Compare for Less */
    JMPC  /* Jump if condition */
    STORE
    STORE_ABS
    LOAD
    LOAD_ABS
    GET_RSP
    INC_RSP
    SET_RSP
    GET_RBP
    INC_RBP
    SET_RBP
    GET_PC
    SET_PC
    CALL
    RET
)

var Names = map[string]Word {
    "NOP": NOP,
    "HLT": HLT,
    // "PUSH": PUSH,
    "ZERO": ZERO,
    "DROP": DROP,
    "DUP": DUP,
    "SWAP": SWAP,
    "ADD": ADD,
    "SUB": SUB,
    "MUL": DIV,
    "INC": INC,
    "DEC": DEC,
    "MAX": MAX,
    "MIN": MIN,
    "ABS": ABS,
    "AND": AND,
    "OR": OR,
    "XOR": XOR,
    "=": EQ,
    "<>": NOT_EQ,
    ">": GREAT,
    ">=": EQ_GREAT,
    "<": LESS,
    "<=": EQ_LESS,
    "JMPC": JMPC,
    "STORE": STORE,
    "LOAD": LOAD,
    "CALL": CALL,
    "RET": RET,
}

type Halt struct {
}

func (e *Halt) Error() string {
	return "Halt"
}

type CPU struct {
    memory []Word // [MemorySize]int
    pc int  // Program counter
    sp int  // Stack pointer (top of the stack)
    rsp int // Return stack pointer
    rbp int // Return stack base
}

func NewCPU(prog []Word) (cpu *CPU) {
    cpu = new(CPU)
    cpu.memory = make([]Word, MemorySize)
    cpu.pc = 0
    cpu.sp = cpu.Size() - 1
    cpu.rbp = len(prog)
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

func (cpu *CPU) PrintMemory() {
    memory := unsafe.Slice((*int)(unsafe.Pointer(&cpu.memory[0])), len(cpu.memory))
    fmt.Println(memory)
}

func (cpu *CPU) Eval() (error) {
    op := cpu.memory[cpu.pc]
    // fmt.Printf("pc: %4d sp: %4d op: %s\n", cpu.pc, cpu.sp, op.String())
    fmt.Printf("pc: %4d  sp: %4d  rbp: %4d  rsp: %4d  op: %s\n", cpu.pc, cpu.sp, cpu.rbp, cpu.rsp, op.String())

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
    case MAX:
        v1, v2, _ := cpu.Pop2()
        if v1 > v2 {
            cpu.Push(v1)
        } else {
            cpu.Push(v2)
        }
        break
    case MIN:
        v1, v2, _ := cpu.Pop2()
        if v1 < v2 {
            cpu.Push(v1)
        } else {
            cpu.Push(v2)
        }
        break
    case ABS:
        v1, _ := cpu.Pop()
        if v1 < 0 {
            cpu.Push(-v1)
        } else {
            cpu.Push(v1)
        }
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
        cpu.memory[int(addr) + cpu.rbp] = value
        break
    case STORE_ABS:
        value, addr, _ := cpu.Pop2()
        cpu.memory[int(addr)] = value
    case LOAD:
        addr, _ := cpu.Pop()
        value := cpu.memory[int(addr) + cpu.rbp]
        cpu.Push(value)
        break
    case LOAD_ABS:
        addr, _ := cpu.Pop()
        value := cpu.memory[int(addr)]
        fmt.Println("LOAD_ABS: ---", int(addr), int(value))
        cpu.Push(value)
    case JMPC:
        cond, addr, _ := cpu.Pop2()
        if cond != 0 {
            cpu.pc = int(addr)
        }
    case GET_RSP:
        cpu.Push(Word(cpu.rsp))
        break
    case INC_RSP:
        cpu.rsp++
        break
    case SET_RSP:
        addr, _ := cpu.Pop()
        cpu.rsp = int(addr)
        break
    case GET_RBP:
        cpu.Push(Word(cpu.rbp))
        break
    case INC_RBP:
        cpu.rbp++
        break
    case SET_RBP:
        addr, _ := cpu.Pop()
        cpu.rbp = int(addr)
        break
    case GET_PC:
        cpu.Push(Word(cpu.pc))
        break
    case SET_PC:
        addr, _ := cpu.Pop()
        cpu.pc = int(addr)
        break
    case CALL:
        addr, _ := cpu.Pop()
        cpu.memory[cpu.rsp] = Word(cpu.rsp) // store rsp
        cpu.memory[cpu.rsp + 1] = Word(cpu.rbp) // store rbp
        cpu.memory[cpu.rsp + 2] = Word(cpu.pc) // store pc
        cpu.rbp = cpu.rsp + 3
        cpu.rsp = cpu.rbp
        cpu.pc = int(addr)
        break
    case RET:
        cpu.pc = int(cpu.memory[cpu.rsp - 1]) // return
        cpu.rbp = int(cpu.memory[cpu.rsp - 2]) // restore rbp
        cpu.rsp = int(cpu.memory[cpu.rsp - 3]) // restore rsp
        break
    }
    return nil
}

// https://www.bernhard-baehr.de/pdp8e/pal8.html
func compile(filename string) ([]Word, error) {
    pc := 0
    program := make([]Word, 0)
    labels := map[string]int{}
    variables := map[string]int{}
    lastVariable := 0
    file, err := os.Open(filename)
    if err != nil {
        fmt.Println(err)
        return nil, err
    }
    defer file.Close()
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        t := scanner.Text()
        if len(t) == 0 {
            continue
        }
        for _, token := range strings.Fields(t) {
            token = strings.ToUpper(token)
            if strings.HasPrefix(token, "/") { // Start of comment. The rest of the current line is ignored.
                break
            }
            if strings.HasPrefix(token, "@") { // Variable
                variable := strings.ToUpper(strings.TrimPrefix(token, "@"))
                addr, exists := variables[variable]
                if ! exists { // New variable
                    addr = lastVariable
                    lastVariable++
                    variables[variable] = addr
                    // RSP++
                    var code = []Word{ INC_RSP }
                    program = append(program, code...)
                    pc += len(code)
                    fmt.Printf("%04d %s\n", pc, code[0])
                }
                fmt.Printf("%04d %s %d\n", pc, PUSH, addr)
                var code = []Word{ PUSH,Word(addr) }
                program = append(program, code...)
                pc += len(code)
                continue
            }
            if strings.HasSuffix(token, ",") { // Define a symbol with the value of the current location counter (used to define labels)
                label := strings.ToUpper(strings.TrimSuffix(token, ","))
                labels[label] = pc
                continue
            }
            if c, exists := Names[token]; exists { // Token
                fmt.Printf("%04d %s\n", pc, Word(c))
                var code = []Word{ Word(c) }
                program = append(program, code...)
                pc = pc + len(code)

            } else if c, exists := labels[token]; exists { // Label
                program = append(program, PUSH, Word(c))
                fmt.Printf("%04d PUSH %d (%s)\n", pc, c, token)
                pc++

            } else { // Push
                value, err := strconv.Atoi(token)
                if err != nil {
                    fmt.Println(token, err)
                    return nil, err
                }
                var code = []Word{ PUSH, Word(value) }
                program = append(program, code...)
                pc = pc + len(code)
                fmt.Printf("%04d PUSH %d\n", pc, value)
            }
        }
    }

    program1 := unsafe.Slice((*int)(unsafe.Pointer(&program[0])), len(program))
    fmt.Println(program1)
    fmt.Println(labels)

    if err := scanner.Err(); err != nil {
        fmt.Println(err)
    }
    return program, nil
}

func main() {
    prog, err := compile(os.Args[1])
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
