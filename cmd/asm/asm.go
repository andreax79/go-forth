package main

import (
	"fmt"
	asm "github.com/andreax79/go-fcpu/pkg/assembler"
	fcpu "github.com/andreax79/go-fcpu/pkg/fcpu"
	"os"
)

func main() {
	var asmFilename string
	var objFilename string
	var err error
	asmFilename = os.Args[1]
	objFilename = fmt.Sprintf("%s.obj", asmFilename)
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
