package fcpu

import (
	"unsafe"
)

type Op byte

const OpSize = Addr(unsafe.Sizeof(Op(0)))

//go:generate stringer -type=Op
const (
	HLT Op = iota
	NOP
	EMIT
	PERIOD

	/* Stack manipulation */
	PUSH   /* Push data onto stack */
	PUSH_B /* Push data (byte) onto stack */
	ZERO   /* Push 0 onto stack */
	DUP    /* Duplicates the top stack item */
	DROP   /* Discards the top stack item */
	SWAP   /* Reverses the top two stack items */
	OVER   /* Make copy of second item on top */
	ROT    /* Rotate third item to top */
	PICK   /* Copy n-th item to too */
	ROLL   /* Rotate n-th Item to top */
	DEPTH  /* Count number of items on stack */

	/* Return Stack manipulation */
	TO_R    /* Move top item to the return stack */
	R_FROM  /* Retrieve item from the return stack */
	R_FETCH /* Copy top of return stack onto stack */

	/* Arithmetic */
	ADD    /* Add */
	SUB    /* Subtract */
	MUL    /* Multiply */
	DIV    /* Divide */
	MAX    /* Leave greater of two numbers */
	MIN    /* Leave lesser of two numbers */
	ABS    /* Absolute value */
	MOD    /* Modulo */
	LSHIFT /* Perform a logical left shift */
	RSHIFT /* Perform a logical right shift */

	/* Logical */
	AND /* Bitwise and */
	OR  /* Bitwise or */
	XOR /* Bitwise xor */
	NOT /* Reverse true value */

	/* Comparison */
	EQ /* Compare Equal */
	NE /* Compare for Not Equal */
	GE /* Compare for Greater Or Equal */
	GT /* Compare for Greater */
	LE /* Compare for Equal or Less */
	LT /* Compare for Less */

	/* Control and subroutines */
	JNZ  /* Jump if not zero */
	JZ   /* Jump if zero */
	JMP  /* Jump */
	CALL /* Subroutine calls */
	RET  /* Subroutine return */

	/* Memory */
	STORE
	STORE_B
	FETCH
	FETCH_B

	/* Registers */
	PUSHRSP /* Push RSP */
	POPRSP  /* Pop -> RSP */
	PUSHRBP /* Push RBP */
	POPRBP  /* Pop -> RBP */
	PUSHPC  /* Push PC */
)
