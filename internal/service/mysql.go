package service

import (
	"bufio"
	"fmt"
	"gecko/internal/shared"
	"os"
	"os/exec"
	"strings"
)

const (
	mysqlExe          = `C:\Gecko\bin\mysql\bin\mysqld.exe`
	mysqlInstallDbExe = `C:\Gecko\bin\mysql\bin\mysql_install_db.exe`
	mysqlDataDir      = `C:\Gecko\etc\config\mysql\`
	mysqlLogError     = `C:\Gecko\logs\mysql\mysql_error.log`
	mysqlBinLog       = `C:\Gecko\logs\mysql\binlog`
)

func runMysqlInstallDb() bool {
	fmt.Printf("%sRunning mysql_install_db.exe...%s\n", shared.ColorYellow, shared.ColorReset)
	cmd := exec.Command(mysqlInstallDbExe, "--datadir="+mysqlDataDir)
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("%sError initializing MySQL database:%s\n", shared.ColorRed, shared.ColorReset)
		fmt.Println(string(output))
		return false
	}

	fmt.Printf("%sMySQL database initialized successfully.%s\n", shared.ColorGreen, shared.ColorReset)
	return true
}

func checkAndInitializeIfNeeded() bool {
	dir, err := os.ReadDir(mysqlDataDir)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(mysqlDataDir, os.ModePerm)
			dir = []os.DirEntry{} 
		} else {
			fmt.Printf("%sError reading data directory: %v%s\n", shared.ColorRed, err, shared.ColorReset)
			return false
		}
	}

	if len(dir) == 0 {
		fmt.Printf("%sMySQL data directory is empty. Automatically initializing...%s\n", shared.ColorYellow, shared.ColorReset)
		return runMysqlInstallDb()
	}

	return false
}

func StartMySQL() {
	checkAndInitializeIfNeeded()

	fmt.Printf("%sAttempting to start MySQL...%s\n", shared.ColorYellow, shared.ColorReset)
	cmd := exec.Command(mysqlExe,
		"--datadir="+mysqlDataDir,
		"--log-error="+mysqlLogError,
		"--log-bin="+mysqlBinLog,
	)
	err := cmd.Start()
	if err != nil {
		fmt.Printf("%sError starting MySQL: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}
	fmt.Printf("%sMySQL started in background.%s\n", shared.ColorGreen, shared.ColorReset)
}

func InitializeMySQL() {
	fmt.Printf("%sManual MySQL database initialization...%s\n", shared.ColorYellow, shared.ColorReset)

	dir, _ := os.ReadDir(mysqlDataDir)

	if len(dir) > 0 {
		fmt.Printf("%sWarning: The MySQL data directory is not empty.%s\n", shared.ColorRed, shared.ColorReset)
		fmt.Print(shared.ColorYellow, "Do you want to delete existing data and reinitialize? (y/n): ", shared.ColorReset)

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))

		if choice != "y" {
			fmt.Printf("%sMySQL initialization cancelled.%s\n", shared.ColorYellow, shared.ColorReset)
			return
		}

		os.RemoveAll(mysqlDataDir)
		os.MkdirAll(mysqlDataDir, os.ModePerm)
		fmt.Printf("%sData directory cleared.%s\n", shared.ColorGreen, shared.ColorReset)
	}

	runMysqlInstallDb()
}

func StopMySQL() {
	cmd := exec.Command("taskkill", "/F", "/IM", "mysqld.exe")
	err := cmd.Run()
	if err == nil {
		fmt.Printf("%sMySQL stopped.%s\n", shared.ColorYellow, shared.ColorReset)
	}
}