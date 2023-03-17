package fcpu

import (
	"unsafe"
)

type Device interface {
	Start() Addr
	End() Addr
	ReadW(address Addr) Word
	WriteW(address Addr, value Word)
}

type DeviceDefinition struct {
	start  Addr
	end    Addr
	device Device
}

type Bus struct {
	Mmu     *MMU               // Memory Management Unit
	Devices []DeviceDefinition // Devices
}

func NewBus() (bus *Bus) {
	bus = new(Bus)
	bus.Mmu = NewMMU()
	bus.Devices = []DeviceDefinition{}
	bus.AddDevice(bus.Mmu)
	bus.AddDevice(NewTerminal())
	return bus
}

func (bus *Bus) AddDevice(device Device) {
	bus.Devices = append(bus.Devices, DeviceDefinition{start: device.Start(), end: device.End(), device: device})
}

// Read a word
func (bus *Bus) ReadW(address Addr) Word {
	for _, def := range bus.Devices {
		if address >= def.start && address < def.end {
			return def.device.ReadW(address - def.start)
		}
	}
	return 0
}

// Write a word into Virtual Memory
func (bus *Bus) WriteW(address Addr, value Word) {
	for _, def := range bus.Devices {
		if address >= def.start && address < def.end {
			def.device.WriteW(address-def.start, value)
		}
	}
}

// Read a byte
func (bus *Bus) ReadB(address Addr) byte {
	// Calculate the offset
	off := address & Addr(MemMask)
	// Read the word
	value := bus.ReadW(address - off)
	return (*[4]byte)(unsafe.Pointer(&value))[off]
}

// Write a byte
func (bus *Bus) WriteB(address Addr, value byte) {
	// Calculate the offset
	off := address & Addr(MemMask)
	// Read the word
	wordValue := bus.ReadW(address - off)
	// Updatew the word
	(*[4]byte)(unsafe.Pointer(&wordValue))[off] = value
	bus.WriteW(address-off, wordValue)
}

// Write multiple words
func (bus *Bus) WriteWords(address Addr, value []Word) {
	for i, v := range value {
		bus.WriteW(address+Addr(i)*WordSize, v)
	}
}

// Write multiple bytes
func (bus *Bus) WriteBytes(address Addr, value []byte) {
	for i, v := range value {
		bus.WriteB(address+Addr(i), v)
	}
}
