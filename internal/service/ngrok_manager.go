package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gecko/internal/shared"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	ngrokExe        = `C:\Gecko\bin\ngrok\ngrok.exe`
	ngrokAPIURL     = "http://127.0.0.1:4040/api/tunnels"
	ngrokConfigDir  = `C:\Gecko\etc\config\ngrok`
	ngrokConfigFile = `C:\Gecko\etc\config\ngrok\ngrok.yml`
)

var activeNgrokURL string
var activeTunURL string

type ngrokTunnelResponse struct {
	Tunnels []struct {
		Name      string `json:"name"`
		PublicURL string `json:"public_url"`
		Proto     string `json:"proto"`
	} `json:"tunnels"`
}

func GetActiveNgrokURL() (string, string) {
	if !IsServiceRunning("ngrok.exe") {
		activeNgrokURL = ""
		activeTunURL = ""
	}
	return activeNgrokURL, activeTunURL
}

func isAuthTokenSet() bool {
	if _, err := os.Stat(ngrokConfigFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func SetAuthToken(reader *bufio.Reader) {
	if !IsNgrokInstalled() {
		fmt.Printf("%sError: ngrok.exe not found. This feature is disabled.%s\n", shared.ColorRed, shared.ColorReset)
		return
	}
	if err := os.MkdirAll(ngrokConfigDir, os.ModePerm); err != nil {
		fmt.Printf("%sError creating ngrok config directory: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}
	fmt.Print(shared.ColorYellow, "Enter your Ngrok authtoken: ", shared.ColorReset)
	token, _ := reader.ReadString('\n')
	token = strings.TrimSpace(token)
	if token == "" {
		fmt.Println("Operation cancelled.")
		return
	}
	cmd := exec.Command(ngrokExe, "config", "add-authtoken", token, "--config", ngrokConfigFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%sError setting authtoken: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		fmt.Println(string(output))
		return
	}
	fmt.Printf("%sNgrok authtoken saved successfully to %s%s\n", shared.ColorGreen, ngrokConfigFile, shared.ColorReset)
}

func StartNgrokTunnel(localDomain string) {
	fmt.Printf("%sAttempting to start Ngrok tunnel for %s...%s\n", shared.ColorYellow, localDomain, shared.ColorReset)

	if !IsNgrokInstalled() {
		fmt.Printf("%sError: ngrok.exe not found at %s%s\n", shared.ColorRed, ngrokExe, shared.ColorReset)
		return
	}

	if !isAuthTokenSet() {
		fmt.Printf("%sError: Ngrok authtoken is not set.%s\n", shared.ColorRed, shared.ColorReset)
		fmt.Println("Please set authtoken to use this features.")
		return
	}

	config, err := GetConfig()
	if err != nil {
		fmt.Printf("%sCould not load config for Ngrok: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}
	
	apachePort := config.ApachePort
	fmt.Printf("%sForwarding to port %s (from gecko-config.json)%s\n", shared.ColorYellow, apachePort, shared.ColorReset)

	baseArgs := []string{"http", apachePort, "--host-header=" + localDomain, "--config", ngrokConfigFile}
	cmd := exec.Command(ngrokExe, baseArgs...)
	err = cmd.Start()
	if err != nil {
		fmt.Printf("%sError starting Ngrok: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	fmt.Printf("%sNgrok process started in the background.%s\n", shared.ColorGreen, shared.ColorReset)
	fmt.Printf("%sWaiting for tunnel to establish...%s\n", shared.ColorYellow, shared.ColorReset)

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		time.Sleep(2 * time.Second)
		fmt.Printf("%sPinging Ngrok API (attempt %d/%d)...%s\n", shared.ColorYellow, i+1, maxRetries, shared.ColorReset)

		resp, err := http.Get(ngrokAPIURL)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		var tunnelData ngrokTunnelResponse
		if err := json.Unmarshal(body, &tunnelData); err != nil {
			continue
		}

		for _, tunnel := range tunnelData.Tunnels {
			if strings.HasPrefix(tunnel.PublicURL, "https") {
				activeNgrokURL = tunnel.PublicURL
				activeTunURL = localDomain
				fmt.Printf("%sTunnel established!%s\n", shared.ColorGreen, shared.ColorReset)
				fmt.Printf("Private URL: %s%s%s\n", shared.ColorGreen, activeTunURL, shared.ColorReset)
				fmt.Printf("Public URL: %s%s%s\n", shared.ColorGreen, activeNgrokURL, shared.ColorReset)
				return
			}
		}
	}

	fmt.Printf("%sCould not establish tunnel after %d attempts.%s\n", shared.ColorRed, maxRetries, shared.ColorReset)
	fmt.Println("Please check the Ngrok agent manually at http://127.0.0.1:4040")
}


func StopNgrokTunnels() {
	fmt.Printf("%sStopping all Ngrok tunnels...%s\n", shared.ColorYellow, shared.ColorReset)
	cmd := exec.Command("taskkill", "/F", "/IM", "ngrok.exe")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%sNo running Ngrok processes found or could not stop them.%s\n", shared.ColorYellow, shared.ColorReset)
	} else {
		fmt.Printf("%sAll Ngrok tunnels stopped successfully.%s\n", shared.ColorGreen, shared.ColorReset)
	}
	activeNgrokURL = ""
}

func IsNgrokInstalled() bool {
	if _, err := os.Stat(ngrokExe); os.IsNotExist(err) {
		return false
	}
	return true
}