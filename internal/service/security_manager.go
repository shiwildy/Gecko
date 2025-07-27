package service

import (
	"fmt"
	"gecko/internal/shared"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func applyPostgresSecuritySettings(isDevMode bool) error {
	confPath := filepath.Join(pgsqlDataDir, "postgresql.conf")
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		return nil
	}

	content, err := os.ReadFile(confPath)
	if err != nil {
		return fmt.Errorf("could not read postgresql.conf: %w", err)
	}

	config, err := GetConfig()
	if err != nil {
		return err
	}

	var newListenAddress string
	if isDevMode {
		newListenAddress = "'*'"
	} else {
		newListenAddress = "'localhost'"
	}

	strContent := string(content)

	reListen := regexp.MustCompile(`(?m)^#?\s*listen_addresses\s*=\s*'.*?'`)
	if !reListen.MatchString(strContent) {
		strContent = "listen_addresses = " + newListenAddress + "\n" + strContent
	} else {
		strContent = reListen.ReplaceAllString(strContent, "listen_addresses = "+newListenAddress)
	}

	rePort := regexp.MustCompile(`(?m)^#?\s*port\s*=\s*\d+`)
	if !rePort.MatchString(strContent) {
		strContent = "port = " + config.PostgresPort + "\n" + strContent
	} else {
		strContent = rePort.ReplaceAllString(strContent, "port = "+config.PostgresPort)
	}

	return os.WriteFile(confPath, []byte(strContent), 0644)
}

func applyApacheSecuritySettings(isDevMode bool) error {
	var oldDirective, newDirective string
	if isDevMode {
		oldDirective = "Require local"
		newDirective = "Require all granted"
		fmt.Printf("%sApplying Development Mode (public access) to Apache...%s\n", shared.ColorYellow, shared.ColorReset)
	} else {
		oldDirective = "Require all granted"
		newDirective = "Require local"
		fmt.Printf("%sApplying Private Mode (local access only) to Apache...%s\n", shared.ColorYellow, shared.ColorReset)
	}

	vhostDir := `C:\Gecko\etc\config\httpd\sites-enabled`
	files, err := os.ReadDir(vhostDir)
	if err != nil {
		return fmt.Errorf("could not read vhost directory: %w", err)
	}
	re := regexp.MustCompile(oldDirective)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".conf") {
			filePath := filepath.Join(vhostDir, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}
			newContent := re.ReplaceAllString(string(content), newDirective)
			os.WriteFile(filePath, []byte(newContent), 0644)
		}
	}
	return nil
}

func ToggleDevelopmentMode() {
	isApacheRunning := IsServiceRunning("httpd.exe")
	isMySQLRunning := IsServiceRunning("mysqld.exe")
	isPgRunning := IsServiceRunning("postgres.exe")

	config, err := GetConfig()
	if err != nil {
		fmt.Printf("%sFailed to load configuration: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	newMode := !config.DevelopmentMode

	if err := applyApacheSecuritySettings(newMode); err != nil {
		fmt.Printf("%sFailed to apply Apache security settings: %v%s\n", shared.ColorRed, err, shared.ColorReset)
	}
	if err := applyPostgresSecuritySettings(newMode); err != nil {
		fmt.Printf("%sFailed to apply PostgreSQL security settings: %v%s\n", shared.ColorRed, err, shared.ColorReset)
	}

	config.DevelopmentMode = newMode
	if err := SaveConfig(config); err != nil {
		fmt.Printf("%sFailed to save new configuration: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	if newMode {
		fmt.Printf("%sDevelopment Mode activated. Services will be accessible from your local network.%s\n", shared.ColorGreen, shared.ColorReset)
	} else {
		fmt.Printf("%sPrivate Mode activated. Services will only be accessible from this computer.%s\n", shared.ColorGreen, shared.ColorReset)
	}

	if isApacheRunning {
		RestartApache()
	}
	if isMySQLRunning {
		StopMySQL()
		time.Sleep(1 * time.Second)
		StartMySQL()
	}
	if isPgRunning {
		RestartPostgreSQL()
	}
}
