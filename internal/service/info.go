package service

import (
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func getVersion(command string, args ...string) string {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "Not Found"
	}

	re := regexp.MustCompile(`\d+\.\d+(\.\d+)?`)
	match := re.FindString(string(output))
	if match == "" {
		return "Unknown Version"
	}
	return match
}

func getPIDsByProcessName(processName string) []string {
	cmd := exec.Command("tasklist", "/NH", "/FO", "CSV", "/FI", "IMAGENAME eq "+processName)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil
	}

	output := strings.TrimSpace(out.String())
	if output == "" {
		return nil
	}

	var pids []string
	re := regexp.MustCompile(`"([^"]+)"`)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := re.FindAllStringSubmatch(line, -1)
		if len(matches) > 1 {
			pids = append(pids, matches[1][1])
		}
	}
	return pids
}

func findPortsByPIDs(pids []string) string {
	if len(pids) == 0 {
		return "N/A"
	}

	cmd := exec.Command("netstat", "-aon")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "N/A"
	}

	output := out.String()
	portMap := make(map[string]bool)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		linePID := fields[len(fields)-1]
		for _, pid := range pids {
			if linePID == pid && (fields[3] == "LISTENING" || fields[3] == "ESTABLISHED") {
				address := fields[1]
				parts := strings.Split(address, ":")
				port := parts[len(parts)-1]
				if _, err := strconv.Atoi(port); err == nil {
					portMap[port] = true
				}
				break
			}
		}
	}

	if len(portMap) == 0 {
		return "N/A"
	}

	var ports []string
	for port := range portMap {
		ports = append(ports, port)
	}
	return strings.Join(ports, ", ")
}

func GetApacheVersion() string {
	return getVersion(`C:\Gecko\bin\httpd\bin\httpd.exe`, "-v")
}

func GetMySQLVersion() string {
	return getVersion(`C:\Gecko\bin\mysql\bin\mysqld.exe`, "--version")
}

func GetPHPVersion() string {
	return getVersion(`C:\Gecko\bin\php\php\php.exe`, "-v")
}

// rollbek use pid detect
func GetApachePort() string {
	pids := getPIDsByProcessName("httpd.exe")
	return findPortsByPIDs(pids)
}

func GetMySQLPort() string {
	pids := getPIDsByProcessName("mysqld.exe")
	return findPortsByPIDs(pids)
}