package service

import (
	"fmt"
	"gecko/internal/shared"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	openSSLExe       = `C:\Gecko\bin\openssl\openssl.exe`
	apacheSSLConfFile = `C:\Gecko\etc\config\httpd\httpd-ssl.conf`
	sslBaseDir       = `C:\Gecko\etc\ssl`
	caKeyPath        = `C:\Gecko\etc\ssl\GeckoRootCA.key`
	caCertPath       = `C:\Gecko\etc\ssl\GeckoRootCA.pem`
	caSubject        = "/C=ID/ST=DKI Jakarta/L=Jakarta Utara/O=Gecko/CN=Gecko Local Development CA"
	vhostCertsDir    = `C:\Gecko\etc\ssl\certs`
	vhostKeysDir     = `C:\Gecko\etc\ssl\keys`
	defaultCertPath  = `C:\Gecko\etc\ssl\gecko.crt`
	defaultKeyPath   = `C:\Gecko\etc\ssl\gecko.key`
)

func activateSSLListener() error {
	input, err := os.ReadFile(apacheSSLConfFile)
	if err != nil {
		return err
	}
	content := string(input)

	if strings.Contains(content, "Listen 443") {
		return nil
	}

	fmt.Printf("%sEnabling 'Listen 443' in httpd-ssl.conf...%s\n", shared.ColorYellow, shared.ColorReset)
	newContent := "Listen 443\n" + content
	return os.WriteFile(apacheSSLConfFile, []byte(newContent), 0644)
}

func GenerateDefaultCertificate() {
	fmt.Printf("%sGenerating default SSL certificate for Gecko (localhost)...%s\n", shared.ColorYellow, shared.ColorReset)
	err := generateCert("localhost", defaultCertPath, defaultKeyPath)
	if err != nil {
		fmt.Printf("%sError generating default certificate: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}
	fmt.Printf("%sDefault certificate 'gecko.crt' created successfully.%s\n", shared.ColorGreen, shared.ColorReset)

	if err := EnableDefaultVHostSSL(); err != nil {
		fmt.Printf("%sError enabling default SSL vhost config: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	if err := activateSSLListener(); err != nil {
		fmt.Printf("%sError activating Apache SSL listener: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	RestartApache()
}

func runCmd(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running command '%s': %v\nOutput: %s", command, err, string(output))
	}
	return nil
}

func generateRootCA() error {
	fmt.Printf("%sGenerating new Gecko Root CA...%s\n", shared.ColorYellow, shared.ColorReset)
	if err := os.MkdirAll(sslBaseDir, os.ModePerm); err != nil {
		return err
	}
	err := runCmd(openSSLExe, "genrsa", "-out", caKeyPath, "4096")
	if err != nil {
		return err
	}
	err = runCmd(openSSLExe, "req", "-x509", "-new", "-nodes", "-key", caKeyPath, "-sha256", "-days", "3650", "-out", caCertPath, "-subj", caSubject)
	if err != nil {
		return err
	}
	fmt.Printf("%sGecko Root CA created at %s%s\n", shared.ColorGreen, caCertPath, shared.ColorReset)
	return nil
}

func installRootCAToWindows() error {
	fmt.Printf("%sAttempting to install Gecko Root CA to Windows Trust Store...%s\n", shared.ColorYellow, shared.ColorReset)
	fmt.Println("A security prompt will appear. Please accept it to trust the new CA.")
	err := runCmd("certutil", "-addstore", "-f", "ROOT", caCertPath)
	if err != nil {
		return err
	}
	fmt.Printf("%sCA successfully installed!%s\n", shared.ColorGreen, shared.ColorReset)
	return nil
}

func InstallGeckoRootCA() {
	if _, err := os.Stat(openSSLExe); os.IsNotExist(err) {
		fmt.Printf("%sError: openssl.exe not found at %s%s\n", shared.ColorRed, openSSLExe, shared.ColorReset)
		fmt.Println("Please download OpenSSL for Windows and place its bin content there.")
		return
	}
	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		if err := generateRootCA(); err != nil {
			fmt.Printf("%sFailed to generate Root CA: %v%s\n", shared.ColorRed, err, shared.ColorReset)
			return
		}
	} else {
		fmt.Printf("%sGecko Root CA already exists. Skipping generation.%s\n", shared.ColorYellow, shared.ColorReset)
	}
	if err := installRootCAToWindows(); err != nil {
		fmt.Printf("%sFailed to install Root CA: %v%s\n", shared.ColorRed, err, shared.ColorReset)
	}
}

func GenerateVHostCert(domainName string) error {
	fmt.Printf("%sGenerating SSL certificate for %s...%s\n", shared.ColorYellow, domainName, shared.ColorReset)
	certPath := filepath.Join(vhostCertsDir, domainName+".crt")
	keyPath := filepath.Join(vhostKeysDir, domainName+".key")
	return generateCert(domainName, certPath, keyPath)
}

func generateCert(domainName, certOutPath, keyOutPath string) error {
	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		return fmt.Errorf("Gecko Root CA not found. Please run 'Install Gecko Root CA' from the menu first")
	}
	_ = os.MkdirAll(filepath.Dir(certOutPath), os.ModePerm)
	_ = os.MkdirAll(filepath.Dir(keyOutPath), os.ModePerm)
	tmpKeyPath := filepath.Join(sslBaseDir, "tmp.key")
	tmpCsrPath := filepath.Join(sslBaseDir, "tmp.csr")
	subject := fmt.Sprintf("/C=ID/ST=DKI Jakarta/L=Jakarta Utara/O=Gecko/CN=%s", domainName)
	if err := runCmd(openSSLExe, "genrsa", "-out", tmpKeyPath, "2048"); err != nil {
		return err
	}
	if err := runCmd(openSSLExe, "req", "-new", "-key", tmpKeyPath, "-out", tmpCsrPath, "-subj", subject); err != nil {
		return err
	}
	extFileContent := fmt.Sprintf("authorityKeyIdentifier=keyid,issuer\nbasicConstraints=CA:FALSE\nkeyUsage=digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment\nsubjectAltName=@alt_names\n\n[alt_names]\nDNS.1 = %s", domainName)
	if domainName == "localhost" {
		extFileContent += "\nDNS.2 = 127.0.0.1"
	}
	extFilePath := filepath.Join(sslBaseDir, "tmp.ext")
	if err := os.WriteFile(extFilePath, []byte(extFileContent), 0644); err != nil {
		return err
	}
	err := runCmd(openSSLExe, "x509", "-req", "-in", tmpCsrPath, "-CA", caCertPath, "-CAkey", caKeyPath, "-CAcreateserial", "-out", certOutPath, "-days", "825", "-sha256", "-extfile", extFilePath)
	if err != nil {
		return err
	}
	if err := os.Rename(tmpKeyPath, keyOutPath); err != nil {
		return err
	}
	os.Remove(tmpCsrPath)
	os.Remove(extFilePath)
	os.Remove(filepath.Join(sslBaseDir, "GeckoRootCA.srl"))
	return nil
}