//go:build !windows

package engine

func checkIsElevated() bool {
	return true // permissive: don't skip elevation-gated techniques on non-Windows
}
