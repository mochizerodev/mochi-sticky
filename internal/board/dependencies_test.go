package board

import "testing"

func TestIsReadyAndValidateNoCycles(t *testing.T) {
	tasks := []Task{
		{ID: "A", Status: "todo", DependsOn: []string{"B"}},
		{ID: "B", Status: "done"},
	}
	index := map[string]Task{
		"A": tasks[0],
		"B": tasks[1],
	}
	ready, unmet := IsReady(tasks[0], index)
	if !ready || len(unmet) != 0 {
		t.Fatalf("expected ready with no unmet, got %v %v", ready, unmet)
	}
	if err := ValidateNoCycles(tasks); err != nil {
		t.Fatalf("expected no cycles, got %v", err)
	}
}

func TestValidateNoCyclesDetectsCycle(t *testing.T) {
	tasks := []Task{
		{ID: "A", DependsOn: []string{"B"}},
		{ID: "B", DependsOn: []string{"A"}},
	}
	if err := ValidateNoCycles(tasks); err == nil {
		t.Fatalf("expected cycle error")
	}
}
