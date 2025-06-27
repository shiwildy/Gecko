package service

import (
	"fmt"
	"gecko/internal/shared"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
)

const (
	cloudflaredExe     = `C:\Gecko\bin\cloudflared\cloudflared.exe`
	cloudflaredLogFile = `C:\Gecko\logs\cloudflared.log`
)

var activeCloudflareURL string
var activeCloudflareLocalURL string

func IsCloudflaredInstalled() bool {
	if _, err := os.Stat(cloudflaredExe); os.IsNotExist(err) {
		return false
	}
	return true
}

func StartCloudflareTunnel(localDomain string) {
	if !IsCloudflaredInstalled() {
		fmt.Printf("%sError: cloudflared.exe is not found at its expected location.%s\n", shared.ColorRed, shared.ColorReset)
		fmt.Printf("%sPlease ensure it is placed at: %s%s\n", shared.ColorYellow, cloudflaredExe, shared.ColorReset)
		return
	}

	config, err := GetConfig()
	if err != nil {
		fmt.Printf("%sCould not load config for Cloudflare: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	apachePort := config.ApachePort
	targetURL := fmt.Sprintf("http://127.0.0.1:%s", apachePort)
	fmt.Printf("%sAttempting to start Cloudflare Tunnel for %s (Host: %s)...%s\n", shared.ColorYellow, targetURL, localDomain, shared.ColorReset)
	os.MkdirAll(filepath.Dir(cloudflaredLogFile), os.ModePerm)

	cmd := exec.Command(cloudflaredExe, "tunnel", "--url", targetURL, "--http-host-header", localDomain, "--logfile", cloudflaredLogFile, "--no-autoupdate", "--edge-ip-version", "4")
	err = cmd.Start()
	if err != nil {
		fmt.Printf("%sError starting Cloudflare Tunnel: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	fmt.Printf("%sCloudflare Tunnel process started in the background.%s\n", shared.ColorGreen, shared.ColorReset)
	fmt.Printf("%sWaiting for tunnel to establish...%s\n", shared.ColorYellow, shared.ColorReset)

	maxRetries := 8
	for i := 0; i < maxRetries; i++ {
		time.Sleep(2 * time.Second)
		logContent, err := os.ReadFile(cloudflaredLogFile)
		if err != nil {
			continue
		}

		re := regexp.MustCompile(`(https://[a-zA-Z0-9-]+.trycloudflare.com)`)
		matches := re.FindStringSubmatch(string(logContent))

		if len(matches) > 1 {
			activeCloudflareURL = matches[1]
			activeCloudflareLocalURL = localDomain
			fmt.Printf("%sTunnel established!%s\n", shared.ColorGreen, shared.ColorReset)
			fmt.Printf("Public URL: %s%s%s\n", shared.ColorGreen, activeCloudflareURL, shared.ColorReset)
			return
		}
	}

	fmt.Printf("%sCould not establish tunnel after multiple attempts.%s\n", shared.ColorRed, shared.ColorReset)
	fmt.Println("Please check the log file for details:", cloudflaredLogFile)
}

func StopCloudflareTunnel() {
	fmt.Printf("%sStopping Cloudflare Tunnel...%s\n", shared.ColorYellow, shared.ColorReset)
	cmd := exec.Command("taskkill", "/F", "/IM", "cloudflared.exe")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%sNo running Cloudflare processes found or could not stop them.%s\n", shared.ColorYellow, shared.ColorReset)
	} else {
		fmt.Printf("%sCloudflare Tunnel stopped successfully.%s\n", shared.ColorGreen, shared.ColorReset)
	}
	activeCloudflareURL = ""
	activeCloudflareLocalURL = ""
	os.Remove(cloudflaredLogFile)
}

func GetActiveCloudflareURL() (string, string) {
	if !IsServiceRunning("cloudflared.exe") {
		activeCloudflareURL = ""
		activeCloudflareLocalURL = ""
	}
	return activeCloudflareURL, activeCloudflareLocalURL
}