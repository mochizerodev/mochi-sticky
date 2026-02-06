package board

import "testing"

func TestValidateBoardID(t *testing.T) {
	if err := validateBoardID(""); err == nil {
		t.Fatalf("expected error for empty id")
	}
	if err := validateBoardID("with/path"); err == nil {
		t.Fatalf("expected error for path separator")
	}
	if err := validateBoardID("with\\path"); err == nil {
		t.Fatalf("expected error for windows separator")
	}
	if err := validateBoardID("ok"); err != nil {
		t.Fatalf("unexpected error for valid id: %v", err)
	}
}

func TestGenerateBoardIDUniqueness(t *testing.T) {
	boards := []Board{{ID: "work"}, {ID: "work-1"}}
	if got := generateBoardID("Work", boards); got != "work-2" {
		t.Fatalf("expected work-2, got %s", got)
	}
}
