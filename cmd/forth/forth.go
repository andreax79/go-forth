package main

import (
	"flag"
	"fmt"
	asm "github.com/andreax79/go-fcpu/pkg/assembler"
	fcpu "github.com/andreax79/go-fcpu/pkg/fcpu"
	forth "github.com/andreax79/go-fcpu/pkg/forth"
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
	cpu.Loop()
	if verbose {
		cpu.PrintMemory()
	}
}

func main() {
	var verbose bool
	var forthFilename string
	var asmFilename string
	var objFilename string
	var err error

	flag.BoolVar(&verbose, "v", false, "Verbose")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("no input file")
		os.Exit(2)
	}
	forthFilename = flag.Args()[0]
	asmFilename = fmt.Sprintf("%s.pal", forthFilename)
	err = forth.Compile(forthFilename, asmFilename)
	if err != nil {
		fmt.Println(err)
		return
	}
	objFilename = fmt.Sprintf("%s.obj", forthFilename)
	err = asm.Compile(asmFilename, objFilename, false)
	if err != nil {
		fmt.Println(err)
		return
	}
	run(objFilename, verbose)
}
