package main

import (
	"fmt"
	asm "github.com/andreax79/go-fcpu/pkg/assembler"
	fcpu "github.com/andreax79/go-fcpu/pkg/fcpu"
	forth "github.com/andreax79/go-fcpu/pkg/forth"
	"os"
)

func main() {
	var forthFilename string
	var asmFilename string
	var objFilename string
	var err error
	forthFilename = os.Args[1]
	asmFilename = fmt.Sprintf("%s.pal", forthFilename)
	err = forth.Compile(forthFilename, asmFilename)
	if err != nil {
		fmt.Println(err)
		return
	}
	objFilename = fmt.Sprintf("%s.obj", forthFilename)
	err = asm.Compile(asmFilename, objFilename)
	if err != nil {
		fmt.Println(err)
		return
	}
	cpu, err := fcpu.NewCPU(objFilename)
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
	cpu.PrintMemory()
}
