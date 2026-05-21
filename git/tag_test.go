package git

import (
	"testing"
)

// --- TestParseVersion ---

func TestParseVersion_Normal3Digits(t *testing.T) {
	prefix, digits, err := ParseVersion("v0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prefix != "v" {
		t.Errorf("prefix = %q, want %q", prefix, "v")
	}
	if len(digits) != 3 || digits[0] != 0 || digits[1] != 0 || digits[2] != 1 {
		t.Errorf("digits = %v, want [0 0 1]", digits)
	}
}

func TestParseVersion_TwoDigits(t *testing.T) {
	prefix, digits, err := ParseVersion("v0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prefix != "v" {
		t.Errorf("prefix = %q, want %q", prefix, "v")
	}
	if len(digits) != 2 || digits[0] != 0 || digits[1] != 1 {
		t.Errorf("digits = %v, want [0 1]", digits)
	}
}

func TestParseVersion_OneDigit(t *testing.T) {
	prefix, digits, err := ParseVersion("v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prefix != "v" {
		t.Errorf("prefix = %q, want %q", prefix, "v")
	}
	if len(digits) != 1 || digits[0] != 1 {
		t.Errorf("digits = %v, want [1]", digits)
	}
}

func TestParseVersion_FourDigits(t *testing.T) {
	prefix, digits, err := ParseVersion("v0.1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prefix != "v" {
		t.Errorf("prefix = %q, want %q", prefix, "v")
	}
	if len(digits) != 4 || digits[0] != 0 || digits[1] != 1 || digits[2] != 2 || digits[3] != 3 {
		t.Errorf("digits = %v, want [0 1 2 3]", digits)
	}
}

func TestParseVersion_FiveDigits(t *testing.T) {
	_, digits, err := ParseVersion("v2.3.4.5.6")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(digits) != 5 || digits[0] != 2 || digits[1] != 3 || digits[2] != 4 {
		t.Errorf("digits = %v, want [2 3 4 5 6]", digits)
	}
}

func TestParseVersion_TPrefix(t *testing.T) {
	prefix, digits, err := ParseVersion("t0.1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prefix != "t" {
		t.Errorf("prefix = %q, want %q", prefix, "t")
	}
	if len(digits) != 4 || digits[0] != 0 || digits[1] != 1 || digits[2] != 2 || digits[3] != 3 {
		t.Errorf("digits = %v, want [0 1 2 3]", digits)
	}
}

func TestParseVersion_NonNumericSegment(t *testing.T) {
	prefix, digits, err := ParseVersion("v1.abc.2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prefix != "v" {
		t.Errorf("prefix = %q, want %q", prefix, "v")
	}
	// "abc" should be ignored
	if len(digits) != 2 || digits[0] != 1 || digits[1] != 2 {
		t.Errorf("digits = %v, want [1 2]", digits)
	}
}

func TestParseVersion_EmptyTag(t *testing.T) {
	_, _, err := ParseVersion("")
	if err == nil {
		t.Error("expected error for empty tag")
	}
}

func TestParseVersion_NoLetterPrefix(t *testing.T) {
	_, _, err := ParseVersion("1.2.3")
	if err == nil {
		t.Error("expected error for tag without letter prefix")
	}
}

// --- TestNormalizeDigits ---

func TestNormalizeDigits_Truncate(t *testing.T) {
	result := NormalizeDigits([]int{1, 2, 3, 4, 5}, 3)
	if len(result) != 3 || result[0] != 1 || result[1] != 2 || result[2] != 3 {
		t.Errorf("NormalizeDigits([1 2 3 4 5], 3) = %v, want [1 2 3]", result)
	}
}

func TestNormalizeDigits_Pad(t *testing.T) {
	result := NormalizeDigits([]int{1}, 3)
	if len(result) != 3 || result[0] != 1 || result[1] != 0 || result[2] != 0 {
		t.Errorf("NormalizeDigits([1], 3) = %v, want [1 0 0]", result)
	}
}

func TestNormalizeDigits_PadTwo(t *testing.T) {
	result := NormalizeDigits([]int{0, 1}, 3)
	if len(result) != 3 || result[0] != 0 || result[1] != 1 || result[2] != 0 {
		t.Errorf("NormalizeDigits([0 1], 3) = %v, want [0 1 0]", result)
	}
}

func TestNormalizeDigits_ExactLength(t *testing.T) {
	result := NormalizeDigits([]int{1, 2, 3}, 3)
	if len(result) != 3 || result[0] != 1 || result[1] != 2 || result[2] != 3 {
		t.Errorf("NormalizeDigits([1 2 3], 3) = %v, want [1 2 3]", result)
	}
}

func TestNormalizeDigits_Empty(t *testing.T) {
	result := NormalizeDigits(nil, 3)
	if len(result) != 3 || result[0] != 0 || result[1] != 0 || result[2] != 0 {
		t.Errorf("NormalizeDigits(nil, 3) = %v, want [0 0 0]", result)
	}
}

// --- TestNextVPatch ---

func TestNextVPatch_Normal(t *testing.T) {
	major, minor, patch := NextVPatch(0, 0, 1)
	if major != 0 || minor != 0 || patch != 2 {
		t.Errorf("NextVPatch(0,0,1) = (%d,%d,%d), want (0,0,2)", major, minor, patch)
	}
}

func TestNextVPatch_Overflow(t *testing.T) {
	major, minor, patch := NextVPatch(0, 0, 99)
	if major != 0 || minor != 1 || patch != 0 {
		t.Errorf("NextVPatch(0,0,99) = (%d,%d,%d), want (0,1,0)", major, minor, patch)
	}
}

func TestNextVPatch_MinorOverflow(t *testing.T) {
	major, minor, patch := NextVPatch(0, 1, 99)
	if major != 0 || minor != 2 || patch != 0 {
		t.Errorf("NextVPatch(0,1,99) = (%d,%d,%d), want (0,2,0)", major, minor, patch)
	}
}

func TestNextVPatch_MajorOverflow(t *testing.T) {
	major, minor, patch := NextVPatch(3, 99, 99)
	if major != 3 || minor != 100 || patch != 0 {
		t.Errorf("NextVPatch(3,99,99) = (%d,%d,%d), want (3,100,0)", major, minor, patch)
	}
}

func TestNextVPatch_ZeroBase(t *testing.T) {
	major, minor, patch := NextVPatch(0, 0, 0)
	if major != 0 || minor != 0 || patch != 1 {
		t.Errorf("NextVPatch(0,0,0) = (%d,%d,%d), want (0,0,1)", major, minor, patch)
	}
}

// --- TestFormatVTag ---

func TestFormatVTag(t *testing.T) {
	got := FormatVTag(0, 1, 5)
	if got != "v0.1.5" {
		t.Errorf("FormatVTag(0,1,5) = %q, want %q", got, "v0.1.5")
	}
}

func TestFormatVTag_Large(t *testing.T) {
	got := FormatVTag(3, 99, 99)
	if got != "v3.99.99" {
		t.Errorf("FormatVTag(3,99,99) = %q, want %q", got, "v3.99.99")
	}
}

// --- TestFormatTTag ---

func TestFormatTTag(t *testing.T) {
	got := FormatTTag(0, 1, 2, 3)
	if got != "t0.1.2.3" {
		t.Errorf("FormatTTag(0,1,2,3) = %q, want %q", got, "t0.1.2.3")
	}
}

func TestFormatTTag_ZeroIter(t *testing.T) {
	got := FormatTTag(0, 1, 2, 0)
	if got != "t0.1.2.0" {
		t.Errorf("FormatTTag(0,1,2,0) = %q, want %q", got, "t0.1.2.0")
	}
}
