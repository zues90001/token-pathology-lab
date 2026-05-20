// Pathology Lab — corpus generator unit tests. No MiMo calls.
package lab

import "testing"

func TestSyntheticCorpusReturnsRequestedCount(t *testing.T) {
	tasks := SyntheticCorpus("code", 5)
	if got, want := len(tasks), 5; got != want {
		t.Fatalf("len(tasks) = %d, want %d", got, want)
	}
	for i, task := range tasks {
		if task.ID == "" {
			t.Errorf("task[%d] missing ID", i)
		}
		if task.Domain != "code" {
			t.Errorf("task[%d] domain = %q, want code", i, task.Domain)
		}
		if task.UserPrompt == "" {
			t.Errorf("task[%d] missing UserPrompt", i)
		}
	}
}

func TestSyntheticCorpusEmptyForZero(t *testing.T) {
	tasks := SyntheticCorpus("math", 0)
	if len(tasks) != 0 {
		t.Fatalf("expected empty slice, got %d", len(tasks))
	}
}
