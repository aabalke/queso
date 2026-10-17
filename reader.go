package queso

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

func NewInstReaderFromStrings(s []string) (*bytes.Reader, error) {
	clean := strings.Join(strings.Fields(strings.Join(s, " ")), "")
	buf, err := hex.DecodeString(clean)
	if err != nil {
		return nil, fmt.Errorf("cli input args are not valid hex: %w", err)
	}

	return bytes.NewReader(buf), nil
}
