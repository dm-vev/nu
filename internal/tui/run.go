package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"nu/internal/tui/input"
	"nu/internal/tui/terminal"
)

// Run starts the interactive loop.
func (a *App) Run(ctx context.Context) (runErr error) {
	a.submitCtx = ctx
	a.editor.SetSubmitHandler(func(value string) {
		if err := a.submit(value); err != nil {
			a.rememberWriteErr(err)
		}
	})
	restore, raw, err := a.term.EnableRaw()
	if err != nil {
		return fmt.Errorf("enable tui raw mode: %w", err)
	}
	if restore != nil {
		defer func() {
			if err := restore(); err != nil && runErr == nil {
				runErr = fmt.Errorf("restore tui raw mode: %w", err)
			}
		}()
	}
	if err := a.ui.Start(); err != nil {
		return fmt.Errorf("start tui: %w", err)
	}
	defer func() {
		// Wait before restoring terminal state so late provider events still render into the managed frame.
		a.promptWG.Wait()
		if err := a.ui.Stop(); err != nil && runErr == nil {
			runErr = fmt.Errorf("stop tui: %w", err)
		}
	}()
	a.render()
	stopStatus := a.startStatusTicker(ctx)
	defer stopStatus()
	if raw {
		stopResize := terminal.WatchResize(a.render)
		defer stopResize()
		return a.runRaw(ctx)
	}
	return a.runLine(ctx)
}

func (a *App) runLine(ctx context.Context) error {
	scanner := bufio.NewScanner(a.term.Stdin())
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("run tui: %w", err)
		}
		line := scanner.Text()
		if line == "/quit" || line == "/exit" {
			return a.writeErr
		}
		if err := a.submit(line); err != nil {
			return err
		}
		if a.shouldQuit() {
			return a.writeErr
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read tui input: %w", err)
	}
	return a.writeErr
}

func (a *App) runRaw(ctx context.Context) error {
	decoder := input.New(a.term.Stdin())
	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("run tui: %w", err)
		}
		ev, err := decoder.Read()
		if err != nil {
			if err == io.EOF {
				return a.writeErr
			}
			return fmt.Errorf("read tui input: %w", err)
		}
		if a.handleRawInput(ev.Data) {
			return a.writeErr
		}
		a.render()
		if a.shouldQuit() {
			return a.writeErr
		}
	}
}

func (a *App) handleRawInput(data string) bool {
	if a.handleModelMenuInput(data) {
		return false
	}
	if a.handleCommandMenuInput(data) {
		return false
	}
	switch data {
	case "\x1b[5~":
		a.ui.ScrollBy(8)
		return false
	case "\x1b[6~":
		a.ui.ScrollBy(-8)
		return false
	case "\x1b[F":
		a.ui.ScrollToBottom()
		return false
	case "\x04":
		if a.editor.Text() != "" {
			a.editor.HandleInput("\x1b[3~")
			return false
		}
		if a.abortActiveTurn() {
			return false
		}
		return true
	case "\x03":
		if a.editor.Text() != "" {
			a.editor.Clear()
			return false
		}
		if a.abortActiveTurn() {
			return false
		}
		return true
	case "\x1b":
		if a.abortActiveTurn() {
			return false
		}
		a.editor.Clear()
		return false
	case "\x0f":
		a.header.Toggle()
		return false
	case "\t":
		if a.completeCommand() {
			return false
		}
	}
	if isWheelUp(data) {
		a.ui.ScrollBy(3)
		return false
	}
	if isWheelDown(data) {
		a.ui.ScrollBy(-3)
		return false
	}
	a.editor.HandleInput(data)
	return false
}

func (a *App) completeCommand() bool {
	completion, ok := a.menu.Completion()
	if !ok {
		return false
	}
	a.editor.SetText(completion)
	return true
}

func (a *App) handleCommandMenuInput(data string) bool {
	if !a.menu.Visible() {
		return false
	}
	switch data {
	case "\x1b[A":
		return a.menu.Move(-1)
	case "\x1b[B":
		return a.menu.Move(1)
	case "\r", "\n":
		command, ok := a.menu.Selected()
		if !ok {
			return false
		}
		a.editor.Clear()
		if err := a.runSlashCommand(command.Name, ""); err != nil {
			a.appendError(err)
		}
		return true
	}
	return false
}

func isWheelUp(data string) bool {
	return strings.HasPrefix(data, "\x1b[<64;") || strings.HasPrefix(data, "\x1b[M`")
}

func isWheelDown(data string) bool {
	return strings.HasPrefix(data, "\x1b[<65;") || strings.HasPrefix(data, "\x1b[Ma")
}

func (a *App) startStatusTicker(ctx context.Context) func() {
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(140 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-stop:
				return
			case <-ticker.C:
				a.mu.Lock()
				active := a.status.Text() != ""
				if active {
					a.status.Step()
				}
				a.mu.Unlock()
				if active {
					a.render()
				}
			}
		}
	}()
	return func() { close(stop) }
}
