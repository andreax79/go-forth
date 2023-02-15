package fcpu

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Pseudo instructions
var Pseudo = map[string]string{
	/* Stack manipulation */
	"DUP":   "dup",
	"?DUP":  "cdup",
	"DROP":  "drop",
	"SWAP":  "swap",
	"OVER":  "over",
	"ROT":   "rot",
	"DEPTH": "depth",
	"2DUP":  "over over", // Duplicate cell pair x1 x2.
	"2DROP": "drop drop", // Drop cell pair x1 x2 from the stack.
	"NIP":   "swap drop", // Drop the first item below the top of stack.

	/* Arithmetic */
	"+":   "add",
	"-":   "sub",
	"*":   "mul",
	"/":   "div",
	"1+":  "push 1 add",
	"1-":  "push 1 sub",
	"2+":  "push 2 add",
	"2-":  "push 2 sub",
	"MAX": "max",
	"MIN": "min",
	"ABS": "abs",
	"MOD": "mod",

	/* Logical */
	"AND":   "and",
	"OR":    "or",
	"XOR":   "xor",
	"NOT":   "not",
	"TRUE":  "push -1",
	"FALSE": "push 0",

	/* Comparison */
	"=":  "=",
	"<>": "<>",
	">":  ">",
	">=": ">=",
	"<":  "<",
	"<=": "<=",
	"0<": "push 0 <",
	"0=": "push 0 =",
	"0>": "push 0 >",

	/* Misc */
	"EMIT": "emit",
	".":    "period",
	"HLT":  "hlt",
	"NOP":  "nop",
}

// TODO:
// SP@ - Push the current data stack pointer
// SP! - Set the data stack pointer
// SP0 - Pointer to the bottom of the data stack
// RP@ - Push the current return stack pointer
// RP! - Set the return stack pointer
// RP0 - Pointer to the bottom of the return stack

type Pass uint8

const (
	First  Pass = 1
	Second      = 2
)

type Phase uint8

const (
	None Phase = iota
	If
	Else
	Then
)

type CompilerError struct {
	message string
}

func NewCompilerError(message string) *CompilerError {
	err := new(CompilerError)
	err.message = message
	return err
}

func (e *CompilerError) Error() string {
	return fmt.Sprintf("Compiler error: %s", e.message)
}

// Compiler status
type CompilerStatus struct {
	output      *os.File
	labels      map[string]bool
	pass        Pass // pass number (First/Second)
	ifStructure int
	phase       Phase
}

func NewCompilerStatus(pass Pass, output *os.File, labels map[string]bool) (status *CompilerStatus) {
	status = new(CompilerStatus)
	status.pass = pass
	status.output = output
	status.ifStructure = 0
	status.phase = None
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
		if strings.HasPrefix(token, "\\") { // Start of comment. The rest of the current line is ignored.
			break
		}

		/*		if strings.HasPrefix(token, "@") { // Variable
				if status.pass == Second {
					status.output.WriteString(strings.ToLower(fmt.Sprintf("  %s", token)))
				}

			} else*/
		if strings.HasSuffix(token, ":") { // Define a symbol with the value of the current location counter (used to define labels)
			label := strings.TrimSuffix(token, ":")
			status.labels[label] = true
			if status.pass == Second {
				status.output.WriteString(fmt.Sprintf("\n%s:", label))
			}

		} else if token == "IF" {
			status.ifStructure++
			status.phase = If
			if status.pass == Second {
				status.output.WriteString(fmt.Sprintf("\n  not push if_%d_else jcc\n", status.ifStructure))
			}

		} else if token == "ELSE" {
			if status.phase != If {
				return NewCompilerError("Unbalanced control structure 'else'")
			}
			if status.pass == Second {
				status.output.WriteString(fmt.Sprintf("\n  push if_%d_then jmp\n", status.ifStructure))
				status.output.WriteString(fmt.Sprintf("if_%d_else:\n", status.ifStructure))
			}
			status.phase = Else

		} else if token == "THEN" {
			if status.phase != If && status.phase != Else {
				return NewCompilerError("Unbalanced control structure 'then'")
			}
			if status.pass == Second {
				fmt.Println("phase", status.phase)
				if status.phase == If {
					status.output.WriteString(fmt.Sprintf("\nif_%d_else:\n", status.ifStructure))
				} else {
					status.output.WriteString(fmt.Sprintf("\nif_%d_then:\n", status.ifStructure))
				}
			}
			status.phase = None

		} else if pseudo, exists := Pseudo[token]; exists { // Token
			if status.pass == Second {
				status.output.WriteString(fmt.Sprintf("  %s", pseudo))
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
					return NewCompilerError(fmt.Sprintf("%s ?", strings.ToLower(token)))
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
