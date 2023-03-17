package fcpu

import (
	"fmt"
	"time"
)

type Terminal struct {
	start Addr
	ready Word
	out   Word
}

func NewTerminal() (term *Terminal) {
	term = new(Terminal)
	term.start = MemoryLimit
	go term.clock()
	return term
}

func (term *Terminal) Start() Addr {
	return term.start
}

func (term *Terminal) End() Addr {
	return term.start + 10
}

func (term *Terminal) ReadW(address Addr) Word {
	switch address {
	case 0:
		return term.ready
	}
	return 0
}

func (term *Terminal) WriteW(address Addr, value Word) {
	switch address {
	case 0:
		// fmt.Println("set ready to", value)
		term.ready = value
	case 4:
		// fmt.Println("set out to", value)
		term.out = value
	}
	// fmt.Println("x")
}

func (term *Terminal) clock() {
	for {
		time.Sleep(1000 / 9600 * 10 * time.Millisecond)
		if term.ready != 0 {
			// fmt.Println("aaa", term.out)
			fmt.Printf("%c", int(term.out))
			term.ready = 0
		}
		// cpu.bus.WriteW(MemoryLimit, Word(0))
	}
}
