package main

import (
	"fmt"
	"github.com/andreax79/go-fcpu/pkg/fcpu"
	"os"
)

func main() {
	prog, err := fcpu.Compile(os.Args[1])
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
