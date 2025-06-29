package service

import (
	"fmt"
	"os/exec"
	"strings"
)

// used to check services with tasklists
func IsServiceRunning(serviceName string) bool {
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", serviceName))
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(strings.ToLower(string(output)), strings.ToLower(serviceName))
}
