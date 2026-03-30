package validate

import (
	"errors"
	"testing"
)

func TestMaxStringBytes(t *testing.T) {
	t.Parallel()
	if err := MaxStringBytes("field", "abc", 3); err != nil {
		t.Fatal(err)
	}
	if err := MaxStringBytes("field", "abcd", 3); !errors.Is(err, ErrStringTooLong) {
		t.Fatalf("want ErrStringTooLong, got %v", err)
	}
	if err := MaxStringBytes("field", "long", -1); err != nil {
		t.Fatal(err)
	}
}

func TestIntInRange(t *testing.T) {
	t.Parallel()
	if err := IntInRange("n", 5, 0, 10); err != nil {
		t.Fatal(err)
	}
	if err := IntInRange("n", 11, 0, 10); !errors.Is(err, ErrIntOutOfRange) {
		t.Fatalf("got %v", err)
	}
}

func TestBoolDefault(t *testing.T) {
	t.Parallel()
	tru := true
	if got := BoolDefault(nil, false); got != false {
		t.Fatalf("got %v", got)
	}
	if got := BoolDefault(&tru, false); got != true {
		t.Fatalf("got %v", got)
	}
}
