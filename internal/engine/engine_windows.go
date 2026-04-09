package engine

import "golang.org/x/sys/windows"

func checkIsElevated() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}
