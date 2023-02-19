package assembler

import (
	"errors"
	"fmt"
)

type UndefinedSymbol struct {
	Label string
	Line  int
}

func (e *UndefinedSymbol) Error() string {
	return fmt.Sprintf("Undefined symbol %s in line %d", e.Label, e.Line)
}

type UndefinedDirective struct {
	Label string
	Line  int
}

func (e *UndefinedDirective) Error() string {
	return fmt.Sprintf("Undefined directive %s in line %d", e.Label, e.Line)
}

type LabelMultipleDefinition struct {
	Label string
	Line  int
}

func (e *LabelMultipleDefinition) Error() string {
	return fmt.Sprintf("Multiple definition of a label %s in line %d", e.Label, e.Line)
}

type UnexpectedToken struct {
	Token string
	Line  int
}

func (e *UnexpectedToken) Error() string {
	return fmt.Sprintf("Unexpected token %s in line %d", e.Token, e.Line)
}

var UnmatechedDelimiter = errors.New("Unmateched delimiter")
