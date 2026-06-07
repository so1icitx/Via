package auth

import (
	"encoding/json"
	"fmt"
	"io"
)

func decodeJSON(r io.Reader, dst any) error {
	if err := json.NewDecoder(r).Decode(dst); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	return nil
}