package service

import (
	"bufio"
	"fmt"
	"gecko/internal/shared"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	phpBaseDir       = `C:\Gecko\bin\php`
	phpActiveSymlink = `C:\Gecko\bin\php\php`
)

func listInstalledPHPVersions() ([]string, error) {
	files, err := os.ReadDir(phpBaseDir)
	if err != nil {
		return nil, err
	}
	var versions []string
	for _, file := range files {
		if file.IsDir() && strings.HasPrefix(file.Name(), "php-") {
			versions = append(versions, file.Name())
		}
	}
	return versions, nil
}

func SwitchPHPVersion(reader *bufio.Reader) {
	versions, err := listInstalledPHPVersions()
	if err != nil || len(versions) == 0 {
		fmt.Printf("%sNo PHP versions found in '%s'.%s\n", shared.ColorRed, phpBaseDir, shared.ColorReset)
		fmt.Println("Please place your PHP version folders (e.g., 'php-84') inside it.")
		return
	}

	fmt.Println(shared.ColorGreen, "Please select a PHP version to activate:", shared.ColorReset)
	for i, v := range versions {
		fmt.Printf("%d. %s\n", i+1, v)
	}
	fmt.Println("0. Cancel")

	fmt.Print(shared.ColorYellow, "\nEnter your choice: ", shared.ColorReset)
	choiceStr, _ := reader.ReadString('\n')
	var choice int
	fmt.Sscanf(strings.TrimSpace(choiceStr), "%d", &choice)

	if choice <= 0 || choice > len(versions) {
		fmt.Println("Operation cancelled.")
		return
	}

	selectedVersionDirName := versions[choice-1]
	targetDir := filepath.Join(phpBaseDir, selectedVersionDirName)
	fmt.Printf("%sSwitching active PHP version to %s...%s\n", shared.ColorYellow, selectedVersionDirName, shared.ColorReset)

	// remove symlink if exists
	if err := os.RemoveAll(phpActiveSymlink); err != nil {
		fmt.Printf("%sFailed to remove old symlink/directory: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	// create symlink
	cmd := exec.Command("cmd", "/c", "mklink", "/D", phpActiveSymlink, targetDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%sError creating symbolic link: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		fmt.Println(string(output))
		fmt.Println("Symbolic links are crucial for this feature. Ensure you're running as Administrator.")
		return
	}

	fmt.Printf("%sSuccessfully switched to %s.%s\n", shared.ColorGreen, selectedVersionDirName, shared.ColorReset)
	RestartApache()
}
