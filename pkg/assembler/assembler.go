package fcpu

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	fcpu "github.com/andreax79/go-fcpu/pkg/fcpu"
	"os"
	"strconv"
	"strings"
)

var Instructions = map[string]fcpu.Word{
	"HLT":    fcpu.HLT,
	"NOP":    fcpu.NOP,
	"EMIT":   fcpu.EMIT,
	"PERIOD": fcpu.PERIOD,

	/* Stack manipulation */
	"PUSH":  fcpu.PUSH,
	"ZERO":  fcpu.ZERO,
	"DUP":   fcpu.DUP,
	"?DUP":  fcpu.CDUP,
	"DROP":  fcpu.DROP,
	"SWAP":  fcpu.SWAP,
	"OVER":  fcpu.OVER,
	"ROT":   fcpu.ROT,
	"PICK":  fcpu.PICK,
	"ROLL":  fcpu.ROLL,
	"DEPTH": fcpu.DEPTH,

	/* Return Stack manipulation */
	"TO_R":    fcpu.TO_R,
	"R_FROM":  fcpu.R_FROM,
	"R_FETCH": fcpu.R_FETCH,

	/* Arithmetic */
	"ADD": fcpu.ADD,
	"SUB": fcpu.SUB,
	"MUL": fcpu.MUL,
	"DIV": fcpu.DIV,
	"1+":  fcpu.ADD_ONE,
	"1-":  fcpu.SUB_ONE,
	"MAX": fcpu.MAX,
	"MIN": fcpu.MIN,
	"ABS": fcpu.ABS,
	"MOD": fcpu.MOD,

	/* Logical */
	"AND": fcpu.AND,
	"OR":  fcpu.OR,
	"XOR": fcpu.XOR,
	"NOT": fcpu.NOT,

	/* Comparison */
	"=":  fcpu.EQ,
	"<>": fcpu.NOT_EQ,
	">":  fcpu.GREAT,
	">=": fcpu.EQ_GREAT,
	"<":  fcpu.LESS,
	"<=": fcpu.EQ_LESS,

	/* Control and subroutines */
	"JCC":  fcpu.JCC,
	"JMP":  fcpu.JMP,
	"CALL": fcpu.CALL,
	"RET":  fcpu.RET,

	/* Memory */
	"STORE": fcpu.STORE,
	"FETCH": fcpu.FETCH,
}

type Pass uint8

const (
	First  Pass = 1
	Second      = 2
)

// Compiler status
type CompilerStatus struct {
	pc               fcpu.Addr
	buf              *bytes.Buffer        // program code
	labels           map[string]fcpu.Addr // map label names to addresses
	variables        map[string]fcpu.Addr // map variable names to addresses
	lastVariableAddr fcpu.Addr            // address of the last variable
	pass             Pass                 // pass number (First/Second)
}

func NewCompilerStatus(pass Pass, labels map[string]fcpu.Addr) (status *CompilerStatus) {
	status = new(CompilerStatus)
	status.pc = 0
	status.buf = new(bytes.Buffer)
	status.variables = map[string]fcpu.Addr{}
	status.lastVariableAddr = 0
	status.pass = pass
	if labels != nil {
		status.labels = labels
	} else {
		status.labels = map[string]fcpu.Addr{}
	}
	return status
}

// Add compiled code to the program
func (status *CompilerStatus) AddCode(code ...fcpu.Word) error {
	if status.pass == Second {
		if code[0] == fcpu.PUSH && len(code) == 2 {
			fmt.Printf("%04d %s %d\n", status.pc, code[0], int(code[1]))
		} else {
			fmt.Printf("%04d %s\n", status.pc, strings.Trim(fmt.Sprint(code), "[]"))
		}
	}
	err := binary.Write(status.buf, binary.LittleEndian, code)
	if err != nil {
		return err
	}
	status.pc += fcpu.Addr(len(code)) * fcpu.WordSize
	return nil
}

// Compile a line, add compiled code to the program
// Each source line contains some combination of the following fields:
// label:    instructions/operands      ; comment
func CompileLine(status *CompilerStatus, line string) error {
	var err error
	for _, token := range strings.Fields(line) {
		token = strings.ToUpper(token)
		if strings.HasPrefix(token, ";") { // Start of comment. The rest of the current line is ignored.
			break
		}

		if strings.HasPrefix(token, "@") { // Variable
			variable := strings.TrimPrefix(token, "@")
			addr, exists := status.variables[variable]
			if !exists { // New variable
				addr = status.lastVariableAddr
				status.lastVariableAddr += fcpu.WordSize
				status.variables[variable] = addr
				if err = status.AddCode(fcpu.INC_RSP); err != nil { // Increment rsp
					return err
				}
			}
			err = status.AddCode(fcpu.PUSH, fcpu.Word(addr))

		} else if strings.HasSuffix(token, ":") { // Define a symbol with the value of the current location counter (used to define labels)
			label := strings.TrimSuffix(token, ":")
			status.labels[label] = status.pc

		} else if c, exists := Instructions[token]; exists { // Instruction
			err = status.AddCode(c)

		} else if c, exists := status.labels[token]; exists { // Label
			err = status.AddCode(fcpu.Word(c))

		} else { // Push
			value, err := strconv.ParseInt(token, 0, 0)
			if err != nil {
				// Ignore undefined labels during the first compilation pass
				if status.pass != First {
					return err
				}
				value = -1 // TODO - test value
			}
			err = status.AddCode(fcpu.Word(value))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Execute a compilation pass
func CompilePass(file *os.File, pass Pass, labels map[string]fcpu.Addr) (*CompilerStatus, error) {
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
