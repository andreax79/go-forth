package fcpu

import (
	"unsafe"
)

type Op byte

const OpSize = Addr(unsafe.Sizeof(Op(0)))

// Number of POP
const POP0 = 0x00 // No POP
const POP1 = 0x40 // 1 POP
const POP2 = 0x80 // 2 POP

//go:generate stringer -type=Op
const (
	HLT    Op = POP0 + iota
	NOP       = POP0 + iota
	EMIT      = POP1 + iota
	PERIOD    = POP1 + iota

	/* Stack manipulation */
	PUSH   = POP0 + iota /* Push data onto stack */
	PUSH_B = POP0 + iota /* Push data (byte) onto stack */
	DUP    = POP1 + iota /* Duplicates the top stack item */
	DROP   = POP1 + iota /* Discards the top stack item */
	SWAP   = POP2 + iota /* Reverses the top two stack items */
	OVER   = POP2 + iota /* Make copy of second item on top */
	PICK   = POP1 + iota /* Copy n-th item to top */
	ROLL   = POP1 + iota /* Rotate n-th Item to top */
	DEPTH  = POP0 + iota /* Count number of items on stack */

	/* Return Stack manipulation */
	TO_R    = POP1 + iota /* Move top item to the return stack */
	R_FROM  = POP0 + iota /* Retrieve item from the return stack */
	R_FETCH = POP0 + iota /* Copy top of return stack onto stack */

	/* Arithmetic */
	ADD    = POP2 + iota /* Add */
	SUB    = POP2 + iota /* Subtract */
	MUL    = POP2 + iota /* Multiply */
	DIV    = POP2 + iota /* Divide */
	MAX    = POP2 + iota /* Leave greater of two numbers */
	MIN    = POP2 + iota /* Leave lesser of two numbers */
	ABS    = POP1 + iota /* Absolute value */
	MOD    = POP2 + iota /* Modulo */
	LSHIFT = POP2 + iota /* Perform a logical left shift */
	RSHIFT = POP2 + iota /* Perform a logical right shift */

	/* Logical */
	AND = POP2 + iota /* Bitwise and */
	OR  = POP2 + iota /* Bitwise or */
	XOR = POP2 + iota /* Bitwise xor */
	NOT = POP1 + iota /* Reverse true value */

	/* Comparison */
	EQ = POP2 + iota /* Compare Equal */
	NE = POP2 + iota /* Compare for Not Equal */
	GE = POP2 + iota /* Compare for Greater Or Equal */
	GT = POP2 + iota /* Compare for Greater */
	LE = POP2 + iota /* Compare for Equal or Less */
	LT = POP2 + iota /* Compare for Less */

	/* Control and subroutines */
	JNZ  = POP2 + iota /* Jump if not zero */
	JZ   = POP2 + iota /* Jump if zero */
	JMP  = POP1 + iota /* Jump */
	CALL = POP1 + iota /* Subroutine calls */
	RET  = POP0 + iota /* Subroutine return */

	/* Memory */
	STORE   = POP2 + iota
	STORE_B = POP2 + iota
	FETCH   = POP1 + iota
	FETCH_B = POP1 + iota

	/* Registers */
	PUSHRSP = POP0 + iota /* Push RSP */
	POPRSP  = POP1 + iota /* Pop -> RSP */
	PUSHRBP = POP0 + iota /* Push RBP */
	POPRBP  = POP1 + iota /* Pop -> RBP */
	PUSHPC  = POP0 + iota /* Push PC */
)
