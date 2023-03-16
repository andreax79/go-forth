package forth

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var Definitions = map[string]string{
	/* Stack manipulation */
	"DUP":   ";code dup ;",
	"DROP":  ";code drop ;",
	"SWAP":  ";code swap ;",
	"OVER":  ";code over ;",
	"ROT":   ";code rot ;",
	"PICK":  ";code pick ;",
	"ROLL":  ";code roll ;",
	"DEPTH": ";code depth ;",
	"NIP":   "swap drop",     // ( x1 x2 -- x2 ) Drop the first item below the top of stack.
	"2DUP":  "over over",     // ( x1 x2 -- x1 x2 x1 x2 ) Duplicate cell pair x1 x2.
	"2DROP": "drop drop",     // ( x1 x2 -- ) Drop cell pair x1 x2 from the stack.
	"2OVER": "3 pick 3 pick", // ( x1 x2 x3 x4 -- x1 x2 x3 x4 x1 x2 ) Copy cell pair x1 x2 to the top of the stack.
	"2SWAP": "rot >r rot r>", // ( x1 x2 x3 x4 -- x3 x4 x1 x2 ) Exchange the top two cell pairs.

	/* Return Stack manipulation */
	">R": ";code to_r ;",    // ( x -- ) ( R: -- x ) Move x to the return stack.
	"R>": ";code r_from ;",  // ( -- x ) ( R: x -- ) Move x from the return stack to the data stack.
	"R@": ";code r_fetch ;", // ( -- x ) ( R: x -- x ) Copy x from the return stack to the data stack.

	/* Arithmetic */
	"+":      ";code add ;",
	"-":      ";code sub ;",
	"*":      ";code mul ;",
	"/":      ";code div ;",
	"/MOD":   ";code divmod ;",
	"MOD":    ";code mod ;",
	"1+":     "1 +",
	"1-":     "1 -",
	"2+":     "2 +",
	"2-":     "2 -",
	"MAX":    ";code max ;",
	"MIN":    ";code min ;",
	"ABS":    ";code abs ;",
	"LSHIFT": ";code lshift ;", // Perform a logical left shift
	"RSHIFT": ";code rshift ;", // Perform a logical right shift
	"NEGATE": "0 swap -",       // Negate n1, giving its arithmetic inverse n2

	/* Logical */
	"AND":    ";code and ;",
	"OR":     ";code or ;",
	"XOR":    ";code xor ;",
	"NOT":    ";code not ;",
	"INVERT": ";code not ;", // ( x1 -- x2 ) Invert all bits of x1, giving its logical inverse x2.
	"TRUE":   "-1",          // ( -- true ) Return a true flag, a value with all bits set (-1).
	"FALSE":  "0",           // ( -- false ) Return a false flag (0).

	/* Comparison */
	"=":  ";code eq ;",
	"<>": ";code ne ;",
	">":  ";code gt ;",
	">=": ";code ge ;",
	"<":  ";code lt ;",
	"<=": ";code le ;",
	"0<": "0 <", // ( n -- flag ) flag is true if and only if n is less than zero.
	"0=": "0 =", // ( x -- flag ) flag is true if and only if x is equal to zero.
	"0>": "0 >", // ( n -- flag ) flag is true if and only if n is greater than zero.

	/* Misc */
	"!":    ";code store ;", // ( x a-addr -- ) Store x at a-addr.
	"@":    ";code fetch ;", // ( a-addr -- x ) x is the value stored at a-addr.
	"EMIT": ";code emit ;",
	".":    ";code period ;",
	"HLT":  ";code hlt ;",
	"NOP":  ";code nop ;",
	"CALL": ";code call ;",
	"JMP":  ";code jmp ;",
	"RET":  ";code ret ;",
}

type Pass uint8

const (
	First  Pass = 1
	Second      = 2
)

type CompilerError struct {
	message string
}

var Constants = map[string]int{
	"BL": 32, // space
}

func NewCompilerError(message string) *CompilerError {
	err := new(CompilerError)
	err.message = message
	return err
}

func (e *CompilerError) Error() string {
	return fmt.Sprintf("Forth compiler error: %s", e.message)
}

// Compiler status
type CompilerStatus struct {
	output     *os.File
	labels     map[string]bool
	constants  map[string]int
	pass       Pass // pass number (First/Second)
	context    *ContextStack
	buf        strings.Builder
	dictionary map[string]string
}

func NewCompilerStatus(pass Pass, output *os.File, labels map[string]bool, constants map[string]int) (status *CompilerStatus) {
	status = new(CompilerStatus)
	status.pass = pass
	status.output = output
	status.context = new(ContextStack)
	if labels != nil {
		status.labels = labels
	} else {
		status.labels = map[string]bool{}
	}
	if constants != nil {
		status.constants = constants
	} else {
		status.constants = Constants
	}
	status.dictionary = Definitions
	return status
}

func (status *CompilerStatus) Add(s string) {
	s = strings.Replace(s, "{ID}", fmt.Sprint(status.context.Id()), -1)
	status.WriteString("\n" + s + "\n")
}

func (status *CompilerStatus) WriteString(s string) {
	if status.context.HasAnchestor(Colon) {
		status.buf.WriteString(s)
	} else {
		status.output.WriteString(s)
	}
}

// Compile a line, add compiled code to the program
func CompileLine(status *CompilerStatus, line string) error {
	var err error
	fields := strings.Fields(line)
	for i := 0; i < len(fields); i++ {
		token := fields[i]

		if status.context.Is(Paren) { // Parenthesis comments
			t := strings.SplitN(token, ")", 2)
			if len(t) < 2 {
				continue
			}
			status.context.Exit()
			token = t[1]
			if len(token) == 0 {
				continue
			}
		} else if status.context.Is(Code) {
			if token == ";" {
				status.context.Exit()
			} else if status.pass == Second {
				status.WriteString(" " + token)
			}
			continue
		}

		token = strings.ToUpper(token)
		if strings.HasPrefix(token, "\\") { // Start of comment. The rest of the current line is ignored.
			break
		}

		definition, hasDefinition := status.dictionary[token]
		_, isLabel := status.labels[token]
		constantValue, isConstant := status.constants[token]

		switch {
		case token == "IF":
			status.context.Enter(If)
			if status.pass == Second {
				status.Add("  not push if_{ID}_else jnz")
			}

		case token == "ELSE":
			if !status.context.Is(If) {
				return NewCompilerError("Unbalanced control structure 'else'")
			}
			if status.pass == Second {
				status.Add("  push if_{ID}_then jmp")
				status.Add("if_{ID}_else:")
			}
			status.context.Change(Else)

		case token == "THEN":
			if !status.context.Is(If) && !status.context.Is(Else) {
				return NewCompilerError("Unbalanced control structure 'then'")
			}
			if status.pass == Second {
				if status.context.Is(If) {
					status.Add("if_{ID}_else:")
				} else {
					status.Add("if_{ID}_then:")
				}
			}
			status.context.Exit()

		case token == "DO":
			status.context.Enter(Do)
			if status.pass == Second {
				status.Add("  swap to_r to_r") // Push limit, i on the return stack
				status.Add("do_{ID}:")
			}

		case token == "I":
			if !status.context.Is(Do) {
				return NewCompilerError("Unbalanced control structure 'i'")
			}
			if status.pass == Second {
				status.Add("  r_fetch") // Fetch i from the return stack
			}
			break

		case token == "LOOP":
			if !status.context.Is(Do) {
				return NewCompilerError("Unbalanced control structure 'loop'")
			}
			if status.pass == Second {
				status.Add("  r_from r_fetch swap") // Push limit, i
				status.Add("  push 1 add")          // Increment i
				status.Add("  dup to_r")            // Store i on the return stack
				status.Add("  gt push do_{ID} jnz") // Loop
				status.Add("do_{ID}_end:")
				status.Add("  r_from drop r_from drop") // Remove limit, i from the return stack
			}
			status.context.Exit()

		case token == "LEAVE":
			if !status.context.HasAnchestor(Do) {
				return NewCompilerError("Unbalanced control structure 'loop'")
			}
			if status.pass == Second {
				status.Add("  r_from drop r_fetch to_r") // i := limit
				status.Add("  push do_{ID}_end jmp")     // Go to end
			}

		case token == "?DUP":
			status.context.Enter(If)
			if status.pass == Second {
				status.Add("  dup")
				status.Add("  not push if_{ID}_then jnz")
				status.Add("  dup")
				status.Add("if_{ID}_then:")
			}
			status.context.Exit()

		case token == ":": // Colon
			status.context.Enter(Colon)
			i++
			if i >= len(fields) {
				return NewCompilerError("missing colon definition")
			}
			label := strings.ToUpper(fields[i])
			status.labels[label+"_COL"] = true
			status.dictionary[label] = fmt.Sprintf("%s_col call", strings.ToLower(label))
			if status.pass == Second {
				status.Add(fmt.Sprintf("%s_col:", strings.ToLower(label)))
			}

		case token == ";": // Semicolon
			if !status.context.HasAnchestor(Colon) {
				return NewCompilerError("not a function")
			}
			status.Add("  ret")
			status.context.Exit()

		case token == ";CODE": // Code
			status.context.Enter(Code)

		case token == "(": // Paren
			status.context.Enter(Paren)

		case isConstant:
			if status.pass == Second {
				status.WriteString(fmt.Sprintf("  push %d", constantValue))
			}

		case isLabel:
			if status.pass == Second {
				status.WriteString(fmt.Sprintf("  push %s", token))
			}

		case strings.HasSuffix(token, ":"): // Define a symbol with the value of the current location counter (used to define labels)
			label := strings.TrimSuffix(token, ":")
			status.labels[label] = true
			if status.pass == Second {
				status.WriteString(fmt.Sprintf("\n%s:", label))
			}

		case hasDefinition:
			if status.pass == Second {
				err = CompileLine(status, definition)
				if err != nil {
					return err
				}
			}

		default:
			// Ignore undefined labels/words during the first compilation pass
			value, err := strconv.ParseInt(token, 0, 0)
			if i+2 < len(fields) && strings.ToUpper(fields[i+1]) == "CONSTANT" {
				if status.pass == First {
					label := strings.ToUpper(fields[i+2])
					status.constants[label] = int(value)
				}
				i = i + 2
			} else if status.pass == Second {
				if err != nil {
					return NewCompilerError(fmt.Sprintf("%s ?", strings.ToLower(token)))
				}
				status.WriteString(fmt.Sprintf("  push %d", value))
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Execute a compilation pass
func CompilePass(input *os.File, output *os.File, pass Pass, labels map[string]bool, constants map[string]int) (*CompilerStatus, error) {
	status := NewCompilerStatus(pass, output, labels, constants)
	scanner := bufio.NewScanner(input)
	if status.pass == Second {
		status.output.WriteString("start:\n")
	}
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
	if status.pass == Second {
		status.output.WriteString(status.buf.String())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return status, nil
}

// Compile a program file and return the compiled code
func Compile(filename string, outputFilename string) error {
	input, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer output.Close()

	// First pass
	var status *CompilerStatus
	if status, err = CompilePass(input, output, First, nil, nil); err != nil {
		return err
	}
	// Second pass
	input.Seek(0, 0) // rewind
	if status, err = CompilePass(input, output, Second, status.labels, status.constants); err != nil {
		return err
	}
	return nil
}
