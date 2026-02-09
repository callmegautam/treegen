/*
Copyright © 2026 GAUTAM SUTHAR iamgautamsuthar@gmail.com
*/

package clipboardutil

import (
	"github.com/atotto/clipboard"
)

func CopyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}
