package fcpu

import (
	"testing"
	"unsafe"
)

func TestMMU(t *testing.T) {
	mmu := NewMMU()
	for i := 0; i < 1024; i++ {
		mmu.WriteB(Addr(i), byte(i%256))
	}
	for i := 0; i < 1024; i++ {
		if mmu.ReadB(Addr(i)) != byte(i%256) {
			t.Fatalf("read error")
		}
	}
	for i := 2047; i >= 1024; i-- {
		off := i & MemMask
		value := mmu.ReadW(Addr(i - off))
		x := (*[4]byte)(unsafe.Pointer(&value))
		x[off] = byte(i % 256)
		mmu.WriteW(Addr(i-off), value)
	}
	for i := 1024; i < 2048; i++ {
		if mmu.ReadB(Addr(i)) != byte(i%256) {
			t.Fatalf("read error 1")
		}
	}
	for i := 1024; i < 2048; i++ {
		off := i & MemMask
		value := mmu.ReadW(Addr(i - off))
		v := (*[4]byte)(unsafe.Pointer(&value))[off]
		if v != byte(i%256) {
			t.Fatalf("read error 2")
		}
	}
}

func TestBus(t *testing.T) {
	bus := NewBus()
	for i := 0; i < 1024; i++ {
		bus.WriteB(Addr(i), byte(i%256))
	}
	for i := 0; i < 1024; i++ {
		if bus.ReadB(Addr(i)) != byte(i%256) {
			t.Fatalf("read error")
		}
	}
	for i := 0; i < 1024; i++ {
		bus.WriteW(Addr(i)*WordSize, Word(i))
	}
	for i := 0; i < 1024; i++ {
		if bus.ReadW(Addr(i)*WordSize) != Word(i) {
			t.Fatalf("read error")
		}
	}
}
