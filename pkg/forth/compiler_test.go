package forth

import (
	"errors"
	"fmt"
	asm "github.com/andreax79/go-fcpu/pkg/assembler"
	fcpu "github.com/andreax79/go-fcpu/pkg/fcpu"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var Halt = new(fcpu.Halt)

func runForth(source string) (*fcpu.CPU, error) {
	var err error
	var tmpDir string
	var forthFilename string
	var asmFilename string
	var objFilename string
	var verbose bool
	verbose = false
	// Create temp directory
	tmpDir, err = os.MkdirTemp("", "test")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir) // clean up
	forthFilename = filepath.Join(tmpDir, "source.ft")
	// Write forth source file
	if err = os.WriteFile(forthFilename, []byte(source+" hlt"), 0666); err != nil {
		return nil, err
	}
	// Forth => Asm
	asmFilename = fmt.Sprintf("%s.pal", forthFilename)
	err = Compile(forthFilename, asmFilename)
	if err != nil {
		return nil, err
	}
	// Asm => bytecode
	objFilename = fmt.Sprintf("%s.obj", forthFilename)
	err = asm.Compile(asmFilename, objFilename, verbose)
	if err != nil {
		return nil, err
	}
	// Execute
	cpu, err := fcpu.NewCPU(objFilename)
	if err != nil {
		return nil, err
	}
	for {
		err := cpu.Eval()
		if err != nil {
			if errors.Is(err, Halt) {
				return cpu, nil
			}
			break
		}
	}
	return cpu, nil
}

func testForth(t *testing.T, source string, compareSource string) {
	compareCpu, compareErr := runForth(compareSource)
	if compareErr != nil {
		t.Fatalf("%s", compareErr)
	}
	cpu, err := runForth(source)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if cpu.Ds.Size() != compareCpu.Ds.Size() {
		t.Fatalf("Wrong stack size: %d", cpu.Ds.Size())
	}
	if !reflect.DeepEqual(cpu.Ds.Array(), compareCpu.Ds.Array()) {
		t.Fatalf("Wrong stack content: %d expected: %d", cpu.Ds.Array(), compareCpu.Ds.Array())
	}
}

func Test2Over(t *testing.T) {
	testForth(t,
		"1 2 3 4 2over",
		"1 2 3 4 1 2",
	)
}

func Test2Swap(t *testing.T) {
	testForth(t,
		"1 2 3 4 2swap",
		"3 4 1 2",
	)
}

func Test1Plus(t *testing.T) {
	testForth(t, "0 1+", "1")
	testForth(t, "-1 1+", "0")
	testForth(t, "1 1+", "2")
}

func TestR(t *testing.T) {
	testForth(t, "123 >r r>", "123")
	testForth(t, "15 >r r@ r> drop", "15")
}

func TestComparison(t *testing.T) {
	testForth(t, "9 10 =", "false")
	testForth(t, "-10 -10 =", "true")
	testForth(t, "9 10 <>", "true")
	testForth(t, "-10 -10 <>", "false")
	testForth(t, "10 9 >", "true")
	testForth(t, "9 10 >", "false")
	testForth(t, "10 10 >", "false")
	testForth(t, "10 9 >=", "true")
	testForth(t, "9 10 >=", "false")
	testForth(t, "10 10 >=", "true")
	testForth(t, "10 9 <", "false")
	testForth(t, "9 10 <", "true")
	testForth(t, "10 10 <", "false")
	testForth(t, "10 9 <=", "false")
	testForth(t, "9 10 <=", "true")
	testForth(t, "10 10 <=", "true")
	testForth(t, "0 0=", "true")
	testForth(t, "1 0=", "false")
	testForth(t, "1 0<", "false")
	testForth(t, "-1 0<", "true")
	testForth(t, "1 0>", "true")
	testForth(t, "-1 0>", "false")
}

func TestIf(t *testing.T) {
	testForth(t, "false if 123 then", "")
	testForth(t, "true if 123 then", "123")
	testForth(t, "false if 123 else 79 then", "79")
	testForth(t, "true if 123 else 79 then", "123")
	testForth(t, "2 3 > if 123 else 79 then", "79")
	testForth(t, "2 3 < if 123 else 79 then", "123")
	testForth(t, "20 21 <= if 123 else 79 then", "123")
	testForth(t, "20 20 = if 123 else 79 then", "123")
	testForth(t, "21 20 >= if 123 else 79 then", "123")
	testForth(t, "21 20 <= if 123 else 79 then", "79")
	testForth(t, "-2 0< if true else false then", "true")
	testForth(t, "2 0< if true else false then", "false")
	testForth(t, "-2 0> if true else false then", "false")
	testForth(t, "0 0= if true else false then", "true")
}

func TestLoop(t *testing.T) {
	testForth(t,
		"10 0 do i loop hlt",
		"0 1 2 3 4 5 6 7 8 9",
	)
}

func TestLoopLeave(t *testing.T) {
	testForth(t,
		"10 0 do i . i 4 > if leave then i 10 * loop hlt",
		"0 10 20 30 40",
	)
}

func TestStack(t *testing.T) {
	testForth(t,
		"10 ?dup 0 ?dup",
		"10 10 0",
	)
}

func TestConstant(t *testing.T) {
	testForth(t, `
        32 constant space
        space BL -
        `,
		"0",
	)
}

func TestDefine(t *testing.T) {
	testForth(t, `
        : plus100  100 + ;
        : minus100  100 - ;
        100 plus100
        50 minus100
        `,
		"200 -50",
	)
}

func TestDivMod(t *testing.T) {
	testForth(t, `
        100 10 /
        100 10 /mod
        99 100 /mod
        -99 100 /mod
        -99 -100 /mod
        -99 -100 mod
        -99 100 mod
        99 100 mod
        `,
		"10 0 10 99 0 -99 -1 -99 0 -99 -99 99",
	)
}

func TestComments(t *testing.T) {
	testForth(t, `
        \ comment
        1
        ( aaa. - ) 2
        ( bbb / ccc ddd)3
        ( eee)
        `,
		"1 2 3",
	)
}

func TestMemory(t *testing.T) {
	testForth(t, `
        1024 constant mem
        ( test !, @, +!)
        999 mem !
        5
        mem @
        1 mem +!
        mem @
        ( test 2!, 2@)
        0 100 mem 2!
        mem @
        mem cell+ @
        mem 2@
        `,
		"5 999 1000 100 0 0 100",
	)
}
