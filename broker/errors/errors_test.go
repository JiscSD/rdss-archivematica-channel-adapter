package errors

import (
	"errors"
	"fmt"
	"testing"
)

func ExampleNew() {
	var err error
	x, y := 1, 2
	if x != y {
		err = New(APPERRMET001, "something went wrong")
	}
	fmt.Println(err)
	// Output: [APPERRMET001]: something went wrong
}

func TestNew_UnknownKind(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("New() with unknown kind did not panic")
		}
	}()
	_ = New(50, "")
}

func TestNew(t *testing.T) {
	const desc = "description"
	tests := []struct {
		k    Kind
		want string
	}{
		{GENERR001, "[GENERR001]: description"},
		{GENERR002, "[GENERR002]: description"},
		{GENERR003, "[GENERR003]: description"},
		{GENERR004, "[GENERR004]: description"},
		{GENERR005, "[GENERR005]: description"},
		{GENERR006, "[GENERR006]: description"},
		{GENERR007, "[GENERR007]: description"},
		{GENERR008, "[GENERR008]: description"},
		{GENERR009, "[GENERR009]: description"},
		{GENERR010, "[GENERR010]: description"},
		{APPERRMET001, "[APPERRMET001]: description"},
		{APPERRMET002, "[APPERRMET002]: description"},
		{APPERRMET003, "[APPERRMET003]: description"},
	}
	for _, tt := range tests {
		e := New(tt.k, desc)
		if got := e.Error(); got != tt.want {
			t.Fatalf("New.Error() failed: want %v; got %v", tt.want, got)
		}
	}
}

func TestNewWithError(t *testing.T) {
	var (
		want = "[APPERRMET001]: description"
		err  = errors.New("description")
		e    = NewWithError(APPERRMET001, err)
	)
	if got := e.Error(); got != want {
		t.Errorf("Unexpected error; got %s, want %s", got, want)
	}
}
