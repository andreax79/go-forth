package fcpu

import (
	"fmt"
	"strings"
)

type Stack struct {
	mmu     *MMU // Memory Management Unit
	origin  Addr
	pointer Addr
}

func NewStack(mmu *MMU, origin Addr) (stack *Stack) {
	stack = new(Stack)
	stack.mmu = mmu
	stack.origin = origin
	stack.pointer = origin
	return stack
}

func (stack *Stack) Push(value Word) error {
	stack.pointer -= WordSize // TODO out of stack
	stack.mmu.WriteW(stack.pointer, value)
	return nil
}

func (stack *Stack) PushBool(value bool) error {
	if value {
		return stack.Push(1)
	} else {
		return stack.Push(0)
	}
}

func (stack *Stack) Get() (Word, error) {
	value := stack.mmu.ReadW(stack.pointer) // TODO out of stack
	return value, nil
}

func (stack *Stack) Pop() (Word, error) {
	value := stack.mmu.ReadW(stack.pointer)
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
	for i := stack.pointer; i < stack.origin; i += WordSize {
		fmt.Fprintf(&buf, "%d ", stack.mmu.ReadW(i))
	}
	return buf.String()
}
