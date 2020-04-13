package model

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
)

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

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
