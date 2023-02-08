package main

import (
    "os"
    "fmt"
    "unsafe"
    "bufio"
    "strings"
    "strconv"
)

// Pseudo instructions
const (
    ADD_TWO Word = iota + 0xffff
    SUB_TWO
)

var Symbols = map[string]Word {
    "HLT": HLT,
    "NOP": NOP,
    "EMIT": EMIT,

    /* Stack manipulation */
    // "PUSH": PUSH,
    "ZERO": ZERO,
    "DUP": DUP,
    "?DUP": CDUP,
    "DROP": DROP,
    "SWAP": SWAP,

    /* Arithmetic */
    "+": ADD,
    "ADD": ADD,
    "-": SUB,
    "SUB": SUB,
    "*": MUL,
    "MUL": MUL,
    "/": DIV,
    "DIV": DIV,
    "1+": ADD_ONE,
    "1-": SUB_ONE,
    "2+": ADD_TWO,
    "2-": SUB_TWO,
    "MAX": MAX,
    "MIN": MIN,
    "ABS": ABS,
    "MOD": MOD,

    /* Logical */
    "AND": AND,
    "OR": OR,
    "XOR": XOR,
    "NOT": NOT,

    /* Comparison */
    "=": EQ,
    "<>": NOT_EQ,
    ">": GREAT,
    ">=": EQ_GREAT,
    "<": LESS,
    "<=": EQ_LESS,

    /* Control and subroutines */
    "JMPC": JMPC,
    "CALL": CALL,
    "RET": RET,

    /* Memory */
    "STORE": STORE,
    "LOAD": LOAD,
}

type Pass uint8
const (
	First Pass = 1
	Second     = 2
)

// Compiler status
type CompilerStatus struct {
    pc int
    code []Word // program code
    labels map[string]int // map label names to addresses
    variables map[string]int // map variable names to addresses
    lastVariableAddr int // address of the last variable
    pass Pass // pass number (First/Second)
}

func NewCompilerStatus(pass Pass, labels map[string]int) (status *CompilerStatus) {
    status = new(CompilerStatus)
    status.pc = 0
    status.code = make([]Word, 0)
    status.variables = map[string]int{}
    status.lastVariableAddr = 0
    status.pass = pass
    if labels != nil {
        status.labels = labels
    } else {
        status.labels = map[string]int{}
    }
    return status
}

// Add compiled code
func (status *CompilerStatus) AddCode(code ...Word) {
    if code[0] == PUSH && len(code) == 2 {
        fmt.Printf("%04d %s %d\n", status.pc, code[0], int(code[1]))
    } else {
        fmt.Printf("%04d %s\n", status.pc, strings.Trim(fmt.Sprint(code), "[]"))
    }
    status.code = append(status.code, code...)
    status.pc += len(code)
}

// Compile a line
func CompileLine(status *CompilerStatus, line string) (error) {
    for _, token := range strings.Fields(line) {
        token = strings.ToUpper(token)
        if strings.HasPrefix(token, "/") { // Start of comment. The rest of the current line is ignored.
            break
        }

        if strings.HasPrefix(token, "@") { // Variable
            variable := strings.TrimPrefix(token, "@")
            addr, exists := status.variables[variable]
            if ! exists { // New variable
                addr = status.lastVariableAddr
                status.lastVariableAddr++
                status.variables[variable] = addr
                status.AddCode(INC_RSP) // Increment rsp
            }
            status.AddCode(PUSH, Word(addr))
            continue

        } else if strings.HasSuffix(token, ":") { // Define a symbol with the value of the current location counter (used to define labels)
            label := strings.TrimSuffix(token, ":")
            status.labels[label] = status.pc

        } else if c, exists := Symbols[token]; exists { // Token
            switch c {
                case ADD_TWO:
                    status.AddCode(ADD_ONE, ADD_ONE)
                case SUB_TWO:
                    status.AddCode(SUB_ONE, SUB_ONE)
                default:
                    status.AddCode(c)
            }
        } else if c, exists := status.labels[token]; exists { // Label
            status.AddCode(PUSH, Word(c))

        } else { // Push
            value, err := strconv.Atoi(token)
            if err != nil {
                // Ignore undefined labels during the first compilation pass
                if status.pass != First {
                    return err
                }
                value = 99 // TEMP
            }
            status.AddCode(PUSH, Word(value))
        }
    }
    return nil
}

// Execute a compilation pass
func CompilePass(file *os.File, pass Pass, labels map[string]int) (*CompilerStatus, error) {
    status := NewCompilerStatus(pass, labels)
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        if len(line) == 0 {
            continue
        }
        if err := CompileLine(status, line); err != nil {
            return nil, err
        }
    }
    if err := scanner.Err(); err != nil {
        return nil, err
    }
    return status, nil
}

// Compile a program file and return the compiled code
func Compile(filename string) ([]Word, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // First pass
    var status *CompilerStatus
    if status, err = CompilePass(file, First, nil); err != nil {
        return nil, err
    }
    // Second pass
    file.Seek(0, 0) // rewind
    if status, err = CompilePass(file, Second, status.labels); err != nil {
        return nil, err
    }

    program1 := unsafe.Slice((*int)(unsafe.Pointer(&status.code[0])), len(status.code))
    fmt.Println(program1)

    return status.code, nil
}
