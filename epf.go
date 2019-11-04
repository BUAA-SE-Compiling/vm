package vm

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

// errros
var (
	ErrWrongSignature = errors.New("wrong signature")
	ErrWrongVersion   = errors.New("wrong version")
)

// EPFHeaderv1 represents the header of an EPFv1 file.
// EPFv1 : Executable Pcode Formart Version 1
type EPFHeaderv1 struct {
	Magic             [4]byte
	Version           int32
	InstructionCounts int32
	EntryPoint        int32
}

// EPFv1 represents the EPF in the memory.
type EPFv1 struct {
	version      int32
	instructions []Instruction
	entrypoint   int32
}

// NewEPFv1FromInstructions creates a new EPF from an array of instructions.
func NewEPFv1FromInstructions(instructions []Instruction, entry int32) *EPFv1 {
	return &EPFv1{
		version:      1,
		instructions: instructions,
		entrypoint:   entry}
}

// NewEPFv1FromFile creates a new EPF from a file.
func NewEPFv1FromFile(file io.Reader) (*EPFv1, error) {
	epf := &EPFv1{
		version:      0,
		instructions: nil,
		entrypoint:   0}
	header := &EPFHeaderv1{}
	if err := binary.Read(file, binary.LittleEndian, header); err != nil {
		return nil, err
	}
	if string(header.Magic[0:4]) != "ZQLS" {
		return nil, ErrWrongSignature
	}
	if header.Version != 1 {
		return nil, ErrWrongVersion
	}
	epf.instructions = make([]Instruction, header.InstructionCounts)
	for i := 0; i < int(header.InstructionCounts); i++ {
		temp := Instruction{}
		if err := binary.Read(file, binary.LittleEndian, &temp); err != nil {
			return nil, err
		}
		epf.instructions[i] = temp
	}
	return epf, nil
}

// WriteFile writes an epf to a file.
func (epf *EPFv1) WriteFile(file *os.File) error {
	header := &EPFHeaderv1{
		Magic:             [4]byte{'Z', 'Q', 'L', 'S'},
		Version:           1,
		InstructionCounts: int32(len(epf.instructions)),
		EntryPoint:        epf.entrypoint}
	if err := binary.Write(file, binary.LittleEndian, header); err != nil {
		return err
	}
	for _, i := range epf.instructions {
		if err := binary.Write(file, binary.LittleEndian, &i); err != nil {
			return err
		}
	}
	return nil
}

// GetInstructions gets all instructions
func (epf *EPFv1) GetInstructions() []Instruction {
	return epf.instructions
}

// GetEntry gets the entry point.
func (epf *EPFv1) GetEntry() int32 {
	return epf.entrypoint
}
