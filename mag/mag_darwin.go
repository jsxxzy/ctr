// +build darwin

package mag

import (
	"os/exec"
)

// Shutdown 关机
func Shutdown() {
	var cmd = exec.Command("/bin/sh", "-c", "sudo shutdown now")
	cmd.Run()
}

// Reboot 重启
func Reboot() {
	var cmd = exec.Command("/bin/sh", "-c", "sudo reboot")
	cmd.Run()
}
