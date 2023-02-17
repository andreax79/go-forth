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
	"PICK":  "pick",
	"ROLL":  "roll",
	"DEPTH": "depth",
	"NIP":   "swap drop",               // ( x1 x2 -- x2 ) Drop the first item below the top of stack.
	"2DUP":  "over over",               // ( x1 x2 -- x1 x2 x1 x2 ) Duplicate cell pair x1 x2.
	"2DROP": "drop drop",               // ( x1 x2 -- ) Drop cell pair x1 x2 from the stack.
	"2OVER": "push 3 pick push 3 pick", // ( x1 x2 x3 x4 -- x1 x2 x3 x4 x1 x2 ) Copy cell pair x1 x2 to the top of the stack.
	"2SWAP": "rot to_r rot r_from",     // ( x1 x2 x3 x4 -- x3 x4 x1 x2 ) Exchange the top two cell pairs.

	/* Return Stack manipulation */
	">R": "to_r",    // ( x -- ) ( R: -- x ) Move x to the return stack.
	"R>": "r_from",  // ( -- x ) ( R: x -- ) Move x from the return stack to the data stack.
	"R@": "r_fetch", // ( -- x ) ( R: x -- x ) Copy x from the return stack to the data stack.

	/* Arithmetic */
	"+":      "add",
	"-":      "sub",
	"*":      "mul",
	"/":      "div",
	"1+":     "push 1 add",
	"1-":     "push 1 sub",
	"2+":     "push 2 add",
	"2-":     "push 2 sub",
	"MAX":    "max",
	"MIN":    "min",
	"ABS":    "abs",
	"MOD":    "mod",
	"NEGATE": "push 0 swap sub", // Negate n1, giving its arithmetic inverse n2.

	/* Logical */
	"AND":    "and",
	"OR":     "or",
	"XOR":    "xor",
	"NOT":    "not",
	"INVERT": "not",     // ( x1 -- x2 ) Invert all bits of x1, giving its logical inverse x2.
	"TRUE":   "push -1", // ( -- true ) Return a true flag, a value with all bits set (-1).
	"FALSE":  "push 0",  // ( -- false ) Return a false flag (0).

	/* Comparison */
	"=":  "=",
	"<>": "<>",
	">":  ">",
	">=": ">=",
	"<":  "<",
	"<=": "<=",
	"0<": "push 0 <", // ( n -- flag ) flag is true if and only if n is less than zero.
	"0=": "push 0 =", // ( x -- flag ) flag is true if and only if x is equal to zero.
	"0>": "push 0 >", // ( n -- flag ) flag is true if and only if n is greater than zero.

	/* Misc */
	"!":    "store", // ( x a-addr -- ) Store x at a-addr.
	"@":    "load",  // ( a-addr -- x ) x is the value stored at a-addr.
	"EMIT": "emit",
	".":    "period",
	"HLT":  "hlt",
	"NOP":  "nop",
}

type Pass uint8

const (
	First  Pass = 1
	Second      = 2
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
	output  *os.File
	labels  map[string]bool
	pass    Pass // pass number (First/Second)
	context *ContextStack
}

func NewCompilerStatus(pass Pass, output *os.File, labels map[string]bool) (status *CompilerStatus) {
	status = new(CompilerStatus)
	status.pass = pass
	status.output = output
	status.context = new(ContextStack)
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

		pseudo, isPseudo := Pseudo[token]
		_, isLabel := status.labels[token]

		/*		if strings.HasPrefix(token, "@") { // Variable
				if status.pass == Second {
					status.output.WriteString(strings.ToLower(fmt.Sprintf("  %s", token)))
				}

			} else*/

		switch {
		case strings.HasSuffix(token, ":"): // Define a symbol with the value of the current location counter (used to define labels)
			label := strings.TrimSuffix(token, ":")
			status.labels[label] = true
			if status.pass == Second {
				status.output.WriteString(fmt.Sprintf("\n%s:", label))
			}

		case token == "IF":
			status.context.Enter(If)
			if status.pass == Second {
				status.output.WriteString(fmt.Sprintf("\n  not push if_%d_else jcc\n", status.context.Id()))
			}

		case token == "ELSE":
			if !status.context.Is(If) {
				return NewCompilerError("Unbalanced control structure 'else'")
			}
			if status.pass == Second {
				status.output.WriteString(fmt.Sprintf("\n  push if_%d_then jmp\n", status.context.Id()))
				status.output.WriteString(fmt.Sprintf("if_%d_else:\n", status.context.Id()))
			}
			status.context.Change(Else)

		case token == "THEN":
			if !status.context.Is(If) && !status.context.Is(Else) {
				return NewCompilerError("Unbalanced control structure 'then'")
			}
			if status.pass == Second {
				if status.context.Is(If) {
					status.output.WriteString(fmt.Sprintf("\nif_%d_else:\n", status.context.Id()))
				} else {
					status.output.WriteString(fmt.Sprintf("\nif_%d_then:\n", status.context.Id()))
				}
			}
			status.context.Exit()

		case isPseudo:
			if status.pass == Second {
				status.output.WriteString(fmt.Sprintf("  %s", pseudo))
			}

		case isLabel:
			if status.pass == Second {
				status.output.WriteString(fmt.Sprintf("  push %s", token))
			}

		default:
			// Ignore undefined labels/words during the first compilation pass
			if status.pass == Second {
				value, err := strconv.ParseInt(token, 0, 0)
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
