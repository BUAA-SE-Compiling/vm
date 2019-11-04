package vm

import (
	"bufio"
	"strings"
	"testing"
)

func parseAll(input string) []Instruction {
	scanner := bufio.NewScanner(strings.NewReader(input))
	result := []Instruction{}
	for scanner.Scan() {
		if i := ParseInstruction(scanner.Text()); i != nil {
			result = append(result, *i)
		}
	}
	return result
}

func TestParseInstructions(t *testing.T) {
	input := `
	===== invalid lines will be ignored ======
	ILL
	LIT
	LOD
	STO 1
	ADD
	SUB
	MUL
	DIV
	WRT
	ill
	lit
	lod
	sto 1
	add
	sub
	mul
	div
	wrt
	Ill
	lIt
	Sto
	AdD
	sUB
	mUl
	DiV
	wRT
	



ADD 




LIT                                                      1
                                     LOD            1
					STO 1                            
	==== Below this line are invalid instructions====
	LIT 2147483648
	WRT WRT
	`
	expected := []Instruction{
		Instruction{OP: iILL},
		Instruction{OP: iSTO, X: 1},
		Instruction{OP: iADD},
		Instruction{OP: iSUB},
		Instruction{OP: iMUL},
		Instruction{OP: iDIV},
		Instruction{OP: iWRT},
		Instruction{OP: iILL},
		Instruction{OP: iSTO, X: 1},
		Instruction{OP: iADD},
		Instruction{OP: iSUB},
		Instruction{OP: iMUL},
		Instruction{OP: iDIV},
		Instruction{OP: iWRT},
		Instruction{OP: iILL},
		Instruction{OP: iADD},
		Instruction{OP: iSUB},
		Instruction{OP: iMUL},
		Instruction{OP: iDIV},
		Instruction{OP: iWRT},
		Instruction{OP: iADD},
		Instruction{OP: iLIT, X: 1},
		Instruction{OP: iLOD, X: 1},
		Instruction{OP: iSTO, X: 1}}
	result := parseAll(input)
	if len(result) != len(expected) {
		t.FailNow()
	}
	for i := 0; i < len(result); i++ {
		if result[i] != expected[i] {
			t.Logf("%v != %v", result[i], expected[i])
			t.FailNow()
		}
	}
}

func testVMRun(t *testing.T, input string, expected string) {
	output := strings.Builder{}
	result := parseAll(input)
	v := NewVM(0, 1024, len(result), &output)
	v.Load(result)
	if err := v.Run(); err != nil {
		t.FailNow()
	}
	if output.String() != expected {
		t.FailNow()
	}
}

func testError(t *testing.T, input string, expected error) {
	output := strings.Builder{}
	result := parseAll(input)
	v := NewVM(0, 1024, len(result), &output)
	v.Load(result)
	var err error
	if err = v.Run(); err == nil {
		t.FailNow()
	}
	if err != expected {
		t.FailNow()
	}
}

func TestAdd(t *testing.T) {
	testVMRun(t,
		`
			LIT 0
			LIT 1
			ADD
			WRT
		`,
		"1\n")
	testVMRun(t,
		`
			LIT 1
			LIT 0
			ADD
			WRT
		`,
		"1\n")
}

func TestSub(t *testing.T) {
	testVMRun(t,
		`
			LIT 0
			LIT 1
			SUB
			WRT
		`,
		"-1\n")
	testVMRun(t,
		`
			LIT 1
			LIT 0
			SUB
			WRT
		`,
		"1\n")
}

func TestMul(t *testing.T) {
	testVMRun(t,
		`
			LIT 2
			LIT 5
			MUL
			WRT
		`,
		"10\n")
	testVMRun(t,
		`
			LIT 5
			LIT 2
			MUL
			WRT
		`,
		"10\n")
}

func TestDiv(t *testing.T) {
	testVMRun(t,
		`
			LIT 3
			LIT 7
			DIV
			WRT
		`,
		"0\n")
	testVMRun(t,
		`
			LIT 7
			LIT 3
			DIV
			WRT
		`,
		"2\n")
}

func TestOverflow(t *testing.T) {
	testError(t,
		`
			LIT 2147483647
			LIT 1
			ADD
		`, ErrAlgorithmOverflow)
	testError(t,
		`
			LIT -2
			LIT 2147483647
			SUB
		`, ErrAlgorithmOverflow)
	testError(t,
		`
			LIT -2147483648
			LIT -1
			DIV
		`, ErrAlgorithmOverflow)
	testError(t,
		`
			LIT -2147483648
			LIT -1
			MUL
		`, ErrAlgorithmOverflow)
	testError(t,
		`
			LIT 2147483640
			LIT 5
			MUL
		`, ErrAlgorithmOverflow)
}
