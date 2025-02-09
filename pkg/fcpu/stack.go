package fcpu

import (
	"fmt"
	"strings"
)

type Stack struct {
	bus     *Bus
	origin  Addr
	pointer Addr
}

func NewStack(bus *Bus, origin Addr) (stack *Stack) {
	stack = new(Stack)
	stack.bus = bus
	stack.origin = origin
	stack.pointer = origin
	return stack
}

func (stack *Stack) Push(value Word) error {
	stack.pointer -= WordSize // TODO out of stack
	stack.bus.WriteW(stack.pointer, value)
	return nil
}

func (stack *Stack) PushBool(value bool) error {
	if value {
		return stack.Push(-1)
	} else {
		return stack.Push(0)
	}
}

func (stack *Stack) Get() (Word, error) {
	value := stack.bus.ReadW(stack.pointer) // TODO out of stack
	return value, nil
}

func (stack *Stack) Pop() (Word, error) {
	value := stack.bus.ReadW(stack.pointer)
	stack.pointer += WordSize // TODO out of stack
	return value, nil
}

func (stack *Stack) Dup() error {
	var value Word
	var err error
	if value, err = stack.Get(); err != nil {
		return err
	}
	return stack.Push(value)
}

// Copy the xu to the top of the stack
func (stack *Stack) Pick(n Word) error {
	addr := stack.pointer + Addr(n)*WordSize
	value := stack.bus.ReadW(addr) // TODO out of stack
	return stack.Push(value)
}

// Rotate u+1 items on the top of the stack
func (stack *Stack) Roll(n Word) error {
	if n <= 0 {
		return nil
	}
	var last Word
	for i := Word(0); i < n; i++ {
		addr := stack.pointer + Addr(i)*WordSize
		value := stack.bus.ReadW(addr) // TODO out of stack
		fmt.Println(i, addr, value, last)
		if i > 0 {
			stack.bus.WriteW(addr, last)
		}
		last = value
	}
	stack.bus.WriteW(stack.pointer, last)
	return nil
}

func (stack *Stack) Pop2() (Word, Word, error) {
	var v1 Word
	var v2 Word
	var err error
	if v2, err = stack.Pop(); err != nil {
		return 0, 0, err
	}
	if v1, err = stack.Pop(); err != nil {
		return 0, 0, err
	}
	return v1, v2, nil
}

func (stack *Stack) Size() Addr {
	return (stack.origin - stack.pointer) / WordSize
}

func (stack *Stack) String() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "(%d) ", stack.Size())
	for i := stack.origin - WordSize; i >= stack.pointer; i -= WordSize {
		fmt.Fprintf(&buf, "%x ", uint32(stack.bus.ReadW(i)))
	}
	return buf.String()
}

func (stack *Stack) Array() []Word {
	array := make([]Word, stack.Size())
	for i := Addr(0); i < stack.Size(); i++ {
		array[stack.Size()-1-i] = stack.bus.ReadW(stack.pointer + i*WordSize)
	}
	return array
}
