package service

import (
	"fmt"
	"gecko/internal/shared"
	"os/exec"
	"time"
)

const (
	apacheExe = `C:\Gecko\bin\httpd\bin\httpd.exe`
	apacheDir = `C:\Gecko\bin\httpd\`
)

func StartApache() {
	cmd := exec.Command(apacheExe, "-d", apacheDir)
	err := cmd.Start()
	if err != nil {
		fmt.Printf("%sError starting Apache: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}
	fmt.Printf("%sApache started in background.%s\n", shared.ColorGreen, shared.ColorReset)
}

func StopApache() {
	cmd := exec.Command("taskkill", "/F", "/IM", "httpd.exe")
	err := cmd.Run()
	if err == nil {
		fmt.Printf("%sApache stopped.%s\n", shared.ColorYellow, shared.ColorReset)
	}
}

func RestartApache() {
	fmt.Printf("%sRestarting Apache to apply changes...%s\n", shared.ColorYellow, shared.ColorReset)
	StopApache()
	time.Sleep(1 * time.Second)
	StartApache()
}
