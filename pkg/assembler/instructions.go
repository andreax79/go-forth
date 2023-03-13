package assembler

import (
	fcpu "github.com/andreax79/go-fcpu/pkg/fcpu"
)

var Instructions = map[string]fcpu.Op{
	"HLT":    fcpu.HLT,
	"NOP":    fcpu.NOP,
	"EMIT":   fcpu.EMIT,
	"PERIOD": fcpu.PERIOD,

	/* Stack manipulation */
	"PUSH":  fcpu.PUSH,
	"ZERO":  fcpu.ZERO,
	"DUP":   fcpu.DUP,
	"DROP":  fcpu.DROP,
	"SWAP":  fcpu.SWAP,
	"OVER":  fcpu.OVER,
	"ROT":   fcpu.ROT,
	"PICK":  fcpu.PICK,
	"ROLL":  fcpu.ROLL,
	"DEPTH": fcpu.DEPTH,

	/* Return Stack manipulation */
	"TO_R":    fcpu.TO_R,
	"R_FROM":  fcpu.R_FROM,
	"R_FETCH": fcpu.R_FETCH,

	/* Arithmetic */
	"ADD":    fcpu.ADD,
	"SUB":    fcpu.SUB,
	"MUL":    fcpu.MUL,
	"DIV":    fcpu.DIV,
	"DIVMOD": fcpu.DIVMOD,
	"MAX":    fcpu.MAX,
	"MIN":    fcpu.MIN,
	"ABS":    fcpu.ABS,
	"MOD":    fcpu.MOD,
	"LSHIFT": fcpu.LSHIFT, // Perform a logical left shift
	"RSHIFT": fcpu.RSHIFT, // Perform a logical right shift

	/* Logical */
	"AND": fcpu.AND,
	"OR":  fcpu.OR,
	"XOR": fcpu.XOR,
	"NOT": fcpu.NOT,

	/* Comparison */
	"EQ": fcpu.EQ,
	"NE": fcpu.NE,
	"GT": fcpu.GT,
	"GE": fcpu.GE,
	"LT": fcpu.LT,
	"LE": fcpu.LE,

	/* Control and subroutines */
	"JNZ":  fcpu.JNZ, // Jump if not zero
	"JZ":   fcpu.JZ,  // Jump if zero
	"JMP":  fcpu.JMP, // Jump
	"CALL": fcpu.CALL,
	"RET":  fcpu.RET,

	/* Memory */
	"STORE":   fcpu.STORE,
	"STORE_B": fcpu.STORE_B,
	"FETCH":   fcpu.FETCH,
	"FETCH_B": fcpu.FETCH_B,

	/* Registers */
	"PUSHRSP": fcpu.PUSHRSP, // Push RSP
	"POPRSP":  fcpu.POPRSP,  // Pop -> RSP
	"PUSHRBP": fcpu.PUSHRSP, // Push RBP
	"POPRBP":  fcpu.POPRBP,  // Pop -> RBP
	"PUSHPC":  fcpu.PUSHPC,  // Push PC
	"POPPC":   fcpu.JMP,     // Pop -> PC ( = JMP)
}
