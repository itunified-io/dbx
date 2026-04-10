// Package confirm provides confirmation gates for destructive operations.
package confirm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Level indicates how much confirmation a tool requires.
type Level int

const (
	LevelNone          Level = iota // Read-only, no confirm
	LevelStandard                   // --confirm flag required
	LevelEchoBack                   // Type exact string to confirm
	LevelDoubleConfirm              // Two prompts (catastrophic)
)

var (
	ErrConfirmRequired = errors.New("confirmation required: pass confirm=true or --confirm")
	ErrConfirmMismatch = errors.New("confirmation input did not match expected value")
)

// Gate handles confirmation prompts for destructive operations.
type Gate struct {
	reader *bufio.Reader
	writer io.Writer
}

// New creates a confirm gate. Pass nil for non-interactive mode.
func New(in io.Reader, out io.Writer) *Gate {
	g := &Gate{writer: out}
	if in != nil {
		g.reader = bufio.NewReader(in)
	}
	return g
}

// Check validates the confirm level. For LevelStandard, confirmFlag=true bypasses.
func (g *Gate) Check(level Level, description, targetName string, confirmFlag bool) error {
	if level == LevelNone {
		return nil
	}
	if confirmFlag {
		return nil
	}
	return ErrConfirmRequired
}

// CheckEchoBack prompts the user to type an exact string to confirm.
func (g *Gate) CheckEchoBack(echoString, description string) error {
	if g.writer != nil {
		fmt.Fprintf(g.writer, "To %s, type: %s\n> ", description, echoString)
	}
	if g.reader == nil {
		return ErrConfirmRequired
	}
	line, err := g.reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading confirmation: %w", err)
	}
	if strings.TrimSpace(line) != echoString {
		return ErrConfirmMismatch
	}
	return nil
}

// CheckDoubleConfirm prompts twice: first echo-back, then a second confirmation string.
func (g *Gate) CheckDoubleConfirm(firstEcho, secondEcho, description string) error {
	if g.writer != nil {
		fmt.Fprintf(g.writer, "To %s, type: %s\n> ", description, firstEcho)
	}
	if g.reader == nil {
		return ErrConfirmRequired
	}
	line, err := g.reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading first confirmation: %w", err)
	}
	if strings.TrimSpace(line) != firstEcho {
		return ErrConfirmMismatch
	}

	if g.writer != nil {
		fmt.Fprintf(g.writer, "WARNING: This is irreversible. Type %s to proceed.\n> ", secondEcho)
	}
	line, err = g.reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading second confirmation: %w", err)
	}
	if strings.TrimSpace(line) != secondEcho {
		return ErrConfirmMismatch
	}
	return nil
}
