package httpbody

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestRead_RejectsOversizedBody(t *testing.T) {
	_, err := Read(strings.NewReader("12345"), 4)
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge, got %v", err)
	}
}

func TestLimit_StreamsWithinLimit(t *testing.T) {
	body, err := Limit(io.NopCloser(strings.NewReader("1234")), 4)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = body.Close() })
	got, err := io.ReadAll(body)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "1234" {
		t.Errorf("body = %q", got)
	}
}

func TestLimit_RejectsOversizedStream(t *testing.T) {
	body, err := Limit(io.NopCloser(strings.NewReader("12345")), 4)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = body.Close() })
	_, err = io.ReadAll(body)
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge, got %v", err)
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
