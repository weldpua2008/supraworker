package model

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// DefaultPath returns PATH variable
func DefaultPath() string {
	defaultPath := "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

	switch runtime.GOOS {
	case "windows":
		defaultPath = "PATH=%PATH%;%SystemRoot%\\system32;%SystemRoot%;%SystemRoot%\\System32\\Wbem"
	}
	return defaultPath
}

// MergeEnvVars merges enviroments variables.
func MergeEnvVars(CmdENVs []string) (uniqueMergedENV []string) {

	mergedENV := append(CmdENVs, os.Environ()...)
	mergedENV = append(mergedENV, DefaultPath())
	unique := make(map[string]bool, len(mergedENV))
	for indx := range mergedENV {
		if len(strings.Split(mergedENV[indx], "=")) != 2 {
			continue
		}
		k := strings.Split(mergedENV[indx], "=")[0]
		if _, ok := unique[k]; !ok {
			uniqueMergedENV = append(uniqueMergedENV, mergedENV[indx])
			unique[k] = true
		}
		// if strings.HasPrefix(j.cmd.Env[indx],"PATH="){
		//     j.cmd.Env[indx]
		// }
	}
	return
}

// CmdWrapper wraps command.
func CmdWrapper(RunAs string, UseSHELL bool, CMD string) (shell string, args []string) {
	cmdSplitted := strings.Fields(CMD)
	if len(RunAs) > 1 {
		shell = "sudo"
		args = []string{"-u", RunAs, CMD}
		switch runtime.GOOS {
		case "windows":
			shell = "runas"
			args = []string{fmt.Sprintf("/user:%s", RunAs), CMD}
		default:
			if bash, err := exec.LookPath("su"); err == nil {
				shell = bash
				args = []string{"-", RunAs, "-c", CMD}
			}
		}

	} else if useCmdAsIs(CMD) {
		shell = cmdSplitted[0]
		args = cmdSplitted[1:]
	} else if UseSHELL {
		shell = "sh"
		args = []string{"-c", CMD}
		switch runtime.GOOS {
		case "windows":
			if ps, err := exec.LookPath("powershell.exe"); err == nil {
				args = []string{"-NoProfile", "-NonInteractive", CMD}
				shell = ps
			} else if bash, err := exec.LookPath("bash.exe"); err == nil {
				shell = bash
			} else {
				shell = "powershell.exe"
				args = []string{"-NoProfile", "-NonInteractive", CMD}
				log.Tracef("Can't fetch powershell nor bash, got %s\n", err)
			}

		default:
			if bash, err := exec.LookPath("bash"); err == nil {
				shell = bash
			}
		}
	} else {
		shell = cmdSplitted[0]
		args = cmdSplitted[1:]
	}
	return
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// useCmdAsIs returns true if cmd has shell
func useCmdAsIs(CMD string) bool {
	cmdSplitted := strings.Fields(CMD)

	if (len(cmdSplitted) > 0) && (fileExists(cmdSplitted[0])) {
		// in case of bash
		if strings.HasSuffix(cmdSplitted[0], "bash") {
			return true
			// in case su or sudo
			// } else if strings.HasSuffix(cmdSplitted[0], "su") || strings.HasSuffix(cmdSplitted[0], "sudo") {
			// 	return true
		} else if cmdSplitted[0] == "/bin/bash" || cmdSplitted[0] == "/bin/sh" {
			return true
		}
	} else if (cmdSplitted[0] == "bash") || (cmdSplitted[0] == "sh") {
		return true
		// } else if (cmdSplitted[0] == "sudo") || (cmdSplitted[0] == "su") {
		// 	return true
	}
	return false
}

func urlProvided(stage string) bool {
	url := viper.GetString(fmt.Sprintf("%s.url", stage))
	return len(url) >= 1
}
