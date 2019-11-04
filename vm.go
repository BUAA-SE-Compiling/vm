package vm

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// errors
var (
	ErrCodeSize           = errors.New("too many or few codes")
	ErrIllegalInstruction = errors.New("illegal instruction")
	ErrMemoryError        = errors.New("memory error")
	ErrAlgorithmOverflow  = errors.New("algorithm error")
	ErrDividedByZeroError = errors.New("divided by zero error")
	ErrStackOverflow      = errors.New("stack overflow")
)

const (
	maxuint32 = ^uint32(0)
	minuint32 = 0
	maxint32  = int32(maxuint32 >> 1)
	minint32  = -maxint32 - 1
)

// VM represents the virtual machine.
// See: (TODO: design docs url.)
type VM struct {
	IP     int32
	SP     int32
	Stack  []int32
	Code   []Instruction
	writer io.Writer
}

// NewVMDefault returns a VM with the default configuration.
func NewVMDefault(codesize int, entry int32) *VM {
	return NewVM(entry, 2048, codesize, os.Stdout)
}

// NewVM creates a VM.
func NewVM(entry int32, stacksize int, codesize int, writer io.Writer) *VM {
	return &VM{
		IP:     entry,
		SP:     0,
		Stack:  make([]int32, stacksize),
		Code:   make([]Instruction, codesize),
		writer: writer}
}

// Micro codes.

func (vm *VM) readInstruction(pos int32) (*Instruction, error) {
	if int(pos) >= len(vm.Code) || pos < 0 {
		return nil, ErrIllegalInstruction
	}
	return &vm.Code[pos], nil
}

func (vm *VM) fetchInstruction() (*Instruction, error) {
	instruction, err := vm.readInstruction(vm.IP)
	if err != nil {
		return nil, err
	}
	if int(instruction.OP) > int(iWRT) || instruction.OP == iILL {
		return nil, ErrIllegalInstruction
	}
	vm.IP++
	return instruction, nil
}

func (vm *VM) readStack(pos int32) (int32, error) {
	if int(pos) >= len(vm.Stack) || pos < 0 {
		return 0, ErrMemoryError
	}
	return vm.Stack[pos], nil
}

func (vm *VM) writeStack(pos int32, x int32) error {
	if int(pos) >= len(vm.Stack) || pos < 0 {
		return ErrMemoryError
	}
	vm.Stack[pos] = x
	return nil
}

func (vm *VM) increaseSP() error {
	if int(vm.SP) == len(vm.Stack)-1 {
		return ErrStackOverflow
	}
	vm.SP++
	return nil
}

func (vm *VM) decreaseSP() error {
	if int(vm.SP) == 0 {
		return ErrStackOverflow
	}
	vm.SP--
	return nil
}

func (vm *VM) putInt32WithNewLine(x int32) error {
	s := fmt.Sprintf("%v", x)
	if _, err := vm.writer.Write([]byte(s + "\n")); err != nil {
		return err
	}
	return nil
}

type opr func(int32, int32) (int32, error)

func (vm *VM) doAlgorithm(f opr) (int32, error) {
	var x int32
	var y int32
	var err error
	if x, err = vm.readStack(vm.SP - 1); err != nil {
		return 0, err
	}
	if y, err = vm.readStack(vm.SP - 2); err != nil {
		return 0, err
	}
	if x, err = f(x, y); err != nil {
		return 0, err
	}
	return x, nil
}

func (vm *VM) doAlgorithmAndWrite(f opr) error {
	var result int32
	var err error
	if result, err = vm.doAlgorithm(f); err != nil {
		return err
	}
	if err = vm.writeStack(vm.SP-2, result); err != nil {
		return err
	}
	if err = vm.decreaseSP(); err != nil {
		return err
	}
	return nil
}

func (vm *VM) runNext() error {
	var next *Instruction
	var err error
	if next, err = vm.fetchInstruction(); err != nil {
		return err
	}
	op := next.OP
	x := next.X
	switch op {
	case iLIT:
		if err = vm.writeStack(vm.SP, x); err != nil {
			return err
		}
		if err = vm.increaseSP(); err != nil {
			return err
		}
	case iLOD:
		var val int32
		if val, err = vm.readStack(x); err != nil {
			return err
		}
		if err = vm.writeStack(vm.SP, val); err != nil {
			return err
		}
		if err = vm.increaseSP(); err != nil {
			return err
		}
	case iSTO:
		var val int32
		if val, err = vm.readStack(vm.SP - 1); err != nil {
			return err
		}
		if err = vm.writeStack(x, val); err != nil {
			return err
		}
		if err = vm.decreaseSP(); err != nil {
			return err
		}
	case iADD:
		add := func(x int32, y int32) (int32, error) {
			if int64(x)+int64(y) > int64(maxint32) || int64(x)+int64(y) < int64(minint32) {
				return 0, ErrAlgorithmOverflow
			}
			return x + y, nil
		}
		if err = vm.doAlgorithmAndWrite(add); err != nil {
			return err
		}
	case iSUB:
		sub := func(x int32, y int32) (int32, error) {
			if int64(y)-int64(x) > int64(maxint32) || int64(y)-int64(x) < int64(minint32) {
				return 0, ErrAlgorithmOverflow
			}
			return y - x, nil
		}
		if err = vm.doAlgorithmAndWrite(sub); err != nil {
			return err
		}
	case iDIV:
		div := func(x int32, y int32) (int32, error) {
			if x == 0 {
				return 0, ErrDividedByZeroError
			}
			if x == -1 && y == minint32 {
				return 0, ErrAlgorithmOverflow
			}
			return y / x, nil
		}
		if err = vm.doAlgorithmAndWrite(div); err != nil {
			return err
		}
	case iMUL:
		mul := func(x int32, y int32) (int32, error) {
			r := x * y
			if y != 0 && r/y == x {
				return y * x, nil
			}
			return 0, ErrAlgorithmOverflow
		}
		if err = vm.doAlgorithmAndWrite(mul); err != nil {
			return err
		}
	case iWRT:
		var x int32
		if x, err = vm.readStack(vm.SP - 1); err != nil {
			return err
		}
		if err = vm.putInt32WithNewLine(x); err != nil {
			return err
		}
		if err = vm.decreaseSP(); err != nil {
			return err
		}
	}
	return nil
}

// Load instructions
func (vm *VM) Load(instructions []Instruction) error {
	if len(instructions) != len(vm.Code) {
		return ErrCodeSize
	}
	for i := 0; i < len(vm.Code); i++ {
		vm.Code[i] = instructions[i]
	}
	return nil
}

// RunSingle runs a single instruction.
func (vm *VM) RunSingle() error {

	return vm.runNext()
}

// Run runs the VM to the end.
func (vm *VM) Run() error {
	for i := 0; i < len(vm.Code); i++ {
		if err := vm.runNext(); err != nil {
			return err
		}
	}
	return nil
}

// StackGraph returns a graph of the stack.
func (vm *VM) StackGraph(maxsize int32) string {
	var min int32
	sb := strings.Builder{}
	if vm.SP-maxsize >= 0 {
		min = vm.SP - maxsize
	} else {
		min = 0
	}
	if vm.SP >= int32(len(vm.Stack)) {
		return "Stack graph is not available."
	}
	for i := vm.SP; i >= min; i-- {
		if i == vm.SP {
			sb.WriteString(fmt.Sprintf("|\t%v\t| <-- sp", vm.Stack[i]))
		} else {
			sb.WriteString(fmt.Sprintf("|\t%v\t| %v", vm.Stack[i], i))
		}
		if i != min {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// InstructionGraph returns a graph of the stack.
func (vm *VM) InstructionGraph(halfsize int32) string {
	var max int32
	var min int32
	sb := strings.Builder{}
	if vm.IP+halfsize >= int32(len(vm.Code))-1 {
		max = int32(len(vm.Code)) - 1
	} else {
		max = vm.IP + halfsize
	}
	if vm.IP-halfsize+1 >= 0 {
		min = vm.IP - halfsize + 1
	} else {
		min = 0
	}
	if min > int32(len(vm.Code)) || max > int32(len(vm.Code)) {
		return "Instructions graph is not available."
	}
	for i := min; i <= max; i++ {
		if i == vm.IP {
			sb.WriteString(fmt.Sprintf("|\t%v\t| <-- ip", vm.Code[i]))
		} else {
			sb.WriteString(fmt.Sprintf("|\t%v\t| %v", vm.Code[i], i))
		}
		if i != max {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// NextInstruction returns the next instruction.
func (vm *VM) NextInstruction() *Instruction {
	i, err := vm.readInstruction(vm.IP)
	if err != nil {
		return &Instruction{OP: iILL}
	}
	return i
}

// GetStackTop gets Stack[SP-1]
func (vm *VM) GetStackTop() *int32 {
	tp, err := vm.readStack(vm.SP - 1)
	if err != nil {
		return nil
	}
	return &tp
}

func (vm VM) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Registers: ip=%v sp=%v\n", vm.IP, vm.SP))
	if int(vm.IP-1) < len(vm.Code) && int(vm.IP-1) >= 0 {
		sb.WriteString(fmt.Sprintf("Last Instruction: %v\n", vm.Code[vm.IP-1]))
	}
	sb.WriteString("Stack:\n")
	sb.WriteString(vm.StackGraph(20))
	sb.WriteString("\nInstructions:\n")
	sb.WriteString(vm.InstructionGraph(5))

	return sb.String()
}
