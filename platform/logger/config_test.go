package logger

import "testing"

func TestFormatUnmarshalText(t *testing.T) {
	var format Format
	if err := format.UnmarshalText([]byte("json")); err != nil {
		t.Fatalf("UnmarshalText() error = %v", err)
	}
	if format != FormatJSON {
		t.Fatalf("format = %q", format)
	}
}

func TestFormatUnmarshalTextRejectsUnknown(t *testing.T) {
	var format Format
	if err := format.UnmarshalText([]byte("xml")); err == nil {
		t.Fatal("UnmarshalText() error = nil")
	}
}
