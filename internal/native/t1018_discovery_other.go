//go:build !windows

package native

import "fmt"

func init() {
	Register("T1018", func() (NativeResult, error) {
		return NativeResult{
			Success:     false,
			ErrorOutput: "T1018 remote system discovery requires Windows",
		}, fmt.Errorf("T1018: not supported on this platform")
	}, nil)
}
