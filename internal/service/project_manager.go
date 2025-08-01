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
	projectsDir = `C:\Gecko\www`
)

type ProjectTemplate struct {
	Name        string
	Description string
	Type        string
	RequiresPHP bool
	RequiresDB  bool
}

var projectTemplates = map[string]ProjectTemplate{
	"laravel":     {Name: "Laravel", Description: "PHP Framework for web artisans", Type: "php", RequiresPHP: true, RequiresDB: true},
	"wordpress":   {Name: "WordPress", Description: "Popular CMS platform", Type: "php", RequiresPHP: true, RequiresDB: true},
	"symfony":     {Name: "Symfony", Description: "High performance PHP framework", Type: "php", RequiresPHP: true, RequiresDB: true},
	"codeigniter": {Name: "CodeIgniter", Description: "Simple & elegant PHP framework", Type: "php", RequiresPHP: true, RequiresDB: true},
	"cakephp":     {Name: "CakePHP", Description: "Rapid development framework", Type: "php", RequiresPHP: true, RequiresDB: true},
	"laminas":     {Name: "Laminas", Description: "Enterprise-ready PHP framework", Type: "php", RequiresPHP: true, RequiresDB: true},
	"react":       {Name: "React", Description: "Library for building user interfaces", Type: "javascript", RequiresPHP: false, RequiresDB: false},
	"vue":         {Name: "Vue.js", Description: "Progressive JavaScript framework", Type: "javascript", RequiresPHP: false, RequiresDB: false},
	"nextjs":      {Name: "Next.js", Description: "React framework for production", Type: "javascript", RequiresPHP: false, RequiresDB: false},
	"nuxtjs":      {Name: "Nuxt.js", Description: "Vue.js framework for production", Type: "javascript", RequiresPHP: false, RequiresDB: false},
	"angular":     {Name: "Angular", Description: "Platform for mobile & desktop web apps", Type: "javascript", RequiresPHP: false, RequiresDB: false},
	"svelte":      {Name: "Svelte", Description: "Cybernetically enhanced web apps", Type: "javascript", RequiresPHP: false, RequiresDB: false},
	"astro":       {Name: "Astro", Description: "Build faster websites", Type: "javascript", RequiresPHP: false, RequiresDB: false},
	"vite":        {Name: "Vite", Description: "Next generation frontend tooling", Type: "javascript", RequiresPHP: false, RequiresDB: false},
	"static":      {Name: "Static HTML", Description: "Simple HTML/CSS/JS website", Type: "html", RequiresPHP: false, RequiresDB: false},
}

func CreateProject(projectType, projectName string, reader *bufio.Reader) {
	template, exists := projectTemplates[projectType]
	if !exists {
		fmt.Printf("%sInvalid project type: %s%s\n", shared.ColorRed, projectType, shared.ColorReset)
		return
	}

	projectPath := filepath.Join(projectsDir, projectName)
	
	if _, err := os.Stat(projectPath); err == nil {
		fmt.Printf("%sProject '%s' already exists. Do you want to replace it? (y/n): %s", shared.ColorYellow, projectName, shared.ColorReset)
		confirm, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(confirm)) != "y" {
			fmt.Println(shared.ColorYellow, "Project creation cancelled.", shared.ColorReset)
			return
		}
		
		fmt.Printf("%sRemoving existing project...%s\n", shared.ColorYellow, shared.ColorReset)
		os.RemoveAll(projectPath)
	}

	fmt.Printf("%sCreating %s project '%s'...%s\n", shared.ColorGreen, template.Name, projectName, shared.ColorReset)

	var success bool
	switch projectType {
	case "laravel":
		CreateLaravelProject(projectPath, projectName)
		success = true
	case "wordpress":
		CreateWordPressProject(projectPath, projectName)
		success = true
	case "symfony":
		CreateSymfonyProject(projectPath, projectName)
		success = true
	case "codeigniter":
		CreateCodeIgniterProject(projectPath, projectName)
		success = true
	case "cakephp":
		CreateCakePHPProject(projectPath, projectName)
		success = true
	case "laminas":
		CreateLaminasProject(projectPath, projectName)
		success = true
	case "react":
		CreateReactProject(projectPath, projectName)
		success = true
	case "vue":
		CreateVueProject(projectPath, projectName)
		success = true
	case "nextjs":
		CreateNextJSProject(projectPath, projectName)
		success = true
	case "nuxtjs":
		CreateNuxtJSProject(projectPath, projectName)
		success = true
	case "angular":
		CreateAngularProject(projectPath, projectName)
		success = true
	case "svelte":
		CreateSvelteProject(projectPath, projectName)
		success = true
	case "astro":
		CreateAstroProject(projectPath, projectName)
		success = true
	case "vite":
		CreateViteProject(projectPath, projectName)
		success = true
	case "static":
		CreateStaticProject(projectPath, projectName)
		success = true
	default:
		fmt.Printf("%sTemplate '%s' not implemented yet.%s\n", shared.ColorYellow, projectType, shared.ColorReset)
		return
	}

	if !success {
		os.RemoveAll(projectPath)
		return
	}

	if template.Type == "javascript" {
		fmt.Printf("%s✓ Project '%s' created successfully!%s\n", shared.ColorGreen, projectName, shared.ColorReset)
		fmt.Printf("%sFor development: cd %s && npm start/npm run dev%s\n", shared.ColorGreen, projectName, shared.ColorReset)
		return
	}

	fmt.Printf("%sCreating virtual host for %s.test...%s\n", shared.ColorYellow, projectName, shared.ColorReset)
	
	CreateVirtualHost(projectName+".test", "y")

	fmt.Printf("%s✓ Project '%s' created successfully!%s\n", shared.ColorGreen, projectName, shared.ColorReset)
	fmt.Printf("%sYou can access it at: https://%s.test%s\n", shared.ColorGreen, projectName, shared.ColorReset)
}

func ListProjects() ([]string, error) {
	var projects []string
	
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "." && entry.Name() != ".." {
			projects = append(projects, entry.Name())
		}
	}

	return projects, nil
}

func DeleteProject(projectName string, reader *bufio.Reader) {
	projectPath := filepath.Join(projectsDir, projectName)
	
	fmt.Printf("%sDeleting project '%s'...%s\n", shared.ColorYellow, projectName, shared.ColorReset)
	
	var hasDatabase bool
	var projectType string
	
	if _, err := os.Stat(filepath.Join(projectPath, "composer.json")); err == nil {
		if _, err := os.Stat(filepath.Join(projectPath, "artisan")); err == nil {
			projectType = "Laravel"
			hasDatabase = true
		}
	}
	
	if _, err := os.Stat(filepath.Join(projectPath, "wp-config.php")); err == nil {
		projectType = "WordPress"
		hasDatabase = true
	}
	
	if _, err := os.Stat(filepath.Join(projectPath, "composer.json")); err == nil {
		if _, err := os.Stat(filepath.Join(projectPath, "symfony.lock")); err == nil {
			projectType = "Symfony"
			hasDatabase = true
		}
	}
	
	if hasDatabase && reader != nil {
		dbName := projectName + "_db"
		fmt.Printf("%s%s project detected with database '%s'.%s\n", shared.ColorYellow, projectType, dbName, shared.ColorReset)
		fmt.Printf("%sDo you want to delete the database '%s' as well? (y/n): %s", shared.ColorYellow, dbName, shared.ColorReset)
		
		deleteDB, _ := reader.ReadString('\n')
		deleteDB = strings.TrimSpace(strings.ToLower(deleteDB))
		
		if deleteDB == "y" || deleteDB == "yes" {
			deleteMySQLDatabase(dbName)
		}
	}
	
	if err := os.RemoveAll(projectPath); err != nil {
		fmt.Printf("%sError removing project directory: %v%s\n", shared.ColorRed, err, shared.ColorReset)
		return
	}
	
	DeleteVirtualHost(projectName + ".test")
	
	fmt.Printf("%s✓ Project '%s' deleted successfully!%s\n", shared.ColorGreen, projectName, shared.ColorReset)
}

func deleteMySQLDatabase(dbName string) {
	fmt.Printf("%sDeleting database '%s'...%s\n", shared.ColorYellow, dbName, shared.ColorReset)
	
	dropDBSQL := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`;", dbName)
	
	cmd := exec.Command("mysql", "-u", "root", "-e", dropDBSQL)
	if err := cmd.Run(); err != nil {
		fmt.Printf("%sWarning: Failed to delete database '%s': %v%s\n", shared.ColorYellow, dbName, err, shared.ColorReset)
		fmt.Printf("%sYou may need to delete the database manually: DROP DATABASE %s;%s\n", shared.ColorYellow, dbName, shared.ColorReset)
	} else {
		fmt.Printf("%s✓ Database '%s' deleted successfully%s\n", shared.ColorGreen, dbName, shared.ColorReset)
	}
}
