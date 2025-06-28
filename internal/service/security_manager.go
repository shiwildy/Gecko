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

func applySecuritySettingsToVhosts(isDevMode bool) error {
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
				fmt.Printf("%sSkipping %s (read error): %v%s\n", shared.ColorYellow, file.Name(), err, shared.ColorReset)
				continue
			}

			newContent := re.ReplaceAllString(string(content), newDirective)

			if err = os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
				fmt.Printf("%sError writing to %s: %v%s\n", shared.ColorRed, file.Name(), err, shared.ColorReset)
			}
		}
	}
	return nil
}

func ToggleDevelopmentMode() {
	isApacheRunning := IsServiceRunning("httpd.exe")
	isMySQLRunning := IsServiceRunning("mysqld.exe")

	config, err := GetConfig()
	if err != nil {
		fmt.Printf("%sFailed to load configuration: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	newMode := !config.DevelopmentMode

	if err := applySecuritySettingsToVhosts(newMode); err != nil {
		fmt.Printf("%sFailed to apply Apache security settings: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
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
		fmt.Println(shared.ColorYellow + "Restarting Apache to apply changes..." + shared.ColorReset)
		RestartApache()
	}

	if isMySQLRunning {
		fmt.Println(shared.ColorYellow + "Restarting MySQL to apply changes..." + shared.ColorReset)
		StopMySQL()
		time.Sleep(1 * time.Second)
		StartMySQL()
	}
}