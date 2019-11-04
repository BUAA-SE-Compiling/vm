package vm

import (
	"fmt"
	"strconv"
	"strings"
)

type operation int32

const (
	iILL operation = iota
	iLIT
	iLOD
	iSTO
	iADD
	iSUB
	iMUL
	iDIV
	iWRT
)

// Instruction represents an instruction.
// See: ()
type Instruction struct {
	OP operation
	X  int32
}

func (op operation) String() string {
	switch op {
	case iILL:
		return "ILL"
	case iLIT:
		return "LIT"
	case iLOD:
		return "LOD"
	case iSTO:
		return "STO"
	case iADD:
		return "ADD"
	case iSUB:
		return "SUB"
	case iMUL:
		return "MUL"
	case iDIV:
		return "DIV"
	case iWRT:
		return "WRT"
	}
	return "BAD"
}

func (i Instruction) String() string {
	switch i.OP {
	case iLIT:
		fallthrough
	case iLOD:
		fallthrough
	case iSTO:
		return fmt.Sprintf("%v %v", i.OP, i.X)
	case iILL:
		fallthrough
	case iADD:
		fallthrough
	case iSUB:
		fallthrough
	case iMUL:
		fallthrough
	case iDIV:
		fallthrough
	case iWRT:
		return fmt.Sprintf("%v", i.OP)
	}
	return "BAD"
}

// ParseInstruction parses an instruction from a string.
func ParseInstruction(line string) *Instruction {
	tp := strings.TrimSpace(line)
	tokens := strings.Split(tp, " ")
	nonemptytokens := []string{}
	for _, token := range tokens {
		if token != "" {
			nonemptytokens = append(nonemptytokens, token)
		}
	}
	tokens = nonemptytokens
	if len(tokens) == 1 {
		switch strings.ToUpper(tokens[0]) {
		case "ILL":
			return &Instruction{OP: iILL, X: 0}
		case "ADD":
			return &Instruction{OP: iADD, X: 0}
		case "SUB":
			return &Instruction{OP: iSUB, X: 0}
		case "MUL":
			return &Instruction{OP: iMUL, X: 0}
		case "DIV":
			return &Instruction{OP: iDIV, X: 0}
		case "WRT":
			return &Instruction{OP: iWRT, X: 0}
		}
		return nil
	}
	if len(tokens) == 2 {
		x, err := strconv.ParseInt(tokens[1], 10, 32)
		if err != nil {
			return nil
		}
		switch strings.ToUpper(tokens[0]) {
		case "LIT":
			return &Instruction{OP: iLIT, X: int32(x)}
		case "LOD":
			return &Instruction{OP: iLOD, X: int32(x)}
		case "STO":
			return &Instruction{OP: iSTO, X: int32(x)}
		}
		return nil
	}
	return nil
}
