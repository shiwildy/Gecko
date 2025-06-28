package cli

import (
	"fmt"
	"gecko/internal/service"
	"gecko/internal/shared"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

var ansi = regexp.MustCompile("\x1b\\[[0-9;]*m")

func stripAnsi(str string) string {
	return ansi.ReplaceAllString(str, "")
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

func ternary(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

func printRow(content ...string) {
	const boxInternalWidth = 62
	const horizontalPadding = 2
	var lineContent string

	if len(content) == 1 {
		text := content[0]
		displayLength := len(stripAnsi(text))
		availableTextWidth := boxInternalWidth - (horizontalPadding * 2)

		paddingNeeded := availableTextWidth - displayLength
		if paddingNeeded < 0 {
			paddingNeeded = 0
		}

		lineContent = fmt.Sprintf("%s%s%s",
			strings.Repeat(" ", horizontalPadding),
			text,
			strings.Repeat(" ", paddingNeeded+horizontalPadding),
		)
	} else if len(content) == 2 {
		text1 := content[0]
		text2 := content[1]
		col1TargetDisplayWidth := 28
		col2TargetDisplayWidth := 28
		spaceBetweenColumns := 2
		
		displayLength1 := len(stripAnsi(text1))
		padding1 := col1TargetDisplayWidth - displayLength1
		if padding1 < 0 {
			padding1 = 0
		}
		paddedText1 := fmt.Sprintf("%s%s", text1, strings.Repeat(" ", padding1))
		
		displayLength2 := len(stripAnsi(text2))
		padding2 := col2TargetDisplayWidth - displayLength2
		if padding2 < 0 {
			padding2 = 0
		}
		paddedText2 := fmt.Sprintf("%s%s", text2, strings.Repeat(" ", padding2))
		
		lineContent = fmt.Sprintf("%s%s%s%s%s",
			strings.Repeat(" ", horizontalPadding),
			paddedText1,
			strings.Repeat(" ", spaceBetweenColumns),
			paddedText2,
			strings.Repeat(" ", horizontalPadding),
		)
	} else {
		lineContent = fmt.Sprintf("%s%s", strings.Repeat(" ", horizontalPadding), strings.Join(content, " "))
	}

	fmt.Printf("   ║%s║\n", lineContent)
}

func DisplayMenu(apacheStatus, mysqlStatus, ngrokStatus, cloudflareStatus bool) {
	clearScreen()
	apacheVersion := service.GetApacheVersion()
	mysqlVersion := service.GetMySQLVersion()
	phpVersion := service.GetPHPVersion()

	var apachePortToDisplay string
	if apacheStatus {
		apachePortToDisplay = service.GetApachePort()
	} else {
		apachePortToDisplay = service.GetApachePort()
	}

	var mysqlPortToDisplay string
	if mysqlStatus {
		mysqlPortToDisplay = service.GetMySQLPort()
	} else {
		mysqlPortToDisplay = service.GetMySQLPort()
	}

	config, _ := service.GetConfig()
	devModeStatus := config.DevelopmentMode

	ngrokURL, ngrokLocalURL := service.GetActiveNgrokURL()
	cloudflareURL, cloudflareLocalURL := service.GetActiveCloudflareURL()

	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println("              ██████╗ ███████╗ ██████╗ ██╗  ██╗ ██████╗ ")
	fmt.Println("             ██╔════╝ ██╔════╝██╔════╝ ██║  ██║██╔═══██╗")
	fmt.Println("             ██║  ███╗█████╗  ██║  ███╗█████║  ██║   ██║")
	fmt.Println("             ██║  ██║ ██╔══╝  ██║  ██║ ██╔══██║██║   ██║")
	fmt.Println("             ╚██████╔╝███████╗╚██████╔╝██║  ██║╚██████╔╝")
	fmt.Println("              ╚═════╝ ╚══════╝ ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ")
	fmt.Println("")
	fmt.Println("  A blazingly fast, CLI-based local environment for web development.")
	fmt.Println()

	fmt.Println("   ╔═════════════════════════ INFORMATION ════════════════════════╗")
	printRow(fmt.Sprintf("Gecko Version : %s1.0.1%s", shared.ColorGreen, shared.ColorReset))
	printRow(fmt.Sprintf("PHP (Active)  : %s%s%s", shared.ColorGreen, phpVersion, shared.ColorReset))
	printRow(fmt.Sprintf("Apache        : %s%s%s", shared.ColorGreen, apacheVersion, shared.ColorReset))
	printRow(fmt.Sprintf("MySQL         : %s%s%s", shared.ColorGreen, mysqlVersion, shared.ColorReset))
	fmt.Println("   ╟════════════════════════════ STATUS ══════════════════════════╢")

	apacheStatusLine := fmt.Sprintf("Apache: %s%-10s%s | Port: %s%s%s",
		ternary(apacheStatus, shared.ColorGreen, shared.ColorRed),
		ternary(apacheStatus, "Running", "Stopped"),
		shared.ColorReset, shared.ColorGreen, apachePortToDisplay, shared.ColorReset,
	)
	printRow(apacheStatusLine)

	mysqlStatusLine := fmt.Sprintf("MySQL:  %s%-10s%s | Port: %s%s%s",
		ternary(mysqlStatus, shared.ColorGreen, shared.ColorRed),
		ternary(mysqlStatus, "Running", "Stopped"),
		shared.ColorReset, shared.ColorGreen, mysqlPortToDisplay, shared.ColorReset,
	)
	printRow(mysqlStatusLine)

	securityStatusLine := fmt.Sprintf("Security: %s%-15s%s",
		ternary(devModeStatus, shared.ColorRed, shared.ColorGreen),
		ternary(devModeStatus, "DEV MODE (Public)", "PRIVATE (Local)"),
		shared.ColorReset,
	)
	printRow(securityStatusLine)
	
	fmt.Println("   ╟────────────────────────── TUNNELS ───────────────────────────╢")

	ngrokStatusLine := fmt.Sprintf("Ngrok:  %s%-10s%s",
		ternary(ngrokStatus, shared.ColorGreen, shared.ColorRed),
		ternary(ngrokStatus, "Active", "Inactive"),
		shared.ColorReset,
	)
	printRow(ngrokStatusLine)
	if ngrokURL != "" {
		printRow(fmt.Sprintf(" Public URL: %s%s%s", shared.ColorGreen, ngrokURL, shared.ColorReset))
		printRow(fmt.Sprintf(" Local URL: %s%s%s", shared.ColorYellow, ngrokLocalURL, shared.ColorReset))
	}

	cloudflareStatusLine := fmt.Sprintf("Cloudflare: %s%-10s%s",
		ternary(cloudflareStatus, shared.ColorGreen, shared.ColorRed),
		ternary(cloudflareStatus, "Active", "Inactive"),
		shared.ColorReset,
	)
	printRow(cloudflareStatusLine)

	if cloudflareURL != "" {
		printRow(fmt.Sprintf(" Public URL: %s%s%s", shared.ColorGreen, cloudflareURL, shared.ColorReset))
		printRow(fmt.Sprintf(" Local URL: %s%s%s", shared.ColorYellow, cloudflareLocalURL, shared.ColorReset))
	}

	fmt.Println("   ╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("   ╔═══════════════════════ GECKO MANAGER ════════════════════════╗")
	printRow(" ")
	printRow(fmt.Sprintf("%s:: SERVICES%s", shared.ColorYellow, shared.ColorReset))
	printRow(ternary(apacheStatus, "1. Stop Apache", "1. Start Apache"), ternary(mysqlStatus, "2. Stop MySQL", "2. Start MySQL"))
	printRow(" ")

	printRow(fmt.Sprintf("%s:: CONFIG & APPS%s", shared.ColorYellow, shared.ColorReset))
	printRow("3. Create VHOST App", "4. Delete VHOST App")
	printRow("5. Reinitialize MySQL DB","6. Change Service Ports")
	printRow(" ")

	printRow(fmt.Sprintf("%s:: TOOLS & TUNNELS%s", shared.ColorYellow, shared.ColorReset))
	printRow("7. Switch PHP Version", "8. Install Root CA (SSL)")
	printRow("9. Install Default SSL", ternary(ngrokStatus, "10. Stop Ngrok", "10. Start Ngrok"))
	printRow("11. Set Ngrok Auth Token", ternary(cloudflareStatus, "12. Stop Cloudflare", "12. Start Cloudflare"))
	
	printRow(" ")
	printRow(fmt.Sprintf("%s:: APPLICATION%s", shared.ColorYellow, shared.ColorReset))
	printRow(ternary(devModeStatus, "13. Deactivate Dev Mode", "13. Activate Dev Mode"), "x. Exit")
	printRow(" ")
	fmt.Println("   ╚══════════════════════════════════════════════════════════════╝")
}