/*
Copyright © 2026 GAUTAM SUTHAR iamgautamsuthar@gmail.com
*/

package clipboardutil

import (
	"errors"

	"github.com/atotto/clipboard"
)

var ErrUnsupported = errors.New("no clipboard available on this system")

func Available() bool {
	return !clipboard.Unsupported
}

func CopyToClipboard(text string) error {
	if clipboard.Unsupported {
		return ErrUnsupported
	}
	if err := clipboard.WriteAll(text); err != nil {
		return err
	}
	return nil
}
