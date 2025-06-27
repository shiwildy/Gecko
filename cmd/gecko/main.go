package main

import (
	"bufio"
	"fmt"
	"gecko/internal/cli"
	"gecko/internal/service"
	"gecko/internal/shared"
	"gecko/internal/utils"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func main() {
	utils.CheckAndRequestAdmin()
	
	_, err := service.LoadConfig()
	if err != nil {
		fmt.Printf("%sFatal Error: Could not load or create configuration file: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		fmt.Println("Press Enter to exit.")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}
	
	mainMenu()
}

func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func mainMenu() {
	reader := bufio.NewReader(os.Stdin)

	for {
		apacheStatus := service.IsServiceRunning("httpd.exe")
		mysqlStatus := service.IsServiceRunning("mysqld.exe")
		ngrokStatus := service.IsServiceRunning("ngrok.exe")
		cloudflareStatus := service.IsServiceRunning("cloudflared.exe")
		cli.DisplayMenu(apacheStatus, mysqlStatus, ngrokStatus, cloudflareStatus)

		fmt.Print(shared.ColorYellow, "\nEnter your choice: ", shared.ColorReset)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		clearScreen()
		switch choice {
		case "1":
			if apacheStatus {
				service.StopApache()
			} else {
				service.StartApache()
			}
		case "2":
			if mysqlStatus {
				service.StopMySQL()
			} else {
				service.StartMySQL()
			}
		case "3":
			handleCreateVHost(reader)
		case "4":
			handleDeleteVHost(reader)
		case "5":
			service.InitializeMySQL()
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "6":
			service.ChangeServicePorts(reader)
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "7":
			service.SwitchPHPVersion(reader)
		case "8":
			service.InstallGeckoRootCA()
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "9":
			if ngrokStatus {
				service.StopNgrokTunnels()
			} else {
				handleStartTunnel(reader, "ngrok")
			}
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "10":
			service.SetAuthToken(reader)
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "11":
			if cloudflareStatus {
				service.StopCloudflareTunnel()
			} else {
				handleStartTunnel(reader, "cloudflare")
			}
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "12":
			service.GenerateDefaultCertificate()
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "x", "X":
			fmt.Println(shared.ColorYellow, "\nStopping all services...", shared.ColorReset)
			service.StopApache()
			service.StopMySQL()
			service.StopNgrokTunnels()
			service.StopCloudflareTunnel()
			fmt.Println(shared.ColorGreen, "Bye!", shared.ColorReset)
			return
		default:
			fmt.Println(shared.ColorRed, "Invalid choice. Please try again.", shared.ColorReset)
		}
		time.Sleep(1 * time.Second)
	}
}

func handleCreateVHost(reader *bufio.Reader) {
	fmt.Print(shared.ColorYellow, "Enter the new domain name (e.g., mysite.test): ", shared.ColorReset)
	domainName, _ := reader.ReadString('\n')
	domainName = strings.TrimSpace(domainName)

	vhostConfigFile := filepath.Join(`C:\Gecko\etc\config\httpd\sites-enabled`, domainName+".conf")
	replaceChoice := ""

	if _, err := os.Stat(vhostConfigFile); err == nil {
		fmt.Printf("%sWarning: VHost for '%s' already exists.%s\n", shared.ColorRed, domainName, shared.ColorReset)
		fmt.Print(shared.ColorYellow, "Do you want to replace it and format its directory? (y/n): ", shared.ColorReset)
		replaceChoice, _ = reader.ReadString('\n')
		replaceChoice = strings.TrimSpace(strings.ToLower(replaceChoice))
		if replaceChoice != "y" {
			fmt.Println(shared.ColorYellow, "Operation cancelled.", shared.ColorReset)
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
			return
		}
	}

	service.CreateVirtualHost(domainName, replaceChoice)
	fmt.Println("\nPress Enter to continue...")
	reader.ReadString('\n')
}

func handleDeleteVHost(reader *bufio.Reader) {
	vhosts, err := service.ListVirtualHosts()
	if err != nil {
		fmt.Printf("%sError listing virtual hosts: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		fmt.Println("\nPress Enter to continue...")
		reader.ReadString('\n')
		return
	}

	if len(vhosts) == 0 {
		fmt.Println(shared.ColorYellow, "No deletable virtual hosts found.", shared.ColorReset)
		fmt.Println("\nPress Enter to continue...")
		reader.ReadString('\n')
		return
	}

	fmt.Println(shared.ColorGreen, "Select a virtual host to delete:", shared.ColorReset)
	for i, vhost := range vhosts {
		fmt.Printf("%d. %s\n", i+1, vhost)
	}
	fmt.Println("0. Cancel")

	fmt.Print(shared.ColorYellow, "\nEnter your choice: ", shared.ColorReset)
	choiceStr, _ := reader.ReadString('\n')
	choice, err := strconv.Atoi(strings.TrimSpace(choiceStr))

	if err != nil || choice < 0 || choice > len(vhosts) {
		fmt.Println(shared.ColorRed, "Invalid choice.", shared.ColorReset)
		fmt.Println("\nPress Enter to continue...")
		reader.ReadString('\n')
		return
	}

	if choice == 0 {
		fmt.Println(shared.ColorYellow, "Delete cancelled.", shared.ColorReset)
		fmt.Println("\nPress Enter to continue...")
		reader.ReadString('\n')
		return
	}

	domainToDelete := vhosts[choice-1]

	fmt.Printf("%sAre you sure you want to permanently delete '%s' and all its files? (y/n): %s", shared.ColorRed, domainToDelete, shared.ColorReset)
	confirm, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(confirm)) == "y" {
		service.DeleteVirtualHost(domainToDelete)
	} else {
		fmt.Println(shared.ColorYellow, "Delete cancelled.", shared.ColorReset)
	}
	fmt.Println("\nPress Enter to continue...")
	reader.ReadString('\n')
}

func handleStartTunnel(reader *bufio.Reader, tunnelType string) {
	vhosts, err := service.ListVirtualHosts()
	if err != nil {
		fmt.Printf("%sError listing virtual hosts: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}
	
	vhosts = append([]string{"localhost"}, vhosts...)

	if len(vhosts) == 0 {
		fmt.Println(shared.ColorYellow, "No virtual hosts found to create a tunnel for.", shared.ColorReset)
		return
	}

	fmt.Printf("%sSelect a host to expose via %s:%s\n", shared.ColorGreen, strings.Title(tunnelType), shared.ColorReset)
	for i, vhost := range vhosts {
		fmt.Printf("%d. %s\n", i+1, vhost)
	}
	fmt.Println("0. Cancel")

	fmt.Print(shared.ColorYellow, "\nEnter your choice: ", shared.ColorReset)
	choiceStr, _ := reader.ReadString('\n')
	choice, err := strconv.Atoi(strings.TrimSpace(choiceStr))

	if err != nil || choice <= 0 || choice > len(vhosts) {
		fmt.Println(shared.ColorRed, "Invalid choice.", shared.ColorReset)
		return
	}

	domainToTunnel := vhosts[choice-1]

	if tunnelType == "ngrok" {
		service.StartNgrokTunnel(domainToTunnel)
	} else if tunnelType == "cloudflare" {
		service.StartCloudflareTunnel(domainToTunnel)
	}
}