//go:build !cgo

package gemini

import (
	"errors"
)

type NeedleEngine struct{}

func GetNeedleEngine() (*NeedleEngine, error) {
	return nil, errors.New("needle C++ engine requires CGO (running in pure Go TLS mode)")
}

func (n *NeedleEngine) CompleteTools(prompt string, toolsJSON string) (*NeedleResponse, error) {
	return nil, errors.New("needle C++ engine not compiled in non-cgo build")
}
