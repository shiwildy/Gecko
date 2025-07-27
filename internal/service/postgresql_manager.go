package service

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"gecko/internal/shared"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	pgsqlBinDir  = `C:\Gecko\bin\pgsql\bin`
	pgsqlDataDir = `C:\Gecko\etc\config\data\pgsql`
	initdbExe    = `C:\Gecko\bin\pgsql\bin\initdb.exe`
	pgctlExe     = `C:\Gecko\bin\pgsql\bin\pg_ctl.exe`
	psqlExe      = `C:\Gecko\bin\pgsql\bin\psql.exe`
	pgsqlLogFile = `C:\Gecko\logs\pgsql.log`
)

func isPostgreSQLInitialized() bool {
	if _, err := os.Stat(pgsqlDataDir); os.IsNotExist(err) {
		return false
	}
	dir, err := os.ReadDir(pgsqlDataDir)
	if err != nil {
		return false
	}
	return len(dir) > 0
}

func generateRandomPassword(length int) (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		result[i] = chars[num.Int64()]
	}
	return string(result), nil
}

func InitializePostgreSQL(reader *bufio.Reader) {
	if _, err := os.Stat(initdbExe); os.IsNotExist(err) {
		fmt.Printf("%sError: initdb.exe not found at %s%s\n", shared.ColorRed, initdbExe, shared.ColorReset)
		return
	}

	if isPostgreSQLInitialized() {
		fmt.Printf("%sPostgreSQL data directory is not empty. Re-initialization will delete all existing data.%s\n", shared.ColorRed, shared.ColorReset)
		fmt.Print(shared.ColorYellow, "Are you sure you want to continue? (y/n): ", shared.ColorReset)
		confirm, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(confirm)) != "y" {
			fmt.Println("Initialization cancelled.")
			return
		}
		if IsServiceRunning("postgres.exe") {
			StopPostgreSQL()
			time.Sleep(1 * time.Second)
		}
		os.RemoveAll(pgsqlDataDir)
	}

	config, err := GetConfig()
	if err != nil {
		fmt.Printf("%sFailed to load configuration: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	fmt.Print(shared.ColorYellow, "Enter a password for 'postgres' (or press Enter for a random one): ", shared.ColorReset)
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if password == "" {
		randomPass, err := generateRandomPassword(16)
		if err != nil {
			fmt.Printf("%sFailed to generate random password: %v%s\n", shared.ColorRed, err, shared.ColorReset)
			return
		}
		password = randomPass
		fmt.Printf("%sGenerated Random Password: %s%s%s%s (This is also saved in the config)\n", shared.ColorGreen, shared.ColorYellow, password, shared.ColorGreen, shared.ColorReset)
	}

	config.PostgresPassword = password
	if err := SaveConfig(config); err != nil {
		fmt.Printf("%sFailed to save password to config: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	pwFilePath := filepath.Join(os.TempDir(), "pgpass.tmp")
	if err := os.WriteFile(pwFilePath, []byte(password), 0600); err != nil {
		fmt.Printf("%sFailed to create temporary password file: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}
	defer os.Remove(pwFilePath)

	fmt.Printf("%sInitializing PostgreSQL database cluster...%s\n", shared.ColorYellow, shared.ColorReset)

	cmd := exec.Command(initdbExe,
		"-D", pgsqlDataDir,
		"-U", "postgres",
		"--pwfile", pwFilePath,
		"-E", "UTF8",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%sError initializing PostgreSQL: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		fmt.Println(string(output))
		return
	}

	fmt.Printf("%sPostgreSQL database initialized successfully.%s\n", shared.ColorGreen, shared.ColorReset)
	applyPostgresSecuritySettings(config.DevelopmentMode)
}

func StartPostgreSQL() {
	if !isPostgreSQLInitialized() {
		fmt.Printf("%sPostgreSQL data directory not found or empty.%s\n", shared.ColorRed, shared.ColorReset)
		fmt.Println("Please initialize the database first from the main menu.")
		return
	}

	config, _ := GetConfig()
	applyPostgresSecuritySettings(config.DevelopmentMode)

	fmt.Printf("%sAttempting to start PostgreSQL server...%s\n", shared.ColorYellow, shared.ColorReset)

	cmd := exec.Command(pgctlExe, "start", "-D", pgsqlDataDir, "-l", pgsqlLogFile)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	err := cmd.Start()
	if err != nil {
		fmt.Printf("%sError starting PostgreSQL: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}
	cmd.Process.Release()

	fmt.Printf("%sPostgreSQL server started in background.%s\n", shared.ColorGreen, shared.ColorReset)
}

func StopPostgreSQL() {
	fmt.Printf("%sStopping PostgreSQL server...%s\n", shared.ColorYellow, shared.ColorReset)
	cmd := exec.Command(pgctlExe, "stop", "-D", pgsqlDataDir, "-m", "fast")
	err := cmd.Run()
	if err != nil {
		if !strings.Contains(err.Error(), "is not running") {
			// kondisi server ga jalan
		}
	} else {
		fmt.Printf("%sPostgreSQL server stopped.%s\n", shared.ColorYellow, shared.ColorReset)
	}
}

func RestartPostgreSQL() {
	fmt.Printf("%sRestarting PostgreSQL to apply changes...%s\n", shared.ColorYellow, shared.ColorReset)
	StopPostgreSQL()
	time.Sleep(2 * time.Second)
	StartPostgreSQL()
}

func ViewPostgresPassword() {
	config, err := GetConfig()
	if err != nil {
		fmt.Printf("%sFailed to load configuration: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	if config.PostgresPassword == "" {
		fmt.Println("No PostgreSQL password has been set yet. Please initialize the database first.")
		return
	}

	fmt.Printf("PostgreSQL Superuser (postgres) Password: %s%s%s\n", shared.ColorGreen, config.PostgresPassword, shared.ColorReset)
}
