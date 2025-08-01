# Changelog

All notable changes to Gecko will be documented in this file.

## [1.0.4] - 2025-08-01

### ✨ New Features

#### Project Management System
- **Universal Project Creator**: Complete project management system with support for multiple frameworks
- **Technology Stack Selection**: Organized menu system for PHP, JavaScript, and Static HTML projects
- **Project Operations**: Create, list, and delete projects with full lifecycle management

#### Supported Project Types

**PHP Frameworks:**
- Laravel - Modern PHP framework for web artisans
- WordPress - Popular CMS platform (downloaded from official source)
- Symfony - High performance PHP framework
- CodeIgniter - Simple & elegant PHP framework
- CakePHP - Rapid development framework
- Laminas - Enterprise-ready PHP framework

**JavaScript Frameworks:**
- React - Library for building user interfaces
- Vue.js - Progressive JavaScript framework
- Next.js - React framework for production
- Nuxt.js - Vue.js framework for production
- Angular - Platform for mobile & desktop web apps
- Svelte - Cybernetically enhanced web apps
- Astro - Build faster websites
- Vite - Next generation frontend tooling

**Static Projects:**
- Static HTML - Pure HTML/CSS/JS projects

#### Universal Database Support
- **Dynamic Database Validation**: Universal database checking system for all project types
- **Multi-Database Support**: Automatic MySQL and PostgreSQL integration

#### Official Sources Only
- **No Fallback Structures**: Removed all manual fallback project templates
- **Official Downloads**: All projects must be downloaded from official sources
  - WordPress from wordpress.org
  - React via `npx create-react-app`
  - Vue via `npx create-vue`
  - Next.js via `npx create-next-app`
  - And more...

### 📋 Menu Structure Updates
```
:: PROJECT MANAGEMENT
10. Create New Project    11. List Projects
12. Delete Project

:: TOOLS & TUNNELS (renumbered)
13. Switch PHP Version    14. Install Root CA (SSL)
15. Install Default SSL   16. Start/Stop Ngrok
17. Set Ngrok Auth Token  18. Start/Stop Cloudflare

:: APPLICATION
19. Activate/Deactivate Dev Mode    x. Exit
```

### 🛠️ Development Experience
- **Error Handling**: Comprehensive validation and user feedback
- **Build System**: Updated build process with Go module support

### 📁 File Structure
```
internal/service/
├── project_manager.go      # High-level project operations
├── template_manager.go     # Universal project creation engine
├── vhost.go               # Virtual host management
└── ...

cmd/gecko/
├── main.go                # Updated with project management handlers
└── ...

internal/cli/
├── menu.go                # Enhanced menu with PROJECT MANAGEMENT section
└── ...
```

### 🔄 Migration Notes
- All existing VHost functionality remains unchanged
- Database services (MySQL/PostgreSQL) work seamlessly with new projects
- SSL certificate generation integrates automatically with new projects
- Tunnel services (Ngrok/Cloudflare) compatible with all project types

### 🏗️ Architecture Benefits
- **Scalable**: Easy to add new project types and frameworks
- **Maintainable**: Configuration-driven approach reduces code duplication
- **Reliable**: Official sources ensure consistent project quality
- **Flexible**: Universal database support for any project type