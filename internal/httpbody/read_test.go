package httpbody

import (
	"strings"
	"testing"
)

func TestRead_RejectsOversizedBody(t *testing.T) {
	_, err := Read(strings.NewReader("12345"), 4)
	if err == nil {
		t.Fatal("expected response size error")
	}
}

func TestRead_ReturnsBodyWithinLimit(t *testing.T) {
	got, err := Read(strings.NewReader("1234"), 4)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "1234" {
		t.Errorf("Read() = %q", got)
	}
}

func TestRead_RejectsInvalidLimit(t *testing.T) {
	_, err := Read(strings.NewReader(""), 0)
	if err == nil {
		t.Fatal("expected invalid limit error")
	}
}
