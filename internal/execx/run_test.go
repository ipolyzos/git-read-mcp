package execx

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	gerr "github.com/ipolyzos/git-read-mcp/internal/errors"
)

func TestMain(m *testing.M) {
	switch os.Getenv("EXECX_HELPER") {
	case "bigout":
		n, _ := strconv.Atoi(os.Getenv("EXECX_BYTES"))
		if n <= 0 {
			n = 1000
		}
		fmt.Print(strings.Repeat("x", n))
		os.Exit(0)
	case "bigboth":
		n, _ := strconv.Atoi(os.Getenv("EXECX_BYTES"))
		if n <= 0 {
			n = 1000
		}
		half := n / 2
		fmt.Print(strings.Repeat("o", half))
		fmt.Fprint(os.Stderr, strings.Repeat("e", n-half))
		os.Exit(0)
	case "exit":
		code, _ := strconv.Atoi(os.Getenv("EXECX_CODE"))
		os.Exit(code)
	case "sleep":
		time.Sleep(5 * time.Second)
		os.Exit(0)
	default:
		os.Exit(m.Run())
	}
}

func TestRun_nonzeroExit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	env := append(os.Environ(), "EXECX_HELPER=exit", "EXECX_CODE=7")
	res, err := Run(ctx, os.Args[0], []string{"-test.v=false"}, env, "", 10*time.Second, -1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.ExitCode != 7 {
		t.Fatalf("exit code: got %d", res.ExitCode)
	}
	if nz := gerr.FromExecResult(res.ExecSummary()); nz == nil {
		t.Fatal("expected FromExecResult for non-zero exit")
	}
}

func TestRun_truncation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	env := append(os.Environ(), "EXECX_HELPER=bigout", "EXECX_BYTES=20000")
	const maxBytes int64 = 500
	res, err := Run(ctx, os.Args[0], []string{"-test.v=false"}, env, "", 10*time.Second, maxBytes)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !res.Truncated {
		t.Fatal("expected truncation")
	}
	if got := int64(len(res.Stdout) + len(res.Stderr)); got > maxBytes {
		t.Fatalf("captured %d bytes, cap %d", got, maxBytes)
	}
	if !gerr.IsTruncated(gerr.FromExecResult(res.ExecSummary())) {
		t.Fatal("expected truncated mapping")
	}
}

func TestRun_truncationSharedBudgetBothStreams(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	env := append(os.Environ(), "EXECX_HELPER=bigboth", "EXECX_BYTES=20000")
	const maxBytes int64 = 400
	res, err := Run(ctx, os.Args[0], []string{"-test.v=false"}, env, "", 10*time.Second, maxBytes)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !res.Truncated {
		t.Fatal("expected truncation")
	}
	if got := int64(len(res.Stdout) + len(res.Stderr)); got > maxBytes {
		t.Fatalf("captured %d bytes, cap %d", got, maxBytes)
	}
}

func TestRun_timeout(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	env := append(os.Environ(), "EXECX_HELPER=sleep")
	res, err := Run(ctx, os.Args[0], []string{"-test.v=false"}, env, "", 200*time.Millisecond, -1)
	if !errors.Is(err, gerr.ErrExecTimeout) {
		t.Fatalf("want timeout error; err=%v", err)
	}
	if !res.TimedOut {
		t.Fatal("expected TimedOut flag")
	}
	if !gerr.IsTimeout(err) {
		t.Fatal("IsTimeout should match")
	}
}

func TestRun_emptyCommand(t *testing.T) {
	t.Parallel()
	_, err := Run(context.Background(), "", nil, nil, "", time.Second, -1)
	if err == nil {
		t.Fatal("expected error")
	}
}
