package service

import (
	"archive/zip"
	"fmt"
	"gecko/internal/shared"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type ProjectType struct {
	Name           string
	NeedsDatabase  bool
	RequiredTools  []string
	DatabaseSuffix string
}

var projectConfigs = map[string]ProjectType{
	"laravel":     {Name: "Laravel", NeedsDatabase: true, RequiredTools: []string{"composer", "php"}, DatabaseSuffix: "_db"},
	"wordpress":   {Name: "WordPress", NeedsDatabase: true, RequiredTools: []string{}, DatabaseSuffix: "_db"},
	"symfony":     {Name: "Symfony", NeedsDatabase: true, RequiredTools: []string{"composer", "php"}, DatabaseSuffix: "_db"},
	"codeigniter": {Name: "CodeIgniter", NeedsDatabase: true, RequiredTools: []string{"composer", "php"}, DatabaseSuffix: "_db"},
	"cakephp":     {Name: "CakePHP", NeedsDatabase: true, RequiredTools: []string{"composer", "php"}, DatabaseSuffix: "_db"},
	"laminas":     {Name: "Laminas", NeedsDatabase: true, RequiredTools: []string{"composer", "php"}, DatabaseSuffix: "_db"},
	"react":       {Name: "React", NeedsDatabase: false, RequiredTools: []string{"npm"}, DatabaseSuffix: ""},
	"vue":         {Name: "Vue.js", NeedsDatabase: false, RequiredTools: []string{"npm"}, DatabaseSuffix: ""},
	"angular":     {Name: "Angular", NeedsDatabase: false, RequiredTools: []string{"npm"}, DatabaseSuffix: ""},
	"svelte":      {Name: "Svelte", NeedsDatabase: false, RequiredTools: []string{"npm"}, DatabaseSuffix: ""},
	"nextjs":      {Name: "Next.js", NeedsDatabase: false, RequiredTools: []string{"npm"}, DatabaseSuffix: ""},
	"nuxtjs":      {Name: "Nuxt.js", NeedsDatabase: false, RequiredTools: []string{"npm"}, DatabaseSuffix: ""},
	"astro":       {Name: "Astro", NeedsDatabase: false, RequiredTools: []string{"npm"}, DatabaseSuffix: ""},
	"vite":        {Name: "Vite", NeedsDatabase: false, RequiredTools: []string{"npm"}, DatabaseSuffix: ""},
	"static":      {Name: "Static HTML", NeedsDatabase: false, RequiredTools: []string{}, DatabaseSuffix: ""},
}

func validateProjectRequirements(projectType, projectName string) bool {
	config, exists := projectConfigs[projectType]
	if !exists {
		fmt.Printf("%sWarning: Unknown project type '%s', proceeding without validation%s\n", shared.ColorYellow, projectType, shared.ColorReset)
		return true
	}

	for _, tool := range config.RequiredTools {
		if !isToolInstalled(tool) {
			fmt.Printf("%sError: %s is required to create %s projects.%s\n", shared.ColorRed, tool, config.Name, shared.ColorReset)
			
			switch tool {
			case "composer":
				fmt.Printf("%sPlease install Composer from https://getcomposer.org%s\n", shared.ColorYellow, shared.ColorReset)
			case "php":
				fmt.Printf("%sPlease ensure PHP is installed and available in PATH%s\n", shared.ColorYellow, shared.ColorReset)
			case "npm":
				fmt.Printf("%sPlease install Node.js from https://nodejs.org%s\n", shared.ColorYellow, shared.ColorReset)
			}
			return false
		}
	}

	if config.NeedsDatabase && !isDatabaseRunning() {
		fmt.Printf("%sDatabase service is required for %s but is not running.%s\n", shared.ColorRed, config.Name, shared.ColorReset)
		fmt.Printf("%sPlease start MySQL or PostgreSQL using the Gecko menu first.%s\n", shared.ColorYellow, shared.ColorReset)
		return false
	}

	if config.NeedsDatabase && config.DatabaseSuffix != "" {
		createDatabase(projectName + config.DatabaseSuffix)
	}

	return true
}

func isToolInstalled(tool string) bool {
	switch tool {
	case "composer":
		return isComposerInstalled()
	case "php":
		return isPHPInstalled()
	case "npm":
		return isNpmInstalled()
	default:
		return true
	}
}

func CreateLaravelProject(projectPath, projectName string) {
	fmt.Printf("%sCreating Laravel project using Composer...%s\n", shared.ColorYellow, shared.ColorReset)
	
	if !validateProjectRequirements("laravel", projectName) {
		return
	}

	os.RemoveAll(projectPath)

	parentDir := filepath.Dir(projectPath)
	projectDirName := filepath.Base(projectPath)

	fmt.Printf("%sRunning: composer create-project laravel/laravel %s%s\n", shared.ColorYellow, projectDirName, shared.ColorReset)
	
	cmd := exec.Command("composer", "create-project", "laravel/laravel", projectDirName, "--prefer-dist", "--no-interaction")
	cmd.Dir = parentDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%sError creating Laravel project: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	configureLaravel(projectPath, projectName)
	
	fmt.Printf("%s✓ Laravel project created successfully%s\n", shared.ColorGreen, shared.ColorReset)
}

func CreateWordPressProject(projectPath, projectName string) {
	fmt.Printf("%sCreating WordPress project using wordpress.org...%s\n", shared.ColorYellow, shared.ColorReset)
	
	if !validateProjectRequirements("wordpress", projectName) {
		return
	}
	
	fmt.Printf("%sDownloading WordPress from wordpress.org...%s\n", shared.ColorYellow, shared.ColorReset)
	
	wordpressURL := "https://wordpress.org/latest.zip"
	zipPath := filepath.Join(projectPath, "wordpress.zip")
	
	os.MkdirAll(projectPath, os.ModePerm)
	
	err := downloadFile(wordpressURL, zipPath)
	if err != nil {
		fmt.Printf("%sError downloading WordPress: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		fmt.Printf("%sWordPress project creation failed. Please check your internet connection.%s\n", shared.ColorRed, shared.ColorReset)
		fmt.Printf("%sYou can manually download WordPress from https://wordpress.org/download/%s\n", shared.ColorYellow, shared.ColorReset)
		return
	}

	fmt.Printf("%sExtracting WordPress...%s\n", shared.ColorYellow, shared.ColorReset)
	err = extractZip(zipPath, projectPath)
	if err != nil {
		fmt.Printf("%sError extracting WordPress: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		fmt.Printf("%sWordPress project creation failed. Please try again.%s\n", shared.ColorRed, shared.ColorReset)
		return
	}

	wordpressDir := filepath.Join(projectPath, "wordpress")
	if _, err := os.Stat(wordpressDir); err == nil {
		moveWordPressFiles(wordpressDir, projectPath)
		os.RemoveAll(wordpressDir)
	}

	os.Remove(zipPath)

	configureWordPress(projectPath, projectName)
	
	fmt.Printf("%s✓ WordPress project created successfully%s\n", shared.ColorGreen, shared.ColorReset)
}

func CreateSymfonyProject(projectPath, projectName string) {
	dirs := []string{
		"src/Controller",
		"src/Entity",
		"src/Repository",
		"config",
		"public",
		"templates",
		"var/cache",
		"var/log",
		"vendor",
	}

	for _, dir := range dirs {
		os.MkdirAll(filepath.Join(projectPath, dir), os.ModePerm)
	}

	indexContent := fmt.Sprintf(`<?php

?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Symfony Project</title>
    <style>
        body { font-family: 'Helvetica Neue', Arial, sans-serif; margin: 0; padding: 40px; background: #f8f9fa; }
        .container { max-width: 600px; margin: 0 auto; text-align: center; }
        .logo { font-size: 72px; margin-bottom: 20px; }
        .title { font-size: 36px; color: #212529; margin-bottom: 20px; }
        .subtitle { font-size: 18px; color: #6c757d; margin-bottom: 30px; }
        .badge { background: #000; color: white; padding: 4px 8px; border-radius: 4px; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">🦎</div>
        <div class="title">%s</div>
        <div class="subtitle">Symfony Project <span class="badge">Created by Gecko</span></div>
        <p>Run <code>composer install</code> to complete the Symfony setup.</p>
    </div>
</body>
</html>`, projectName, projectName)

	os.WriteFile(filepath.Join(projectPath, "public", "index.php"), []byte(indexContent), 0644)

	fmt.Printf("%s✓ Basic Symfony structure created%s\n", shared.ColorGreen, shared.ColorReset)
}

func CreateReactProject(projectPath, projectName string) {
	fmt.Printf("%sCreating React project using npx create-react-app...%s\n", shared.ColorYellow, shared.ColorReset)
	
	if !validateProjectRequirements("react", projectName) {
		return
	}

	os.RemoveAll(projectPath)

	parentDir := filepath.Dir(projectPath)
	projectDirName := filepath.Base(projectPath)

	fmt.Printf("%sRunning: npx create-react-app %s%s\n", shared.ColorYellow, projectDirName, shared.ColorReset)
	
	cmd := exec.Command("npx", "create-react-app", projectDirName)
	cmd.Dir = parentDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%sError creating React project: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		fmt.Printf("%sReact project creation failed. Please check your internet connection and npm installation.%s\n", shared.ColorRed, shared.ColorReset)
		fmt.Printf("%sYou can try running 'npx create-react-app %s' manually%s\n", shared.ColorYellow, projectDirName, shared.ColorReset)
		return
	}
	
	fmt.Printf("%s✓ React project created successfully%s\n", shared.ColorGreen, shared.ColorReset)
	fmt.Printf("%sNext steps:%s\n", shared.ColorYellow, shared.ColorReset)
	fmt.Printf("%s  cd %s%s\n", shared.ColorGreen, projectDirName, shared.ColorReset)
	fmt.Printf("%s  npm start         # Development server (localhost:3000)%s\n", shared.ColorGreen, shared.ColorReset)
	fmt.Printf("%s  npm run build     # Production build%s\n", shared.ColorGreen, shared.ColorReset)
}

func CreateVueProject(projectPath, projectName string) {
	fmt.Printf("%sCreating Vue.js project using npx create-vue...%s\n", shared.ColorYellow, shared.ColorReset)
	
	if !validateProjectRequirements("vue", projectName) {
		return
	}

	os.RemoveAll(projectPath)

	parentDir := filepath.Dir(projectPath)
	projectDirName := filepath.Base(projectPath)

	fmt.Printf("%sRunning: npx create-vue@latest %s%s\n", shared.ColorYellow, projectDirName, shared.ColorReset)
	
	cmd := exec.Command("npx", "create-vue@latest", projectDirName, "--default")
	cmd.Dir = parentDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%sError creating Vue project: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		fmt.Printf("%sVue.js project creation failed. Please check your internet connection and npm installation.%s\n", shared.ColorRed, shared.ColorReset)
		fmt.Printf("%sYou can try running 'npx create-vue@latest %s' manually%s\n", shared.ColorYellow, projectDirName, shared.ColorReset)
		return
	}
	
	fmt.Printf("%s✓ Vue.js project created successfully%s\n", shared.ColorGreen, shared.ColorReset)
	fmt.Printf("%sNext steps:%s\n", shared.ColorYellow, shared.ColorReset)
	fmt.Printf("%s  cd %s%s\n", shared.ColorGreen, projectDirName, shared.ColorReset)
	fmt.Printf("%s  npm install       # Install dependencies%s\n", shared.ColorGreen, shared.ColorReset)
	fmt.Printf("%s  npm run dev       # Development server (localhost:5173)%s\n", shared.ColorGreen, shared.ColorReset)
	fmt.Printf("%s  npm run build     # Production build%s\n", shared.ColorGreen, shared.ColorReset)
}

func CreateStaticProject(projectPath, projectName string) {
	dirs := []string{
		"css",
		"js",
		"images",
		"assets",
	}

	for _, dir := range dirs {
		os.MkdirAll(filepath.Join(projectPath, dir), os.ModePerm)
	}

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Static Website</title>
    <style>
        :root {
            --bg-color: #1a1a1a;
            --card-color: #2c2c2c;
            --text-color: #f0f0f0;
            --muted-color: #a0a0a0;
            --accent-color: #4ade80;
            --border-color: #444;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji";
            background-color: var(--bg-color);
            color: var(--text-color);
            display: flex;
            justify-content: center;
            align-items: center;
            text-align: center;
            min-height: 100vh;
            padding: 2rem;
            animation: fadeIn 0.5s ease-in-out;
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .content {
            max-width: 600px;
        }
        
        h1 {
            font-size: 4.5rem;
            font-weight: 800;
            color: var(--accent-color);
            display: flex;
            justify-content: center;
            align-items: center;
            gap: 1rem;
            margin-bottom: 1rem;
        }

        .subtitle {
            font-size: 1.25rem;
            color: var(--muted-color);
            margin-bottom: 3rem;
        }
        
        .next-steps {
            font-size: 1rem;
            color: var(--muted-color);
        }

        .next-steps code {
            background-color: #333;
            color: #f0f0f0;
            padding: 0.2rem 0.4rem;
            border-radius: 4px;
            font-family: "Courier New", Courier, monospace;
        }

        footer {
            position: absolute;
            bottom: 1.5rem;
            font-size: 0.9rem;
            color: #666;
        }

        .nodeco {
            text-decoration: none;
            color: #a0a0a0;
        }

        .badge {
            background: var(--accent-color);
            color: var(--bg-color);
            padding: 0.3rem 0.8rem;
            border-radius: 12px;
            font-size: 0.8rem;
            font-weight: 600;
            margin-left: 0.5rem;
        }
    </style>
</head>
<body>
    <main class="content">
        <h1><span>%s</span> 🌐</h1>
        <p class="subtitle">Your static website is ready to customize! <span class="badge">Created by Gecko</span></p>
        
        <div class="next-steps">
            <p>Start building your amazing website by editing this file and adding your content.</p>
            <p>Your assets are organized in <code>css/</code>, <code>js/</code>, and <code>images/</code> folders.</p>
        </div>
    </main>
    
    <footer>
        <a href="https://github.com/shiwildy/Gecko" class="nodeco">Powered by Gecko</a>
    </footer>
</body>
</html>`, projectName, projectName)

	os.WriteFile(filepath.Join(projectPath, "index.html"), []byte(htmlContent), 0644)

	cssContent := `/* Custom styles for your website */
.custom-section {
    margin: 2rem 0;
    padding: 1.5rem;
    background: var(--card-color);
    border-radius: 8px;
    border: 1px solid var(--border-color);
}

.button {
    display: inline-block;
    background: var(--accent-color);
    color: var(--bg-color);
    padding: 0.75rem 1.5rem;
    text-decoration: none;
    border-radius: 6px;
    font-weight: 600;
    transition: all 0.2s ease;
}

.button:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(74, 222, 128, 0.3);
}

@media (max-width: 768px) {
    .content {
        padding: 1rem;
    }
    
    h1 {
        font-size: 3rem;
    }
}`

	os.WriteFile(filepath.Join(projectPath, "css", "style.css"), []byte(cssContent), 0644)

	jsContent := `// JavaScript for your static website
console.log('Welcome to your new static website!');

document.addEventListener('DOMContentLoaded', function() {
    console.log('Website loaded successfully!');
    
    const heading = document.querySelector('h1');
    if (heading) {
        heading.addEventListener('click', function() {
            this.style.transform = 'scale(1.05)';
            this.style.transition = 'transform 0.3s ease';
            
            setTimeout(() => {
                this.style.transform = 'scale(1)';
            }, 300);
        });
    }
});`

	os.WriteFile(filepath.Join(projectPath, "js", "script.js"), []byte(jsContent), 0644)

	fmt.Printf("%s✓ Basic static HTML structure created%s\n", shared.ColorGreen, shared.ColorReset)
}

func CreateCodeIgniterProject(projectPath, projectName string) {
	fmt.Printf("%sCreating CodeIgniter project using Composer...%s\n", shared.ColorYellow, shared.ColorReset)
	
	if !validateProjectRequirements("codeigniter", projectName) {
		return
	}

	cmd := exec.Command("composer", "create-project", "codeigniter4/appstarter", projectPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("%sError creating CodeIgniter project: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	fmt.Printf("%s✓ CodeIgniter project created successfully!%s\n", shared.ColorGreen, shared.ColorReset)
}

func CreateCakePHPProject(projectPath, projectName string) {
	fmt.Printf("%sCreating CakePHP project using Composer...%s\n", shared.ColorYellow, shared.ColorReset)
	
	if !validateProjectRequirements("cakephp", projectName) {
		return
	}

	cmd := exec.Command("composer", "create-project", "--prefer-dist", "cakephp/app:~4.0", projectPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("%sError creating CakePHP project: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	fmt.Printf("%s✓ CakePHP project created successfully!%s\n", shared.ColorGreen, shared.ColorReset)
}

func CreateLaminasProject(projectPath, projectName string) {
	fmt.Printf("%sCreating Laminas project using Composer...%s\n", shared.ColorYellow, shared.ColorReset)
	
	if !validateProjectRequirements("laminas", projectName) {
		return
	}

	cmd := exec.Command("composer", "create-project", "laminas/laminas-mvc-skeleton", projectPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("%sError creating Laminas project: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	fmt.Printf("%s✓ Laminas project created successfully!%s\n", shared.ColorGreen, shared.ColorReset)
}

func CreateNextJSProject(projectPath, projectName string) {
	fmt.Printf("%sCreating Next.js project using npx...%s\n", shared.ColorYellow, shared.ColorReset)
	
	if !validateProjectRequirements("nextjs", projectName) {
		return
	}

	parentDir := filepath.Dir(projectPath)
	projectDirName := filepath.Base(projectPath)

	cmd := exec.Command("npx", "create-next-app@latest", projectDirName, "--yes")
	cmd.Dir = parentDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("%sError creating Next.js project: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		fmt.Printf("%sNext.js project creation failed. Please check your internet connection and npm installation.%s\n", shared.ColorRed, shared.ColorReset)
		fmt.Printf("%sYou can try running 'npx create-next-app@latest %s' manually%s\n", shared.ColorYellow, projectDirName, shared.ColorReset)
		return
	}

	fmt.Printf("%s✓ Next.js project created successfully!%s\n", shared.ColorGreen, shared.ColorReset)
}

func CreateNuxtJSProject(projectPath, projectName string) {
	fmt.Printf("%sCreating Nuxt.js project using npx...%s\n", shared.ColorYellow, shared.ColorReset)
	
	if !validateProjectRequirements("nuxtjs", projectName) {
		return
	}

	parentDir := filepath.Dir(projectPath)
	projectDirName := filepath.Base(projectPath)

	cmd := exec.Command("npx", "nuxi@latest", "init", projectDirName)
	cmd.Dir = parentDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("%sError creating Nuxt.js project: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	fmt.Printf("%s✓ Nuxt.js project created successfully!%s\n", shared.ColorGreen, shared.ColorReset)
}

func CreateAngularProject(projectPath, projectName string) {
	fmt.Printf("%sCreating Angular project using npx...%s\n", shared.ColorYellow, shared.ColorReset)
	
	if !validateProjectRequirements("angular", projectName) {
		return
	}

	parentDir := filepath.Dir(projectPath)
	projectDirName := filepath.Base(projectPath)

	cmd := exec.Command("npx", "@angular/cli@latest", "new", projectDirName, "--routing", "--style=css", "--skip-git")
	cmd.Dir = parentDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("%sError creating Angular project: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	fmt.Printf("%s✓ Angular project created successfully!%s\n", shared.ColorGreen, shared.ColorReset)
}

func CreateSvelteProject(projectPath, projectName string) {
	fmt.Printf("%sCreating Svelte project using npx...%s\n", shared.ColorYellow, shared.ColorReset)
	
	if !validateProjectRequirements("svelte", projectName) {
		return
	}

	parentDir := filepath.Dir(projectPath)
	projectDirName := filepath.Base(projectPath)

	cmd := exec.Command("npx", "sv", "create", projectDirName)
	cmd.Dir = parentDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("%sError creating Svelte project: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	fmt.Printf("%s✓ Svelte project created successfully!%s\n", shared.ColorGreen, shared.ColorReset)
}

func CreateAstroProject(projectPath, projectName string) {
	fmt.Printf("%sCreating Astro project using npx...%s\n", shared.ColorYellow, shared.ColorReset)
	
	if !validateProjectRequirements("astro", projectName) {
		return
	}

	parentDir := filepath.Dir(projectPath)
	projectDirName := filepath.Base(projectPath)

	cmd := exec.Command("npx", "create-astro@latest", projectDirName, "--template", "minimal", "--yes")
	cmd.Dir = parentDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("%sError creating Astro project: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	fmt.Printf("%s✓ Astro project created successfully!%s\n", shared.ColorGreen, shared.ColorReset)
}

func CreateViteProject(projectPath, projectName string) {
	fmt.Printf("%sCreating Vite project using npx...%s\n", shared.ColorYellow, shared.ColorReset)
	
	if !validateProjectRequirements("vite", projectName) {
		return
	}

	parentDir := filepath.Dir(projectPath)
	projectDirName := filepath.Base(projectPath)

	cmd := exec.Command("npx", "create-vite@latest", projectDirName, "--template", "vanilla-ts")
	cmd.Dir = parentDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("%sError creating Vite project: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}

	fmt.Printf("%s✓ Vite project created successfully!%s\n", shared.ColorGreen, shared.ColorReset)
}

func isDatabaseRunning() bool {
	return isMySQLRunning() || isPostgreSQLRunning()
}

func isMySQLRunning() bool {
	cmd := exec.Command("mysql", "-u", "root", "--connect-timeout=1", "-e", "SELECT 1;")
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	return err == nil
}

func isPostgreSQLRunning() bool {
	return IsServiceRunning("postgres.exe")
}

func createDatabase(dbName string) {
	if isMySQLRunning() {
		createMySQLDatabase(dbName)
	}
	if isPostgreSQLRunning() {
		createPostgreSQLDatabase(dbName)
	}
}

func createMySQLDatabase(dbName string) {
	fmt.Printf("%sCreating MySQL database '%s'...%s\n", shared.ColorYellow, dbName, shared.ColorReset)
	
	createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", dbName)
	
	cmd := exec.Command("mysql", "-u", "root", "-e", createDBSQL)
	if err := cmd.Run(); err != nil {
		fmt.Printf("%sWarning: Failed to create MySQL database '%s': %v%s\n", shared.ColorYellow, dbName, err, shared.ColorReset)
		fmt.Printf("%sPlease create the database manually: CREATE DATABASE %s;%s\n", shared.ColorYellow, dbName, shared.ColorReset)
	} else {
		fmt.Printf("%s✓ MySQL database '%s' created successfully%s\n", shared.ColorGreen, dbName, shared.ColorReset)
	}
}

func createPostgreSQLDatabase(dbName string) {
	fmt.Printf("%sCreating PostgreSQL database '%s'...%s\n", shared.ColorYellow, dbName, shared.ColorReset)
	
	cmd := exec.Command("createdb", "-U", "postgres", dbName)
	if err := cmd.Run(); err != nil {
		fmt.Printf("%sWarning: Failed to create PostgreSQL database '%s': %v%s\n", shared.ColorYellow, dbName, err, shared.ColorReset)
		fmt.Printf("%sPlease create the database manually: CREATE DATABASE %s;%s\n", shared.ColorYellow, dbName, shared.ColorReset)
	} else {
		fmt.Printf("%s✓ PostgreSQL database '%s' created successfully%s\n", shared.ColorGreen, dbName, shared.ColorReset)
	}
}

func isPHPInstalled() bool {
	cmd := exec.Command("php", "--version")
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	return err == nil
}

func isNpmInstalled() bool {
	cmd := exec.Command("npm", "--version")
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	return err == nil
}

func isComposerInstalled() bool {
	cmd := exec.Command("composer", "--version")
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	return err == nil
}

func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer r.Close()

	os.MkdirAll(dest, os.ModePerm)

	for _, f := range r.File {
		path := filepath.Join(dest, f.Name)
		if !filepath.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.FileInfo().Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}

		if err := extractFile(f, path); err != nil {
			return fmt.Errorf("failed to extract file %s: %v", f.Name, err)
		}
	}

	return nil
}

func extractFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)
	return err
}

func downloadFile(url, filepath string) error {
	fmt.Printf("%sDownloading from %s...%s\n", shared.ColorYellow, url, shared.ColorReset)
	
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	if resp.ContentLength > 0 {
		sizeMB := float64(resp.ContentLength) / (1024 * 1024)
		fmt.Printf("%sDownloading %.1f MB...%s\n", shared.ColorYellow, sizeMB, shared.ColorReset)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	
	fmt.Printf("%s✓ Download completed%s\n", shared.ColorGreen, shared.ColorReset)
	return nil
}

func moveWordPressFiles(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	os.MkdirAll(filepath.Dir(dst), os.ModePerm)
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func configureWordPress(projectPath, projectName string) {
	configContent := fmt.Sprintf(`<?php

define('DB_NAME', '%s_db');
define('DB_USER', 'root');
define('DB_PASSWORD', '');
define('DB_HOST', 'localhost');
define('DB_CHARSET', 'utf8mb4');
define('DB_COLLATE', '');

define('AUTH_KEY',         'gecko-auth-key-' . uniqid());
define('SECURE_AUTH_KEY',  'gecko-secure-auth-key-' . uniqid());
define('LOGGED_IN_KEY',    'gecko-logged-in-key-' . uniqid());
define('NONCE_KEY',        'gecko-nonce-key-' . uniqid());
define('AUTH_SALT',        'gecko-auth-salt-' . uniqid());
define('SECURE_AUTH_SALT', 'gecko-secure-auth-salt-' . uniqid());
define('LOGGED_IN_SALT',   'gecko-logged-in-salt-' . uniqid());
define('NONCE_SALT',       'gecko-nonce-salt-' . uniqid());

$table_prefix = 'wp_';

define('WP_DEBUG', true);
define('WP_DEBUG_LOG', true);
define('WP_DEBUG_DISPLAY', false);

if (!defined('ABSPATH')) {
    define('ABSPATH', __DIR__ . '/');
}

require_once ABSPATH . 'wp-settings.php';
?>`, projectName)

	os.WriteFile(filepath.Join(projectPath, "wp-config.php"), []byte(configContent), 0644)
}

func configureLaravel(projectPath, projectName string) {
	fmt.Printf("%sConfiguring Laravel project...%s\n", shared.ColorYellow, shared.ColorReset)
	
	envPath := filepath.Join(projectPath, ".env")
	if _, err := os.Stat(envPath); err == nil {
		envContent := fmt.Sprintf(`APP_NAME=%s
APP_ENV=local
APP_KEY=
APP_DEBUG=true
APP_URL=https://%s.test

LOG_CHANNEL=stack
LOG_DEPRECATIONS_CHANNEL=null
LOG_LEVEL=debug

DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=%s_db
DB_USERNAME=root
DB_PASSWORD=

BROADCAST_DRIVER=log
CACHE_DRIVER=file
FILESYSTEM_DISK=local
QUEUE_CONNECTION=sync
SESSION_DRIVER=file
SESSION_LIFETIME=120

MEMCACHED_HOST=127.0.0.1

REDIS_HOST=127.0.0.1
REDIS_PASSWORD=null
REDIS_PORT=6379

MAIL_MAILER=smtp
MAIL_HOST=mailpit
MAIL_PORT=1025
MAIL_USERNAME=null
MAIL_PASSWORD=null
MAIL_ENCRYPTION=null
MAIL_FROM_ADDRESS="hello@example.com"
MAIL_FROM_NAME="${%s}"`, projectName, projectName, projectName, projectName)

		os.WriteFile(envPath, []byte(envContent), 0644)
	}
	
	fmt.Printf("%sGenerating Laravel application key...%s\n", shared.ColorYellow, shared.ColorReset)
	keyCmd := exec.Command("php", "artisan", "key:generate")
	keyCmd.Dir = projectPath
	keyCmd.Stdout = os.Stdout
	keyCmd.Stderr = os.Stderr
	
	if err := keyCmd.Run(); err != nil {
		fmt.Printf("%sWarning: Failed to generate Laravel key: %v%s\n", shared.ColorYellow, err, shared.ColorReset)
		fmt.Printf("%sPlease run 'php artisan key:generate' manually in the project directory%s\n", shared.ColorYellow, shared.ColorReset)
	} else {
		fmt.Printf("%s✓ Laravel application key generated%s\n", shared.ColorGreen, shared.ColorReset)
	}
	
	fmt.Printf("%s✓ Laravel project configured%s\n", shared.ColorGreen, shared.ColorReset)
}
