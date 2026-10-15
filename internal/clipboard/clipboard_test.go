package clipboardutil

import (
	"errors"
	"testing"
)

func TestCopyToClipboardDegradesGracefully(t *testing.T) {
	err := CopyToClipboard("hello")

	if Available() {
		return
	}

	if !errors.Is(err, ErrUnsupported) {
		t.Errorf("without a clipboard, expected ErrUnsupported, got %v", err)
	}
}

func TestAvailableMatchesCopyOutcome(t *testing.T) {
	if !Available() && CopyToClipboard("x") == nil {
		t.Error("Available reported no clipboard but the copy succeeded")
	}
}
