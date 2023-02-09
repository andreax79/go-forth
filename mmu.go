package main

import (
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
	pages map[Page][]Word // memory pages
}

func NewMMU() (mmu *MMU) {
	mmu = new(MMU)
	mmu.pages = map[Page][]Word{}
	mmu.pages[0] = make([]Word, VirtualPageSize)
	return mmu
}

// Read a word from Virtual Memory
func (mmu *MMU) ReadW(address Addr) Word {
	var page []Word
	page = mmu.GetPage(address)
	return page[mmu.GetOffset(address)]
}

// Write a word into Virtual Memory
func (mmu *MMU) WriteW(address Addr, value Word) {
	var page []Word
	page = mmu.GetPage(address)
	page[mmu.GetOffset(address)] = value
}

// Write multiple words into Virtual Memory
func (mmu *MMU) WriteWords(address Addr, value []Word) {
	for i, v := range value {
		mmu.WriteW(address+Addr(i), v)
	}
}

func (mmu *MMU) PrintMemory() {
	for page, data := range mmu.pages {
		memory := unsafe.Slice((*int)(unsafe.Pointer(&data[0])), VirtualPageSize)
		fmt.Printf("Page: %d\n", page)
		fmt.Println(memory)
	}
}

// Get a memory page by Virtual Address
func (mmu *MMU) GetPage(address Addr) []Word {
	var pageNumber Page
	pageNumber = mmu.GetPageNumber(address)
	page, exists := mmu.pages[pageNumber]
	if !exists {
		page = make([]Word, VirtualPageSize)
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
