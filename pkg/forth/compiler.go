package fcpu

import (
	"bufio"
	"fmt"
	fcpu "github.com/andreax79/go-fcpu/pkg/fcpu"
	"os"
	"strconv"
	"strings"
)

// Pseudo instructions
var Pseudo = map[string]string{
	"1+":    "push 1 add",
	"1-":    "push 1 sub",
	"2+":    "push 2 add",
	"2-":    "push 2 sub",
	"0<":    "push 0 <",
	"0=":    "push 0 =",
	"0>":    "push 0 >",
	"TRUE":  "push 1",
	"FALSE": "push 0",
}

// TODO:
// SP@ - Push the current data stack pointer
// SP! - Set the data stack pointer
// SP0 - Pointer to the bottom of the data stack
// RP@ - Push the current return stack pointer
// RP! - Set the return stack pointer
// RP0 - Pointer to the bottom of the return stack

var ForthSymbols = map[string]fcpu.Word{
	"HLT":  fcpu.HLT,
	"NOP":  fcpu.NOP,
	"EMIT": fcpu.EMIT,

	/* Stack manipulation */
	"DUP":   fcpu.DUP,
	"?DUP":  fcpu.CDUP,
	"DROP":  fcpu.DROP,
	"SWAP":  fcpu.SWAP,
	"OVER":  fcpu.OVER,
	"ROT":   fcpu.ROT,
	"DEPTH": fcpu.DEPTH,

	/* Arithmetic */
	"+":   fcpu.ADD,
	"-":   fcpu.SUB,
	"*":   fcpu.MUL,
	"/":   fcpu.DIV,
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
	"JMPC": fcpu.JMPC,
	"CALL": fcpu.CALL,
	"RET":  fcpu.RET,

	/* Memory */
	"STORE": fcpu.STORE,
	"LOAD":  fcpu.LOAD,
}

type Pass uint8

const (
	First  Pass = 1
	Second      = 2
)

// Compiler status
type CompilerStatus struct {
	output *os.File
	labels map[string]bool
	pass   Pass // pass number (First/Second)
}

func NewCompilerStatus(pass Pass, output *os.File, labels map[string]bool) (status *CompilerStatus) {
	status = new(CompilerStatus)
	status.pass = pass
	status.output = output
	if labels != nil {
		status.labels = labels
	} else {
		status.labels = map[string]bool{}
	}
	return status
}

// Compile a line, add compiled code to the program
func CompileLine(status *CompilerStatus, line string) error {
	var err error
	for _, token := range strings.Fields(line) {
		token = strings.ToUpper(token)
		// if strings.HasPrefix(token, "/") { // Start of comment. The rest of the current line is ignored.
		// 	break
		// }

		if strings.HasPrefix(token, "@") { // Variable
			if status.pass == Second {
				status.output.WriteString(strings.ToLower(fmt.Sprintf("  %s", token)))
			}

		} else if strings.HasSuffix(token, ":") { // Define a symbol with the value of the current location counter (used to define labels)
			label := strings.TrimSuffix(token, ":")
			status.labels[label] = true
			if status.pass == Second {
				status.output.WriteString(fmt.Sprintf("\n%s:", label))
			}

		} else if pseudo, exists := Pseudo[token]; exists { // Token
			if status.pass == Second {
				status.output.WriteString(fmt.Sprintf("  %s", pseudo))
			}

		} else if c, exists := ForthSymbols[token]; exists { // Token
			if status.pass == Second {
				status.output.WriteString(strings.ToLower(fmt.Sprintf("  %s", c)))
			}

		} else if _, exists := status.labels[token]; exists { // Label
			if status.pass == Second {
				status.output.WriteString(fmt.Sprintf("  push %s", token))
			}

		} else { // Push
			// Ignore undefined labels during the first compilation pass
			if status.pass == Second {
				value, err := strconv.Atoi(token)
				if err != nil {
					return err
				}
				status.output.WriteString(fmt.Sprintf("  push %d", value))
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Execute a compilation pass
func CompilePass(input *os.File, output *os.File, pass Pass, labels map[string]bool) (*CompilerStatus, error) {
	status := NewCompilerStatus(pass, output, labels)
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		if err := CompileLine(status, line); err != nil {
			return nil, err
		}
		if status.pass == Second {
			status.output.WriteString("\n")
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return status, nil
}

// Compile a program file and return the compiled code
func Compile(filename string) (string, error) {
	outputFilename := fmt.Sprintf("%s.pal", filename)
	input, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer input.Close()
	output, err := os.Create(outputFilename)
	if err != nil {
		return "", err
	}
	defer output.Close()

	// First pass
	var status *CompilerStatus
	if status, err = CompilePass(input, output, First, nil); err != nil {
		return "", err
	}
	// Second pass
	input.Seek(0, 0) // rewind
	if status, err = CompilePass(input, output, Second, status.labels); err != nil {
		return "", err
	}
	return outputFilename, nil
}
