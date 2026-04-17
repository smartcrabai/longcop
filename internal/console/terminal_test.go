package console

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestReadFeatureRequestRequiresContent(t *testing.T) {
	t.Parallel()

	terminal := New(strings.NewReader("   \n"), strings.NewReader(""), &bytes.Buffer{})
	_, err := terminal.ReadFeatureRequest()
	if !errors.Is(err, ErrEmptyFeatureRequest) {
		t.Fatalf("expected ErrEmptyFeatureRequest, got %v", err)
	}
}

func TestAskYesNoAndChoiceUseInteractiveReader(t *testing.T) {
	t.Parallel()

	output := &bytes.Buffer{}
	terminal := New(strings.NewReader("feature"), strings.NewReader("\n2\n"), output)

	accepted, err := terminal.AskYesNo("Implement with TDD?", true)
	if err != nil {
		t.Fatalf("AskYesNo returned error: %v", err)
	}
	if !accepted {
		t.Fatal("expected default yes answer")
	}

	selected, err := terminal.AskChoice("Select the mode:", []string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("AskChoice returned error: %v", err)
	}
	if selected != 1 {
		t.Fatalf("expected option index 1, got %d", selected)
	}
}
