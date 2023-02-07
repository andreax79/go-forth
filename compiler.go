package main

import (
    "os"
    "fmt"
    "unsafe"
    "bufio"
    "strings"
    "strconv"
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
    "ADD": ADD,
    "SUB": SUB,
    "MUL": DIV,
    "INC": INC,
    "DEC": DEC,
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

type CompilerStatus struct {
    pc int
    program []Word
    labels map[string]int
    variables map[string]int
    lastVariableAddr int
}

func NewCompilerStatus() (status *CompilerStatus) {
    status = new(CompilerStatus)
    status.pc = 0
    status.program = make([]Word, 0)
    status.labels = map[string]int{}
    status.variables = map[string]int{}
    status.lastVariableAddr = 0
    return status
}

func (status *CompilerStatus) AddCode(code ...Word) {
    if code[0] == PUSH && len(code) == 2 {
        fmt.Printf("%04d %s %d\n", status.pc, code[0], int(code[1]))
    } else {
        fmt.Printf("%04d %s\n", status.pc, strings.Trim(fmt.Sprint(code), "[]"))
    }
    status.program = append(status.program, code...)
    status.pc += len(code)
}

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
            status.AddCode(Word(c))

        } else if c, exists := status.labels[token]; exists { // Label
            status.AddCode(PUSH, Word(c))

        } else { // Push
            value, err := strconv.Atoi(token)
            if err != nil {
                fmt.Println(token, err)
                return err
            }
            status.AddCode(PUSH, Word(value))
        }
    }
    return nil
}


func Compile(filename string) ([]Word, error) {
    var status = NewCompilerStatus()
    file, err := os.Open(filename)
    if err != nil {
        fmt.Println(err)
        return nil, err
    }
    defer file.Close()

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

    program1 := unsafe.Slice((*int)(unsafe.Pointer(&status.program[0])), len(status.program))
    fmt.Println(program1)

    if err := scanner.Err(); err != nil {
        fmt.Println(err)
    }
    return status.program, nil
}
