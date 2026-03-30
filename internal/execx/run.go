package execx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	gerr "github.com/ipolyzos/git-read-mcp/internal/errors"
)

// Result holds captured output and process outcome metadata.
type Result struct {
	Stdout    []byte
	Stderr    []byte
	ExitCode  int
	TimedOut  bool
	Truncated bool
}

// ExecSummary converts Result into a value suitable for gerr.FromExecResult.
func (r Result) ExecSummary() gerr.ExecResult {
	return gerr.ExecResult{
		TimedOut:  r.TimedOut,
		Truncated: r.Truncated,
		ExitCode:  r.ExitCode,
		Stderr:    r.Stderr,
	}
}

type sharedCap struct {
	max   int64
	used  int64
	trunc bool
	mu    sync.Mutex
}

type cappedWriter struct {
	buf    *bytes.Buffer
	shared *sharedCap
}

func (w *cappedWriter) Write(p []byte) (int, error) {
	if w.shared.max <= 0 {
		_, err := w.buf.Write(p)
		return len(p), err
	}
	w.shared.mu.Lock()
	defer w.shared.mu.Unlock()
	remain := w.shared.max - w.shared.used
	if remain <= 0 {
		w.shared.trunc = true
		return len(p), nil
	}
	take := int64(len(p))
	if take > remain {
		take = remain
		w.shared.trunc = true
	}
	if take > 0 {
		w.buf.Write(p[:take])
		w.shared.used += take
	}
	if int64(len(p)) > take {
		w.shared.trunc = true
	}
	return len(p), nil
}

// Run executes cmd with args (no shell). env selects the child environment: nil inherits
// the current process environment; an empty slice starts the child with no inherited vars.
// cwd is the working directory; empty means the current directory.
// timeout bounds wall time when greater than zero (in addition to ctx).
// maxBytes caps total bytes stored from stdout and stderr combined; zero or negative means unlimited.
//
// Errors are returned only for invalid input, start failures, copy failures (excluding context
// cancellation), or unexpected wait errors. A non-zero process exit populates ExitCode with a nil error.
// DeadlineExceeded sets TimedOut and returns ErrExecTimeout. Output truncation sets Truncated with nil error.
func Run(ctx context.Context, cmd string, args []string, env []string, cwd string, timeout time.Duration, maxBytes int64) (Result, error) {
	var zero Result
	if cmd == "" {
		return zero, fmt.Errorf("execx: command is empty")
	}

	runCtx := ctx
	var cancel context.CancelFunc
	if timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	if cancel != nil {
		defer cancel()
	}

	c := exec.CommandContext(runCtx, cmd, args...)
	c.Dir = cwd
	if env != nil {
		c.Env = env
	}

	stdoutPipe, err := c.StdoutPipe()
	if err != nil {
		return zero, fmt.Errorf("execx: stdout pipe: %w", err)
	}
	stderrPipe, err := c.StderrPipe()
	if err != nil {
		return zero, fmt.Errorf("execx: stderr pipe: %w", err)
	}

	cap := &sharedCap{max: maxBytes}
	var outBuf, errBuf bytes.Buffer
	outW := cappedWriter{buf: &outBuf, shared: cap}
	errW := cappedWriter{buf: &errBuf, shared: cap}

	errCh := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, copyErr := io.Copy(&outW, stdoutPipe)
		errCh <- copyErr
	}()
	go func() {
		defer wg.Done()
		_, copyErr := io.Copy(&errW, stderrPipe)
		errCh <- copyErr
	}()

	if err := c.Start(); err != nil {
		return zero, fmt.Errorf("execx: start: %w", err)
	}

	wg.Wait()
	var copyErr error
	for range 2 {
		if err := <-errCh; err != nil && !errors.Is(err, context.Canceled) && copyErr == nil {
			copyErr = err
		}
	}
	if copyErr != nil {
		_ = c.Wait()
		return zero, fmt.Errorf("execx: copy output: %w", copyErr)
	}

	waitErr := c.Wait()
	res := Result{
		Stdout:    bytes.Clone(outBuf.Bytes()),
		Stderr:    bytes.Clone(errBuf.Bytes()),
		Truncated: cap.trunc,
	}

	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		res.TimedOut = true
		return res, gerr.ErrExecTimeout
	}

	if errors.Is(waitErr, context.Canceled) {
		return res, waitErr
	}

	var exitErr *exec.ExitError
	if errors.As(waitErr, &exitErr) {
		res.ExitCode = exitErr.ExitCode()
		return res, nil
	}
	if waitErr != nil {
		return res, fmt.Errorf("execx: wait: %w", waitErr)
	}

	res.ExitCode = 0
	return res, nil
}
