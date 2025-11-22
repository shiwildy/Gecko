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
		pgStatus := service.IsServiceRunning("postgres.exe")
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
			if pgStatus {
				service.StopPostgreSQL()
			} else {
				service.StartPostgreSQL()
			}
		case "4":
			service.InitializePostgreSQL(reader)
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "5":
			handleCreateVHost(reader)
		case "6":
			handleDeleteVHost(reader)
		case "7":
			service.InitializeMySQL()
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "8":
			service.ChangeServicePorts(reader)
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "9":
			service.ViewPostgresPassword()
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "10":
			handleCreateProject(reader)
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "11":
			handleListProjects(reader)
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "12":
			handleDeleteProject(reader)
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "13":
			service.SwitchPHPVersion(reader)
		case "14":
			service.InstallGeckoRootCA()
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "15":
			service.GenerateDefaultCertificate()
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "16":
			if ngrokStatus {
				service.StopNgrokTunnels()
			} else {
				handleStartTunnel(reader, "ngrok")
			}
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "17":
			service.SetAuthToken(reader)
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "18":
			if cloudflareStatus {
				service.StopCloudflareTunnel()
			} else {
				handleStartTunnel(reader, "cloudflare")
			}
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "19":
			service.ToggleDevelopmentMode()
			fmt.Println("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "x", "X":
			fmt.Println(shared.ColorYellow, "\nStopping all services...", shared.ColorReset)
			service.StopApache()
			service.StopMySQL()
			service.StopPostgreSQL()
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

func handleCreateProject(reader *bufio.Reader) {
	fmt.Println(shared.ColorGreen, "\n╔═══════════════════ PROJECT CREATOR ═══════════════════╗", shared.ColorReset)
	fmt.Println("║", shared.ColorWhite, " Select a technology stack:", shared.ColorReset, "                      ║")
	fmt.Println("║                                                        ║")
	fmt.Println("║  1. PHP Projects                                       ║")
	fmt.Println("║  2. JavaScript Projects                                ║")
	fmt.Println("║  3. Static HTML                                        ║")
	fmt.Println("║  0. Cancel                                             ║")
	fmt.Println("║                                                        ║")
	fmt.Println(shared.ColorGreen, "╚════════════════════════════════════════════════════════╝", shared.ColorReset)

	fmt.Print(shared.ColorYellow, "\nEnter your choice: ", shared.ColorReset)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "0":
		fmt.Println(shared.ColorYellow, "Project creation cancelled.", shared.ColorReset)
		return
	case "1":
		handlePHPProjectCreation(reader)
	case "2":
		handleJavaScriptProjectCreation(reader)
	case "3":
		handleStaticProjectCreation(reader)
	default:
		fmt.Println(shared.ColorRed, "Invalid choice.", shared.ColorReset)
		return
	}
}

func handlePHPProjectCreation(reader *bufio.Reader) {
	fmt.Println(shared.ColorGreen, "\n╔═══════════════════ PHP FRAMEWORKS ═══════════════════╗", shared.ColorReset)
	fmt.Println("║", shared.ColorWhite, " Select a PHP framework:", shared.ColorReset, "                         ║")
	fmt.Println("║                                                        ║")
	fmt.Println("║  1. Laravel - Modern PHP framework for web artisans   ║")
	fmt.Println("║  2. WordPress - Popular CMS platform                  ║")
	fmt.Println("║  3. Symfony - High performance PHP framework          ║")
	fmt.Println("║  4. CodeIgniter - Simple & elegant PHP framework      ║")
	fmt.Println("║  5. CakePHP - Rapid development framework             ║")
	fmt.Println("║  6. Laminas - Enterprise-ready PHP framework          ║")
	fmt.Println("║  0. Back to main menu                                  ║")
	fmt.Println("║                                                        ║")
	fmt.Println(shared.ColorGreen, "╚════════════════════════════════════════════════════════╝", shared.ColorReset)

	fmt.Print(shared.ColorYellow, "\nEnter your choice: ", shared.ColorReset)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	var projectType string
	switch choice {
	case "0":
		handleCreateProject(reader)
		return
	case "1":
		projectType = "laravel"
	case "2":
		projectType = "wordpress"
	case "3":
		projectType = "symfony"
	case "4":
		projectType = "codeigniter"
	case "5":
		projectType = "cakephp"
	case "6":
		projectType = "laminas"
	default:
		fmt.Println(shared.ColorRed, "Invalid choice.", shared.ColorReset)
		return
	}

	createProjectWithType(projectType, reader)
}

func handleJavaScriptProjectCreation(reader *bufio.Reader) {
	fmt.Println(shared.ColorGreen, "\n╔═══════════════════ JS FRAMEWORKS ═══════════════════╗", shared.ColorReset)
	fmt.Println("║", shared.ColorWhite, " Select a JavaScript framework:", shared.ColorReset, "                 ║")
	fmt.Println("║                                                        ║")
	fmt.Println("║  1. React - Library for building user interfaces      ║")
	fmt.Println("║  2. Vue.js - Progressive JavaScript framework         ║")
	fmt.Println("║  3. Next.js - React framework for production          ║")
	fmt.Println("║  4. Nuxt.js - Vue.js framework for production         ║")
	fmt.Println("║  5. Angular - Platform for mobile & desktop web apps  ║")
	fmt.Println("║  6. Svelte - Cybernetically enhanced web apps         ║")
	fmt.Println("║  7. Astro - Build faster websites                     ║")
	fmt.Println("║  8. Vite - Next generation frontend tooling           ║")
	fmt.Println("║  0. Back to main menu                                  ║")
	fmt.Println("║                                                        ║")
	fmt.Println(shared.ColorGreen, "╚════════════════════════════════════════════════════════╝", shared.ColorReset)

	fmt.Print(shared.ColorYellow, "\nEnter your choice: ", shared.ColorReset)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	var projectType string
	switch choice {
	case "0":
		handleCreateProject(reader)
		return
	case "1":
		projectType = "react"
	case "2":
		projectType = "vue"
	case "3":
		projectType = "nextjs"
	case "4":
		projectType = "nuxtjs"
	case "5":
		projectType = "angular"
	case "6":
		projectType = "svelte"
	case "7":
		projectType = "astro"
	case "8":
		projectType = "vite"
	default:
		fmt.Println(shared.ColorRed, "Invalid choice.", shared.ColorReset)
		return
	}

	createProjectWithType(projectType, reader)
}

func handleStaticProjectCreation(reader *bufio.Reader) {
	createProjectWithType("static", reader)
}

func createProjectWithType(projectType string, reader *bufio.Reader) {
	fmt.Print(shared.ColorYellow, "Enter project name (e.g., myapp): ", shared.ColorReset)
	projectName, _ := reader.ReadString('\n')
	projectName = strings.TrimSpace(projectName)

	if projectName == "" {
		fmt.Println(shared.ColorRed, "Project name cannot be empty.", shared.ColorReset)
		return
	}

	service.CreateProject(projectType, projectName, reader)
}

func handleListProjects(reader *bufio.Reader) {
	projects, err := service.ListProjects()
	if err != nil {
		fmt.Printf("%sError listing projects: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	if len(projects) == 0 {
		fmt.Println(shared.ColorYellow, "No projects found.", shared.ColorReset)
		return
	}

	fmt.Println(shared.ColorGreen, "\n╔═══════════════════ YOUR PROJECTS ═══════════════════╗", shared.ColorReset)
	for i, project := range projects {
		fmt.Printf("║ %d. %-45s ║\n", i+1, project)
	}
	fmt.Println(shared.ColorGreen, "╚══════════════════════════════════════════════════════╝", shared.ColorReset)
}

func handleDeleteProject(reader *bufio.Reader) {
	projects, err := service.ListProjects()
	if err != nil {
		fmt.Printf("%sError listing projects: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	if len(projects) == 0 {
		fmt.Println(shared.ColorYellow, "No projects found to delete.", shared.ColorReset)
		return
	}

	fmt.Println(shared.ColorGreen, "Select a project to delete:", shared.ColorReset)
	for i, project := range projects {
		fmt.Printf("%d. %s\n", i+1, project)
	}
	fmt.Println("0. Cancel")

	fmt.Print(shared.ColorYellow, "\nEnter your choice: ", shared.ColorReset)
	choiceStr, _ := reader.ReadString('\n')
	choice, err := strconv.Atoi(strings.TrimSpace(choiceStr))

	if err != nil || choice < 0 || choice > len(projects) {
		fmt.Println(shared.ColorRed, "Invalid choice.", shared.ColorReset)
		return
	}

	if choice == 0 {
		fmt.Println(shared.ColorYellow, "Delete cancelled.", shared.ColorReset)
		return
	}

	projectToDelete := projects[choice-1]

	fmt.Printf("%sAre you sure you want to permanently delete '%s' and all its files? (y/n): %s", shared.ColorRed, projectToDelete, shared.ColorReset)
	confirm, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(confirm)) == "y" {
		service.DeleteProject(projectToDelete, reader)
	} else {
		fmt.Println(shared.ColorYellow, "Delete cancelled.", shared.ColorReset)
	}
}
