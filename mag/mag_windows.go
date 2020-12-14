// +build windows

package mag

// #######
// https://studygolang.com/articles/7440
//
// https://gist.github.com/ptflp/ff56b78dfc30c6d2d2044c432a6f2dad
//
// https://stackoverflow.com/q/39598382
// #######

import (
	. "github.com/CodyGuo/win"
)

func Reboot() {
	getPrivileges()
	ExitWindowsEx(EWX_REBOOT, 0)
}

func Shutdown() {
	getPrivileges()
	ExitWindowsEx(EWX_SHUTDOWN, 0)
}

func getPrivileges() {
	var hToken HANDLE
	var tkp TOKEN_PRIVILEGES

	OpenProcessToken(GetCurrentProcess(), TOKEN_ADJUST_PRIVILEGES|TOKEN_QUERY, &hToken)
	LookupPrivilegeValueA(nil, StringToBytePtr(SE_SHUTDOWN_NAME), &tkp.Privileges[0].Luid)
	tkp.PrivilegeCount = 1
	tkp.Privileges[0].Attributes = SE_PRIVILEGE_ENABLED
	AdjustTokenPrivileges(hToken, false, &tkp, 0, nil, nil)
}
