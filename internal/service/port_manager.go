package service

import (
	"bufio"
	"fmt"
	"gecko/internal/shared"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func ChangeServicePorts(reader *bufio.Reader) {
	config, err := GetConfig()
	if err != nil {
		fmt.Printf("%sFailed to load configuration: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	fmt.Println(shared.ColorGreen, "Select a service to change its port:", shared.ColorReset)
	fmt.Printf("1. Apache (HTTP: %s, HTTPS: %s)\n", config.ApachePort, config.ApacheSSLPort)
	fmt.Printf("2. MySQL (Current: %s)\n", config.MySQLPort)
	fmt.Printf("3. PostgreSQL (Current: %s)\n", config.PostgresPort)
	fmt.Println("x. Back to main menu")

	fmt.Print(shared.ColorYellow, "\nEnter your choice: ", shared.ColorReset)
	choiceStr, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(choiceStr)

	switch choice {
	case "1":
		changeApachePorts(reader, config)
		if IsServiceRunning("httpd.exe") {
			RestartApache()
		}
	case "2":
		changeMySQLPort(reader, config)
		if IsServiceRunning("mysqld.exe") {
			StopMySQL()
			StartMySQL()
		}
	case "3":
		changePostgresPort(reader, config)
		if IsServiceRunning("postgres.exe") {
			RestartPostgreSQL()
		}
	case "x":
		fmt.Println("Returning to main menu.")
	default:
		fmt.Println(shared.ColorRed, "Invalid choice.", shared.ColorReset)
	}
}

func changeApachePorts(reader *bufio.Reader, config *Config) {
	oldPortHTTP := config.ApachePort
	oldPortSSL := config.ApacheSSLPort

	fmt.Printf(shared.ColorYellow+"Enter new HTTP port (current: %s): "+shared.ColorReset, oldPortHTTP)
	newPortHTTP, _ := reader.ReadString('\n')
	newPortHTTP = strings.TrimSpace(newPortHTTP)
	if newPortHTTP == "" {
		newPortHTTP = oldPortHTTP
	}

	fmt.Printf(shared.ColorYellow+"Enter new HTTPS/SSL port (current: %s): "+shared.ColorReset, oldPortSSL)
	newPortSSL, _ := reader.ReadString('\n')
	newPortSSL = strings.TrimSpace(newPortSSL)
	if newPortSSL == "" {
		newPortSSL = oldPortSSL
	}

	if newPortHTTP == oldPortHTTP && newPortSSL == oldPortSSL {
		fmt.Println("Ports are the same. Operation cancelled.")
		return
	}

	fmt.Printf("%sUpdating Apache configuration files...%s\n", shared.ColorYellow, shared.ColorReset)

	// change apache http port
	httpPortConfPath := `C:\Gecko\etc\config\httpd\httpd.conf`
	updateFileWithPatterns(httpPortConfPath, map[string]string{
		`(?m)^Listen\s+` + oldPortHTTP: "Listen " + newPortHTTP,
	})

	// change apache https port
	sslConfPath := `C:\Gecko\etc\config\httpd\httpd-ssl.conf`
	updateFileWithPatterns(sslConfPath, map[string]string{
		`(?m)^Listen\s+` + oldPortSSL:              "Listen " + newPortSSL,
		`<VirtualHost\s+[^:]+:` + oldPortSSL + `>`: "<VirtualHost _default_:" + newPortSSL + ">",
	})

	vhostDir := `C:\Gecko\etc\config\httpd\sites-enabled`
	files, err := os.ReadDir(vhostDir)
	if err != nil {
		fmt.Printf("%sError reading vhost directory: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".conf") {
			filePath := filepath.Join(vhostDir, file.Name())
			fmt.Printf("Updating %s...\n", file.Name())
			patterns := map[string]string{
				`\*:` + oldPortHTTP: "*:" + newPortHTTP,
				`\*:` + oldPortSSL:  "*:" + newPortSSL,
			}
			updateFileWithPatterns(filePath, patterns)
		}
	}

	config.ApachePort = newPortHTTP
	config.ApacheSSLPort = newPortSSL
	if err := SaveConfig(config); err != nil {
		fmt.Printf("%sFailed to save new port configuration: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	fmt.Printf("%sApache ports updated successfully in all config files.%s\n", shared.ColorGreen, shared.ColorReset)
	fmt.Println(shared.ColorYellow + "Restart Apache to apply the new ports." + shared.ColorReset)
}

func changeMySQLPort(reader *bufio.Reader, config *Config) {
	currentPort := config.MySQLPort
	fmt.Printf(shared.ColorYellow+"Enter new port for MySQL (current: %s): "+shared.ColorReset, currentPort)
	newPortStr, _ := reader.ReadString('\n')
	newPortStr = strings.TrimSpace(newPortStr)

	if newPortStr == "" || newPortStr == currentPort {
		fmt.Println("Operation cancelled or port is the same.")
		return
	}

	config.MySQLPort = newPortStr
	if err := SaveConfig(config); err != nil {
		fmt.Printf("%sFailed to save new MySQL port configuration: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		config.MySQLPort = currentPort
		return
	}

	phpMyAdminConfigPath := `C:\Gecko\etc\phpmyadmin\config.inc.php`
	if _, err := os.Stat(phpMyAdminConfigPath); err == nil {
		fmt.Printf("%sUpdating phpMyAdmin configuration...%s\n", shared.ColorYellow, shared.ColorReset)

		content, err := os.ReadFile(phpMyAdminConfigPath)
		if err != nil {
			fmt.Printf("%sFailed to read phpMyAdmin config: %v%s\n", shared.ColorRed, err, shared.ColorReset)
			return
		}

		lines := strings.Split(string(content), "\n")
		var resultLines []string
		portLineFound := false

		for _, line := range lines {
			match, _ := regexp.MatchString(`^\s*\$cfg\['Servers'\]\[\$i\]\['port'\]`, line)
			if match {
				resultLines = append(resultLines, fmt.Sprintf("$cfg['Servers'][$i]['port'] = '%s';", newPortStr))
				portLineFound = true
			} else {
				resultLines = append(resultLines, line)
			}
		}

		if !portLineFound {
			resultLines = append(resultLines, fmt.Sprintf("\n$cfg['Servers'][$i]['port'] = '%s';", newPortStr))
		}

		output := strings.Join(resultLines, "\n")
		os.WriteFile(phpMyAdminConfigPath, []byte(output), 0644)
	}

	fmt.Printf("%sMySQL port updated to %s in gecko-config.json.%s\n", shared.ColorGreen, newPortStr, shared.ColorReset)
	fmt.Println(shared.ColorYellow + "Restart MySQL to apply the new port." + shared.ColorReset)
}

func changePostgresPort(reader *bufio.Reader, config *Config) {
	currentPort := config.PostgresPort
	fmt.Printf(shared.ColorYellow+"Enter new port for PostgreSQL (current: %s): "+shared.ColorReset, currentPort)
	newPortStr, _ := reader.ReadString('\n')
	newPortStr = strings.TrimSpace(newPortStr)

	if newPortStr == "" || newPortStr == currentPort {
		fmt.Println("Operation cancelled or port is the same.")
		return
	}

	config.PostgresPort = newPortStr
	if err := SaveConfig(config); err != nil {
		fmt.Printf("%sFailed to save new PostgreSQL port configuration: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		config.PostgresPort = currentPort
		return
	}

	applyPostgresSecuritySettings(config.DevelopmentMode)

	fmt.Printf("%sPostgreSQL port has been updated to %s.%s\n", shared.ColorGreen, newPortStr, shared.ColorReset)
	fmt.Println(shared.ColorYellow + "Restart PostgreSQL to apply the new port." + shared.ColorReset)
}

func updateFileWithPatterns(filePath string, patterns map[string]string) {
	input, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("%sSkipping %s (read error): %v%s\n", shared.ColorYellow, filePath, err, shared.ColorReset)
		return
	}
	content := string(input)

	for pattern, replacement := range patterns {
		re := regexp.MustCompile(pattern)
		content = re.ReplaceAllString(content, replacement)
	}

	if err = os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("%sError writing to %s: %v%s\n", shared.ColorRed, filePath, err, shared.ColorReset)
	}
}
