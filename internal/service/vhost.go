package service

import (
	"bufio"
	"fmt"
	"gecko/internal/shared"
	"os"
	"path/filepath"
	"strings"
)

const (
	wwwDir           = `C:\Gecko\www`
	sitesEnabledDir  = `C:\Gecko\etc\config\httpd\sites-enabled`
	defaultVHostFile = `C:\Gecko\etc\config\httpd\sites-enabled\00-default.conf`
	hostsFilePath    = `C:\Windows\System32\drivers\etc\hosts`
	geckoStartBlock  = "#GeckoStart"
	geckoEndBlock    = "#GeckoEnd"
	prohibitedVHosts = "00-default.conf"
)

func createVHostFile(docRoot, domainName string, useSSL bool) error {
	docRootApache := filepath.ToSlash(docRoot)
	var configContent string

	if useSSL {
		certPath := filepath.ToSlash(filepath.Join(vhostCertsDir, domainName+".crt"))
		keyPath := filepath.ToSlash(filepath.Join(vhostKeysDir, domainName+".key"))
		vhostTemplate := `<VirtualHost *:80>
    ServerName %[1]s
    DocumentRoot "%[2]s"
    <Directory "%[2]s">
        AllowOverride All
        Require all granted
    </Directory>
</VirtualHost>

<VirtualHost *:443>
    ServerName %[1]s
    DocumentRoot "%[2]s"
    <Directory "%[2]s">
        AllowOverride All
        Require all granted
    </Directory>
    
    SSLEngine on
    SSLCertificateFile      "%[3]s"
    SSLCertificateKeyFile   "%[4]s"
</VirtualHost>`
		configContent = fmt.Sprintf(vhostTemplate, domainName, docRootApache, certPath, keyPath)
	} else {
		vhostTemplate := `<VirtualHost *:80>
    ServerName %[1]s
    DocumentRoot "%[2]s"
    <Directory "%[2]s">
        AllowOverride All
        Require all granted
    </Directory>
</VirtualHost>`
		configContent = fmt.Sprintf(vhostTemplate, domainName, docRootApache)
	}

	configPath := filepath.Join(sitesEnabledDir, domainName+".conf")
	return os.WriteFile(configPath, []byte(configContent), 0644)
}

func isSSLEnabled() bool {
	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func EnableDefaultVHostSSL() error {
	fmt.Printf("%sEnabling SSL configuration for default host...%s\n", shared.ColorYellow, shared.ColorReset)
	content, err := os.ReadFile(defaultVHostFile)
	if err != nil {
		return err
	}
	if strings.Contains(string(content), "<VirtualHost _default_:443>") {
		fmt.Printf("%sDefault host SSL config already enabled. Skipping.%s\n", shared.ColorYellow, shared.ColorReset)
		return nil
	}
	sslBlock := `
<VirtualHost *:443>
    ServerName localhost
    DocumentRoot "C:/Gecko/www"
    <Directory "C:/Gecko/www/">
        AllowOverride All
        Require all granted
    </Directory>
    SSLEngine on
    SSLCertificateFile      "C:/Gecko/etc/ssl/gecko.crt"
    SSLCertificateKeyFile   "C:/Gecko/etc/ssl/gecko.key"
</VirtualHost>`
	file, err := os.OpenFile(defaultVHostFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(sslBlock); err != nil {
		return err
	}
	fmt.Printf("%sDefault host SSL config enabled successfully.%s\n", shared.ColorGreen, shared.ColorReset)
	return nil
}

func CreateVirtualHost(domainName, choice string) {
	domainName = strings.ToLower(strings.TrimSpace(domainName))
	if domainName == "" {
		fmt.Printf("%sDomain name cannot be empty.%s\n", shared.ColorRed, shared.ColorReset)
		return
	}
	docRoot := filepath.Join(wwwDir, domainName)
	vhostConfigFile := filepath.Join(sitesEnabledDir, domainName+".conf")
	if _, err := os.Stat(vhostConfigFile); err == nil && choice != "y" {
		return
	}
	fmt.Printf("%sProcessing Virtual Host for %s...%s\n", shared.ColorYellow, domainName, shared.ColorReset)
	if choice == "y" {
		if err := formatVHostDirectory(docRoot, domainName); err != nil {
			fmt.Printf("%sError formatting directory: %v%s\n", shared.ColorRed, err, shared.ColorReset)
			return
		}
	} else {
		if err := createDocRoot(docRoot, domainName); err != nil {
			fmt.Printf("%sError creating document root: %v%s\n", shared.ColorRed, err, shared.ColorReset)
			return
		}
	}
	sslEnabled := isSSLEnabled()
	if sslEnabled {
		if err := activateSSLListener(); err != nil {
			fmt.Printf("%sError activating Apache SSL listener: %v%s\n", shared.ColorRed, err, shared.ColorReset)
			return
		}
		if err := GenerateVHostCert(domainName); err != nil {
			fmt.Printf("%sError generating SSL certificate: %v%s\n", shared.ColorRed, err, shared.ColorReset)
			return
		}
	} else {
		fmt.Printf("%sSSL is not enabled. Creating HTTP-only virtual host.%s\n", shared.ColorYellow, shared.ColorReset)
	}
	if err := createVHostFile(docRoot, domainName, sslEnabled); err != nil {
		fmt.Printf("%sError creating vhost config file: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}
	if err := updateHostsFile(domainName, true); err != nil {
		fmt.Printf("%sError updating hosts file: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}
	RestartApache()
	if sslEnabled {
		fmt.Printf("%sSuccessfully processed virtual host. You can access it at https://%s%s\n", shared.ColorGreen, domainName, shared.ColorReset)
	} else {
		fmt.Printf("%sSuccessfully processed virtual host. You can access it at http://%s%s\n", shared.ColorGreen, domainName, shared.ColorReset)
	}
}

func DeleteVirtualHost(domainName string) {
	domainName = strings.ToLower(strings.TrimSpace(domainName))
	fmt.Printf("%sDeleting virtual host %s...%s\n", shared.ColorYellow, domainName, shared.ColorReset)
	_ = os.Remove(filepath.Join(sitesEnabledDir, domainName+".conf"))
	_ = os.RemoveAll(filepath.Join(wwwDir, domainName))
	_ = os.Remove(filepath.Join(vhostCertsDir, domainName+".crt"))
	_ = os.Remove(filepath.Join(vhostKeysDir, domainName+".key"))
	if err := updateHostsFile(domainName, false); err != nil {
		fmt.Printf("%sCould not update hosts file: %v%s\n", shared.ColorRed, err, shared.ColorReset)
	}
	RestartApache()
	fmt.Printf("%sVirtual host %s deleted successfully.%s\n", shared.ColorGreen, domainName, shared.ColorReset)
}

func ListVirtualHosts() ([]string, error) {
	files, err := os.ReadDir(sitesEnabledDir)
	if err != nil {
		return nil, err
	}
	var vhosts []string
	for _, file := range files {
		if file.Name() == prohibitedVHosts {
			continue
		}
		if strings.HasSuffix(file.Name(), ".conf") {
			domainName := strings.TrimSuffix(file.Name(), ".conf")
			vhosts = append(vhosts, domainName)
		}
	}
	return vhosts, nil
}

func formatVHostDirectory(path, domainName string) error {
	fmt.Printf("%sFormatting directory %s...%s\n", shared.ColorYellow, path, shared.ColorReset)
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return createDocRoot(path, domainName)
}

func createDocRoot(path, domainName string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	indexPath := filepath.Join(path, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to %[1]s</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; display: flex; justify-content: center; align-items: center; height: 90vh; text-align: center; background-color: #1a1a1a; color: #f0f0f0; margin: 0; }
        div { border: 1px solid #444; padding: 2rem 4rem; border-radius: 8px; background-color: #2c2c2c; }
        h1 { color: #4ade80; }
    </style>
</head>
<body>
    <div>
        <h1>Welcome to %[1]s!</h1>
        <p>Powered by Gecko 🦎</p>
    </div>
</body>
</html>`
		htmlContent := fmt.Sprintf(htmlTemplate, domainName)
		return os.WriteFile(indexPath, []byte(htmlContent), 0644)
	}
	return nil
}

func updateHostsFile(domainName string, add bool) error {
	file, err := os.Open(hostsFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	var lines []string
	var geckoLines []string
	inGeckoBlock := false
	scanner := bufio.NewReader(file)
	for {
		line, err := scanner.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if strings.HasPrefix(line, geckoStartBlock) {
			inGeckoBlock = true
			continue
		}
		if strings.HasPrefix(line, geckoEndBlock) {
			inGeckoBlock = false
			continue
		}
		if inGeckoBlock {
			geckoLines = append(geckoLines, line)
		} else {
			lines = append(lines, line)
		}
		if err != nil {
			break
		}
	}
	newGeckoLines := []string{}
	entry := "127.0.0.1 " + domainName
	found := false
	for _, line := range geckoLines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.Contains(line, domainName) {
			found = true
			if add {
				newGeckoLines = append(newGeckoLines, entry)
			}
		} else {
			newGeckoLines = append(newGeckoLines, line)
		}
	}
	if add && !found {
		newGeckoLines = append(newGeckoLines, entry)
	}
	finalContent := strings.Join(lines, "\n")
	if len(newGeckoLines) > 0 {
		finalContent += "\n\n" + geckoStartBlock + "\n"
		finalContent += strings.Join(newGeckoLines, "\n") + "\n"
		finalContent += geckoEndBlock
	}
	return os.WriteFile(hostsFilePath, []byte(finalContent), 0644)
}