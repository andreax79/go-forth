package main

import (
	_ "encoding/binary"
	"fmt"
	"unsafe"
)

type Page uint

// Offset size
const VirtualPageShift = 10

// Page size
const VirtualPageSize = 1 << VirtualPageShift

// Offset mask
const VirtualOffsetMask = VirtualPageSize - 1

type MMU struct {
	pages map[Page][]byte // memory pages
}

func NewMMU() (mmu *MMU) {
	mmu = new(MMU)
	mmu.pages = map[Page][]byte{}
	mmu.pages[0] = make([]byte, VirtualPageSize)
	return mmu
}

// Read a byte from Virtual Memory
func (mmu *MMU) ReadB(address Addr) byte {
	var page []byte
	var offset Addr
	offset = mmu.GetOffset(address)
	page = mmu.GetPage(address)
	return page[offset]
}

// Read a word from Virtual Memory
func (mmu *MMU) ReadW(address Addr) Word {
	var page []byte
	var offset Addr
	offset = mmu.GetOffset(address)
	page = mmu.GetPage(address)
	return *(*Word)(unsafe.Pointer(&page[offset]))
}

// Write a byte into Virtual Memory
func (mmu *MMU) WriteB(address Addr, value byte) {
	var page []byte
	var offset Addr
	offset = mmu.GetOffset(address)
	page = mmu.GetPage(address)
	page[offset] = value
}

// Write a word into Virtual Memory
func (mmu *MMU) WriteW(address Addr, value Word) {
	var page []byte
	var offset Addr
	offset = mmu.GetOffset(address)
	page = mmu.GetPage(address)
	*(*Word)(unsafe.Pointer(&page[offset])) = value
}

// Write multiple words into Virtual Memory
func (mmu *MMU) WriteWords(address Addr, value []Word) {
	for i, v := range value {
		mmu.WriteW(address+Addr(i)*WordSize, v)
	}
}

// Write multiple bytes into Virtual Memory
func (mmu *MMU) WriteBytes(address Addr, value []byte) {
	for i, v := range value {
		mmu.WriteB(address+Addr(i), v)
	}
}

func (mmu *MMU) PrintMemory() {
	for page, data := range mmu.pages {
		memory := unsafe.Slice((*int32)(unsafe.Pointer(&data[0])), VirtualPageSize/WordSize)
		fmt.Printf("%04d ", page)
		fmt.Println(memory)
	}
}

// Get a memory page by Virtual Address
func (mmu *MMU) GetPage(address Addr) []byte {
	var pageNumber Page
	pageNumber = mmu.GetPageNumber(address)
	// fmt.Println("addr", address, "page", pageNumber, mmu.GetOffset(address))
	page, exists := mmu.pages[pageNumber]
	if !exists {
		page = make([]byte, VirtualPageSize)
		mmu.pages[pageNumber] = page
	}
	return page
}

// Get the page number from a Virtual Address
func (mmu *MMU) GetPageNumber(address Addr) Page {
	return Page(address >> VirtualPageShift)
}

// Get the page offset from a Virtual Address
func (mmu *MMU) GetOffset(address Addr) Addr {
	return address & VirtualOffsetMask
}
