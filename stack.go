package main

import (
// "fmt"
// "unsafe"
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
	stack.mmu.WriteW(stack.pointer, value)
	stack.pointer-- // TODO out of stack
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
	value := stack.mmu.ReadW(stack.pointer + 1) // TODO out of stack
	return value, nil
}

func (stack *Stack) Pop() (Word, error) {
	stack.pointer++ // TODO out of stack
	value := stack.mmu.ReadW(stack.pointer)
	stack.mmu.WriteW(stack.pointer, 0) // TODO - TEMP for debugging
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

func (stack *Stack) PrintStack() {
	// if stack.pointer < stack.origin {
	// 	stack := unsafe.Slice((*int)(unsafe.Pointer(&stack.mmu.memory[stack.pointer+1])), stack.origin-stack.pointer)
	// 	fmt.Println("stack: ", stack)
	// } else {
	// 	fmt.Println("stack:  []")
	// }
}
