package main

import (
	"flag"
	"fmt"
	asm "github.com/andreax79/go-fcpu/pkg/assembler"
	fcpu "github.com/andreax79/go-fcpu/pkg/fcpu"
	"os"
)

// Run obj file
func run(objFilename string, verbose bool) {
	cpu, err := fcpu.NewCPU(objFilename)
	cpu.Verbose = verbose
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		err := cpu.Eval()
		if err != nil {
			break
		}
	}
	if verbose {
		cpu.PrintMemory()
	}
}

func main() {
	var verbose bool
	var asmFilename string
	var objFilename string
	var err error

	flag.BoolVar(&verbose, "v", false, "Verbose")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("no input file")
		os.Exit(2)
	}
	asmFilename = flag.Args()[0]
	objFilename = fmt.Sprintf("%s.obj", asmFilename)
	err = asm.Compile(asmFilename, objFilename, verbose)
	if err != nil {
		fmt.Println(err)
		return
	}
	run(objFilename, verbose)
}
