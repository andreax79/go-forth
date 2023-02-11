package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Pseudo instructions
const (
	ADD_TWO Word = iota + 0xffff
	SUB_TWO
	ZERO_LESS
	ZERO_EQ
	ZERO_GREAT
)

var Symbols = map[string]Word{
	"HLT":  HLT,
	"NOP":  NOP,
	"EMIT": EMIT,

	/* Stack manipulation */
	// "PUSH": PUSH,
	"ZERO":  ZERO,
	"DUP":   DUP,
	"?DUP":  CDUP,
	"DROP":  DROP,
	"SWAP":  SWAP,
	"OVER":  OVER,
	"ROT":   ROT,
	"DEPTH": DEPTH,

	/* Arithmetic */
	"+":   ADD,
	"ADD": ADD,
	"-":   SUB,
	"SUB": SUB,
	"*":   MUL,
	"MUL": MUL,
	"/":   DIV,
	"DIV": DIV,
	"1+":  ADD_ONE,
	"1-":  SUB_ONE,
	"2+":  ADD_TWO,
	"2-":  SUB_TWO,
	"MAX": MAX,
	"MIN": MIN,
	"ABS": ABS,
	"MOD": MOD,

	/* Logical */
	"AND": AND,
	"OR":  OR,
	"XOR": XOR,
	"NOT": NOT,

	/* Comparison */
	"=":  EQ,
	"<>": NOT_EQ,
	">":  GREAT,
	">=": EQ_GREAT,
	"<":  LESS,
	"<=": EQ_LESS,
	"0<": ZERO_LESS,
	"0=": ZERO_EQ,
	"0>": ZERO_GREAT,

	/* Control and subroutines */
	"JMPC": JMPC,
	"CALL": CALL,
	"RET":  RET,

	/* Memory */
	"STORE": STORE,
	"LOAD":  LOAD,
}

type Pass uint8

const (
	First  Pass = 1
	Second      = 2
)

// Compiler status
type CompilerStatus struct {
	pc               Addr
	buf              *bytes.Buffer   // program code
	labels           map[string]Addr // map label names to addresses
	variables        map[string]Addr // map variable names to addresses
	lastVariableAddr Addr            // address of the last variable
	pass             Pass            // pass number (First/Second)
}

func NewCompilerStatus(pass Pass, labels map[string]Addr) (status *CompilerStatus) {
	status = new(CompilerStatus)
	status.pc = 0
	status.buf = new(bytes.Buffer)
	status.variables = map[string]Addr{}
	status.lastVariableAddr = 0
	status.pass = pass
	if labels != nil {
		status.labels = labels
	} else {
		status.labels = map[string]Addr{}
	}
	return status
}

// Add compiled code to the program
func (status *CompilerStatus) AddCode(code ...Word) error {
	if status.pass == Second {
		if code[0] == PUSH && len(code) == 2 {
			fmt.Printf("%04d %s %d\n", status.pc, code[0], int(code[1]))
		} else {
			fmt.Printf("%04d %s\n", status.pc, strings.Trim(fmt.Sprint(code), "[]"))
		}
	}
	err := binary.Write(status.buf, binary.LittleEndian, code)
	if err != nil {
		return err
	}
	status.pc += Addr(len(code)) * WordSize
	return nil
}

// Compile a line, add compiled code to the program
func CompileLine(status *CompilerStatus, line string) error {
	var err error
	for _, token := range strings.Fields(line) {
		token = strings.ToUpper(token)
		if strings.HasPrefix(token, "/") { // Start of comment. The rest of the current line is ignored.
			break
		}

		if strings.HasPrefix(token, "@") { // Variable
			variable := strings.TrimPrefix(token, "@")
			addr, exists := status.variables[variable]
			if !exists { // New variable
				addr = status.lastVariableAddr
				status.lastVariableAddr += WordSize
				status.variables[variable] = addr
				if err = status.AddCode(INC_RSP); err != nil { // Increment rsp
					return err
				}
			}
			err = status.AddCode(PUSH, Word(addr))

		} else if strings.HasSuffix(token, ":") { // Define a symbol with the value of the current location counter (used to define labels)
			label := strings.TrimSuffix(token, ":")
			status.labels[label] = status.pc

		} else if c, exists := Symbols[token]; exists { // Token
			switch c {
			case ADD_TWO:
				err = status.AddCode(ADD_ONE, ADD_ONE)
			case SUB_TWO:
				err = status.AddCode(SUB_ONE, SUB_ONE)
			case ZERO_LESS:
				err = status.AddCode(PUSH, 0, LESS)
			case ZERO_EQ:
				err = status.AddCode(PUSH, 0, EQ)
			case ZERO_GREAT:
				err = status.AddCode(PUSH, 0, GREAT)
			default:
				err = status.AddCode(c)
			}
		} else if c, exists := status.labels[token]; exists { // Label
			err = status.AddCode(PUSH, Word(c))

		} else { // Push
			value, err := strconv.Atoi(token)
			if err != nil {
				// Ignore undefined labels during the first compilation pass
				if status.pass != First {
					return err
				}
				value = -1 // TODO - test value
			}
			err = status.AddCode(PUSH, Word(value))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Execute a compilation pass
func CompilePass(file *os.File, pass Pass, labels map[string]Addr) (*CompilerStatus, error) {
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
func Compile(filename string) ([]byte, error) {
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
	return status.buf.Bytes(), nil
}
