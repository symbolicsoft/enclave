// SPDX-FileCopyrightText: © 2024 Nadim Kobeissi <nadim@symbolic.software>
// SPDX-License-Identifier: GPL-2.0-only

package util

import (
	"os"
	"os/exec"
	"runtime"
)

func ClearManually() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cls")
	case "linux", "darwin":
		cmd = exec.Command("clear")
	default:
		// Unknown territory,
		// don't run commands we're not sure of.
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}
