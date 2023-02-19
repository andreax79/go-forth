package assembler

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type tokenTest struct {
	Type   Type
	Symbol string
	Line   int
}

// Create a file from the given source and execute the lexer on it
func runLexer(source string) (*Lexer, error) {
	var err error
	var tmpDir string
	var asmFilename string
	// Create temp directory
	tmpDir, err = os.MkdirTemp("", "test")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir) // clean up
	asmFilename = filepath.Join(tmpDir, "source.pal")
	// Write asm source file
	if err = os.WriteFile(asmFilename, []byte(source), 0666); err != nil {
		return nil, err
	}
	file, err := os.Open(asmFilename)
	if err != nil {
		return nil, err
	}
	// defer file.Close()
	return NewLexer(file), nil
}

// Check lexer output
func testLexer(t *testing.T, lexer *Lexer, tests []tokenTest) {
	for _, test := range tests {
		token, err := lexer.NextToken()
		if err != nil {
			t.Fatalf("%s", err)
		}
		if token.Symbol != test.Symbol {
			t.Fatalf("expected: \"%s\" got: \"%s\"", test.Symbol, token.Symbol)
		}
		if token.Type != test.Type {
			t.Fatalf("\"%s\" - expected: %d got: %d", token.Symbol, test.Type, token.Type)
		}
		if token.Line != test.Line {
			t.Fatalf("\"%s\" - expected: %d got: %d", token.Symbol, test.Line, token.Line)
		}
	}
}

func TestEof(t *testing.T) {
	lexer, err := runLexer(`
start:
    100 $y store
    10 $i store
    $y load
    $i load
    max
    print call
    hlt

print:
    emit
    ret

print0:
    push 2 add
    emit
    ret

square:
    mul
    dup
    print call
    ret
`)
	if err != nil {
		t.Fatalf("%s", err)
	}
	for {
		_, err := lexer.NextToken()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("%s", err)
		}
	}
}

func TestComment(t *testing.T) {
	lexer, err := runLexer(`
; comment 1
# comment 2`)
	if err != nil {
		t.Fatalf("%s", err)
	}

	_, err = lexer.NextToken()
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected end of file")
	}
}

func TestLabel(t *testing.T) {
	lexer, err := runLexer(`
label1:
label2:
`)

	tests := []tokenTest{
		{LABEL, "LABEL1", 1},
		{LABEL, "LABEL2", 2},
	}

	if err != nil {
		t.Fatalf("%s", err)
	}
	testLexer(t, lexer, tests)
}

func TestString(t *testing.T) {
	lexer, err := runLexer(`
label1: .asciz "String"
label2: .asciz "tab\ttab"
label3: .asciz "multiline \
string"
`)

	tests := []tokenTest{
		{LABEL, "LABEL1", 1}, {DIRECTIVE, ".ASCIZ", 1}, {STRING, "String", 1},
		{LABEL, "LABEL2", 2}, {DIRECTIVE, ".ASCIZ", 2}, {STRING, "tab\ttab", 2},
		{LABEL, "LABEL3", 3}, {DIRECTIVE, ".ASCIZ", 3}, {STRING, "multiline string", 3},
	}

	if err != nil {
		t.Fatalf("%s", err)
	}
	testLexer(t, lexer, tests)
}

func TestNumbers(t *testing.T) {
	lexer, err := runLexer(`
    1 1000 ADD
    0x1a -1 MUL
    PUSH 0xabcd
    PUSH 000
`)

	tests := []tokenTest{
		{NUMBER, "1", 1}, {NUMBER, "1000", 1}, {INSTRUCTION, "ADD", 1},
		{NUMBER, "0x1a", 2}, {NUMBER, "-1", 2}, {INSTRUCTION, "MUL", 2},
		{INSTRUCTION, "PUSH", 3}, {NUMBER, "0xabcd", 3},
		{INSTRUCTION, "PUSH", 4}, {NUMBER, "000", 4},
	}

	if err != nil {
		t.Fatalf("%s", err)
	}
	testLexer(t, lexer, tests)
}

func TestIllegalString(t *testing.T) {
	lexer, err := runLexer(`
.asciz "Open\
`)

	lexer.NextToken()
	_, err = lexer.NextToken()
	if err == nil {
		t.Fatalf("error expected")
	}

}
