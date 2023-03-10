package assembler

import (
	"errors"
	"fmt"
	fcpu "github.com/andreax79/go-fcpu/pkg/fcpu"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var Halt = new(fcpu.Halt)

func runAsm(source string) (*fcpu.CPU, error) {
	var err error
	var tmpDir string
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
	asmFilename = filepath.Join(tmpDir, "source.pal")
	// Write asm source file
	if err = os.WriteFile(asmFilename, []byte(source+"\nhlt\n"), 0666); err != nil {
		return nil, err
	}
	// Asm => bytecode
	objFilename = fmt.Sprintf("%s.obj", asmFilename)
	err = Compile(asmFilename, objFilename, verbose)
	if err != nil {
		return nil, err
	}
	// Execute
	cpu, err := fcpu.NewCPU(objFilename)
	if err != nil {
		return nil, err
	}
	// cpu.Limit = 100
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

func testAsm(t *testing.T, source string, compareArray []fcpu.Word) {
	cpu, err := runAsm(source)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if cpu.Ds.Size() != fcpu.Addr(len(compareArray)) {
		t.Fatalf("Wrong stack size: %d", cpu.Ds.Size())
	}
	if !reflect.DeepEqual(cpu.Ds.Array(), compareArray) {
		t.Fatalf("Wrong stack content: %d expected: %d", cpu.Ds.Array(), compareArray)
	}
}

func TestPush(t *testing.T) {
	testAsm(t,
		"push 10 push 2 mul",
		[]fcpu.Word{20},
	)
}

func TestLShift(t *testing.T) {
	testAsm(t,
		`push 1 push 0 lshift
		 push 1 push 1 lshift
		 push 1 push 2 lshift`,
		[]fcpu.Word{1, 2, 4},
	)
}

func TestRShift(t *testing.T) {
	testAsm(t,
		`push 1 push 0 rshift
		 push 1 push 1 rshift
		 push 2 push 1 rshift
		 push 4 push 2 rshift`,
		[]fcpu.Word{1, 0, 1, 1},
	)
}
