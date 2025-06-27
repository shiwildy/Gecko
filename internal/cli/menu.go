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
	const boxInternalWidth = 58
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
		col1TargetDisplayWidth := 26
		col2TargetDisplayWidth := 26
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
		text := strings.Join(content, " ")
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
	}

	fmt.Printf("   ║%s║\n", lineContent)
}

func DisplayMenu(apacheStatus, mysqlStatus, ngrokStatus bool) {
	clearScreen()
	apacheVersion := service.GetApacheVersion()
	mysqlVersion := service.GetMySQLVersion()
	phpVersion := service.GetPHPVersion()
	apachePort := service.GetApachePort()
	mysqlPort := service.GetMySQLPort()
	ngrokURL, privateURL := service.GetActiveNgrokURL()

	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println("             ██████╗ ███████╗ ██████╗ ██╗  ██╗ ██████╗ ")
	fmt.Println("            ██╔════╝ ██╔════╝██╔════╝ ██║  ██║██╔═══██╗")
	fmt.Println("            ██║  ███╗█████╗  ██║  ███╗█████║  ██║  ██║")
	fmt.Println("            ██║  ██║ ██╔══╝  ██║  ██║ ██╔══██║██║  ██║")
	fmt.Println("            ╚██████╔╝███████╗╚██████╔╝██║  ██║╚██████╔╝")
	fmt.Println("             ╚═════╝ ╚══════╝ ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ")
	fmt.Println("")
	fmt.Println(" A blazingly fast, CLI-based local environment for web development.")
	fmt.Println()

	fmt.Println("   ╔═══════════════════════ INFORMATION ══════════════════════╗")
	printRow(fmt.Sprintf("Gecko Version : %s1.0.0%s", shared.ColorGreen, shared.ColorReset))
	printRow(fmt.Sprintf("PHP (Active)  : %s%s%s", shared.ColorGreen, phpVersion, shared.ColorReset))
	printRow(fmt.Sprintf("Apache        : %s%s%s", shared.ColorGreen, apacheVersion, shared.ColorReset))
	printRow(fmt.Sprintf("MySQL         : %s%s%s", shared.ColorGreen, mysqlVersion, shared.ColorReset))
	fmt.Println("   ╟══════════════════════════ STATUS ════════════════════════╢")

	apacheStatusLine := fmt.Sprintf("Apache: %s%-10s%s | Port: %s%s%s",
		ternary(apacheStatus, shared.ColorGreen, shared.ColorRed),
		ternary(apacheStatus, "Running", "Stopped"),
		shared.ColorReset, shared.ColorGreen, apachePort, shared.ColorReset,
	)
	printRow(apacheStatusLine)

	mysqlStatusLine := fmt.Sprintf("MySQL:  %s%-10s%s | Port: %s%s%s",
		ternary(mysqlStatus, shared.ColorGreen, shared.ColorRed),
		ternary(mysqlStatus, "Running", "Stopped"),
		shared.ColorReset, shared.ColorGreen, mysqlPort, shared.ColorReset,
	)
	printRow(mysqlStatusLine)

	ngrokStatusLine := fmt.Sprintf("Ngrok:  %s%-10s%s",
		ternary(ngrokStatus, shared.ColorGreen, shared.ColorRed),
		ternary(ngrokStatus, "Active", "Inactive"),
		shared.ColorReset,
	)
	printRow(ngrokStatusLine)

	if ngrokURL != "" {
		printRow(fmt.Sprintf("Public URL: %s%s%s", shared.ColorGreen, ngrokURL, shared.ColorReset))
	}

	if privateURL != "" {
		printRow(fmt.Sprintf("Private URL: %s%s%s", shared.ColorYellow, privateURL, shared.ColorReset))
	} else {
		printRow("")
	}
	fmt.Println("   ╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("   ╔══════════════════════ GECKO MANAGER ═════════════════════╗")
	printRow(" ")
	printRow(fmt.Sprintf("%s:: SERVICES%s", shared.ColorYellow, shared.ColorReset))
	printRow(ternary(apacheStatus, "1. Stop Apache", "1. Start Apache"), ternary(mysqlStatus, "2. Stop MySQL", "2. Start MySQL"))

	printRow(" ")
	printRow(fmt.Sprintf("%s:: VHOST & DATABASE%s", shared.ColorYellow, shared.ColorReset))
	printRow("3. Reinitialize MySQL DB", "4. Create VHOST App")
	printRow("5. Delete VHOST App", " ")

	printRow(" ")
	printRow(fmt.Sprintf("%s:: TOOLS%s", shared.ColorYellow, shared.ColorReset))
	printRow("6. Install Root CA", "7. Switch PHP Version")
	printRow("8. Install default SSL", ternary(ngrokStatus, "9. Stop Ngrok Tunnel", "9. Start Ngrok Tunnel"))
	printRow("10. Set Ngrok Auth Token", " ")

	printRow(" ")
	printRow(fmt.Sprintf("%s:: APPLICATION%s", shared.ColorYellow, shared.ColorReset))
	printRow("x. Exit", " ")
	printRow(" ")
	fmt.Println("   ╚══════════════════════════════════════════════════════════╝")
}
