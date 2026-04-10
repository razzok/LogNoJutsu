//go:build !windows

package native

import "fmt"

func init() {
	Register("T1046", func() (NativeResult, error) {
		return NativeResult{
			Success:     false,
			ErrorOutput: "T1046 network scan requires Windows",
		}, fmt.Errorf("T1046: not supported on this platform")
	}, nil)
}
