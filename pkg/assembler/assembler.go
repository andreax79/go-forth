package assembler

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	fcpu "github.com/andreax79/go-fcpu/pkg/fcpu"
	"io"
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
	"EQ": fcpu.EQ,
	"NE": fcpu.NE,
	"GT": fcpu.GT,
	"GE": fcpu.GE,
	"LT": fcpu.LT,
	"LE": fcpu.LE,

	/* Control and subroutines */
	"JNZ":  fcpu.JNZ, // jump if not zero
	"JZ":   fcpu.JZ,  // jump if zero
	"JMP":  fcpu.JMP, // jump
	"CALL": fcpu.CALL,
	"RET":  fcpu.RET,

	/* Memory */
	"STORE":   fcpu.STORE,
	"FETCH":   fcpu.FETCH,
	"FETCH_B": fcpu.FETCH_B,
}

type Pass uint8

const (
	First  Pass = 1
	Second      = 2
)

type Directive uint8

const (
	None Directive = iota
	Word
	Byte
	Asciz
	Ascii
)

const TextSegment fcpu.Addr = 0x8048100
const DataSegment fcpu.Addr = 0x8074000

type Segment struct {
	start fcpu.Addr
	addr  fcpu.Addr
	buf   *bytes.Buffer
}

// Compiler status
type CompilerStatus struct {
	text    Segment              // text segment
	data    Segment              // data segment
	segment *Segment             // current segment
	labels  map[string]fcpu.Addr // map label names to addresses
	pass    Pass                 // pass number (First/Second)
}

func NewCompilerStatus(pass Pass, labels map[string]fcpu.Addr) (status *CompilerStatus) {
	status = new(CompilerStatus)
	// Text segment
	status.text.start = TextSegment
	status.text.addr = status.text.start
	status.text.buf = new(bytes.Buffer)
	status.segment = &status.text
	// Data segment
	status.data.start = DataSegment
	status.data.addr = status.data.start
	status.data.buf = new(bytes.Buffer)
	status.pass = pass
	if labels != nil {
		status.labels = labels
	} else {
		status.labels = map[string]fcpu.Addr{}
	}
	return status
}

// Add data to the program
func (status *CompilerStatus) AddData(data fcpu.Word) error {
	if status.pass == Second {
		fmt.Printf("%04x %x\n", status.segment.addr, uint32(data))
	}
	err := binary.Write(status.segment.buf, binary.LittleEndian, data)
	if err != nil {
		return err
	}
	status.segment.addr += fcpu.WordSize
	return nil
}

// Add data to the program
func (status *CompilerStatus) AddBytes(bytes []byte) error {
	if status.pass == Second {
		fmt.Printf("%04x %v\n", status.segment.addr, bytes)
	}
	status.segment.buf.Write(bytes)
	status.segment.addr += fcpu.Addr(len(bytes))
	return nil
}

// Add compiled code to the program
func (status *CompilerStatus) AddCode(code ...fcpu.Word) error {
	if status.pass == Second {
		fmt.Printf("%04x %s\n", status.segment.addr, strings.Trim(fmt.Sprint(code), "[]"))
	}
	err := binary.Write(status.segment.buf, binary.LittleEndian, code)
	if err != nil {
		return err
	}
	status.segment.addr += fcpu.Addr(len(code)) * fcpu.WordSize
	return nil
}

// Execute a compilation pass
// Each source line contains some combination of the following fields:
// label:    instructions/operands      ; comment
func CompilePass(file *os.File, pass Pass, labels map[string]fcpu.Addr) (*CompilerStatus, error) {
	status := NewCompilerStatus(pass, labels)
	lexer := NewLexer(file)
	directive := None
	for {
		token, err := lexer.NextToken()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return status, nil
			}
			return nil, err
		}

		switch token.Type {
		case INSTRUCTION:
			err = status.AddCode(Instructions[token.Symbol])
			directive = None
		case DIRECTIVE:
			switch token.Symbol {
			case ".TEXT": // changes the current segment to text
				status.segment = &status.text
				directive = None
			case ".DATA": // change the current segment to data
				status.segment = &status.data
				directive = None
			case ".WORD":
				directive = Word
			case ".BYTE":
				directive = Byte
			case ".ASCIZ":
				directive = Asciz
			case ".ASCII":
				directive = Ascii
			default:
				return nil, &UndefinedDirective{Label: token.Symbol, Line: token.Line}
			}

		case IDENTIFIER:
			// Ignore undefined identifier during the first compilation pass
			if status.pass != First {
				label, exists := status.labels[token.Symbol]
				if !exists {
					return nil, &UndefinedSymbol{Label: token.Symbol, Line: token.Line}
				}
				err = status.AddData(fcpu.Word(label))
			} else {
				err = status.AddData(fcpu.Word(-1))
			}
			directive = None

		case LABEL:
			if _, exists := status.labels[token.Symbol]; exists && status.pass == First {
				return nil, &LabelMultipleDefinition{Label: token.Symbol, Line: token.Line}
			}
			status.labels[token.Symbol] = status.segment.addr
			directive = None

		case NUMBER:
			value, err := strconv.ParseInt(token.Symbol, 0, 0)
			if err != nil {
				return nil, err
			}
			switch directive {
			case Byte:
				err = status.AddBytes([]byte{byte(value)})
			case None:
				fallthrough
			case Word:
				err = status.AddData(fcpu.Word(value))
			default:
				return nil, &UnexpectedToken{Token: token.Symbol, Line: token.Line}
			}

		case STRING:
			switch directive {
			case Asciz:
				err = status.AddBytes([]byte(token.Symbol + string(rune(0))))
			case Ascii:
				err = status.AddBytes([]byte(token.Symbol))
			default:
				return nil, &UnexpectedToken{Token: token.Symbol, Line: token.Line}
			}
		}
	}
	return status, nil
}

func WriteBinary(status *CompilerStatus, outputFilename string) error {
	output, err := os.Create(outputFilename)
	defer output.Close()
	if err != nil {
		return err
	}
	// Prepare header
	var header fcpu.BinaryHeader
	header.Magic = fcpu.BinaryMagic
	header.TextSize = fcpu.Addr(status.text.buf.Len())
	header.DataSize = fcpu.Addr(status.data.buf.Len())
	header.TextBase = status.text.start
	header.DataBase = status.data.start
	// Write header
	err = binary.Write(output, binary.LittleEndian, header)
	if err != nil {
		return err
	}
	// Write text
	_, err = output.Write(status.text.buf.Bytes())
	if err != nil {
		return err
	}
	// Write data
	_, err = output.Write(status.data.buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

// Compile a program file and return the compiled code
func Compile(filename string, outputFilename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// First pass
	var status *CompilerStatus
	if status, err = CompilePass(file, First, nil); err != nil {
		return err
	}
	// Second pass
	file.Seek(0, 0) // rewind
	if status, err = CompilePass(file, Second, status.labels); err != nil {
		return err
	}
	// Write output
	return WriteBinary(status, outputFilename)
}
