package main

import (
	"fmt"
	asm "github.com/andreax79/go-fcpu/pkg/assembler"
	fcpu "github.com/andreax79/go-fcpu/pkg/fcpu"
	forth "github.com/andreax79/go-fcpu/pkg/forth"
	"os"
)

func main() {
	var asmFilename string
	var err error
	var prog []byte
	asmFilename, err = forth.Compile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	prog, err = asm.Compile(asmFilename)
	if err != nil {
		fmt.Println(err)
		return
	}
	cpu := fcpu.NewCPU(prog)
	for {
		err := cpu.Eval()
		if err != nil {
			break
		}
	}
	cpu.PrintMemory()
}
