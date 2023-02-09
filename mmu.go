package main

import (
	"fmt"
	"unsafe"
)

type MMU struct {
	memory []Word // [MemorySize]int
}

func NewMMU(size uint) (mmu *MMU) {
	mmu = new(MMU)
	mmu.memory = make([]Word, size)
	return mmu
}

func (mmu *MMU) Size() int {
	return len(mmu.memory)
}

func (mmu *MMU) ReadW(addr Addr) Word {
	return mmu.memory[addr]
}

func (mmu *MMU) WriteW(addr Addr, value Word) {
	mmu.memory[addr] = value
}

func (mmu *MMU) PrintMemory() {
	memory := unsafe.Slice((*int)(unsafe.Pointer(&mmu.memory[0])), len(mmu.memory))
	fmt.Println(memory)
}
