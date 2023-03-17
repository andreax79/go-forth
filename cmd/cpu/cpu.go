package main

import (
	"flag"
	"fmt"
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
	cpu.Loop()
	if verbose {
		cpu.PrintMemory()
	}
}

func main() {
	var verbose bool
	var objFilename string

	flag.BoolVar(&verbose, "v", false, "Verbose")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("no input file")
		os.Exit(2)
	}
	objFilename = flag.Args()[0]
	run(objFilename, verbose)
}
