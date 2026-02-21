package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Uttam-Mahata/RootAccess/backend/internal/config"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/database"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/repositories"
	"github.com/Uttam-Mahata/RootAccess/backend/internal/services"
)

var adminService *services.AdminService

func main() {
	cfg := config.LoadConfig()

	// Connect to the appropriate database based on DB_TYPE
	if cfg.DBType == "turso" {
		database.ConnectTurso(cfg.TursoURL, cfg.TursoToken)
	} else {
		database.ConnectDB(cfg.MongoURI, cfg.DBName)
	}

	// Initialize repository and service layers
	userRepo := repositories.NewUserRepository()
	adminService = services.NewAdminService(userRepo)

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("╔═══════════════════════════════════════════╗")
	fmt.Println("║   RootAccess CTF - Admin Management      ║")
	fmt.Println("╚═══════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("1. Create Admin User")
	fmt.Println("2. Promote User to Admin")
	fmt.Println("3. Demote Admin to User")
	fmt.Println("4. List All Users")
	fmt.Println("5. Exit")
	fmt.Println()
	fmt.Print("Choose an option: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		createAdminUser(reader)
	case "2":
		promoteToAdmin(reader)
	case "3":
		demoteToUser(reader)
	case "4":
		listAllUsers()
	case "5":
		fmt.Println("Goodbye!")
		os.Exit(0)
	default:
		fmt.Println("Invalid option!")
		os.Exit(1)
	}
}

func createAdminUser(reader *bufio.Reader) {
	fmt.Println("\n=== Create Admin User ===")

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Use admin service to create user
	if err := adminService.CreateAdminUser(username, email, password); err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	fmt.Printf("\n✅ Admin user '%s' created successfully!\n", username)
	fmt.Println("   Email:", email)
	fmt.Println("   Role: admin")
	fmt.Println("   Email verified: true")
}

func promoteToAdmin(reader *bufio.Reader) {
	fmt.Println("\n=== Promote User to Admin ===")

	fmt.Print("Enter username or email: ")
	identifier, _ := reader.ReadString('\n')
	identifier = strings.TrimSpace(identifier)

	// Use admin service to promote user
	user, err := adminService.PromoteToAdmin(identifier)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n✅ User '%s' promoted to admin!\n", user.Username)
	fmt.Println("   Email:", user.Email)
	fmt.Println("   New Role: admin")
}

func demoteToUser(reader *bufio.Reader) {
	fmt.Println("\n=== Demote Admin to User ===")

	fmt.Print("Enter username or email: ")
	identifier, _ := reader.ReadString('\n')
	identifier = strings.TrimSpace(identifier)

	// Use admin service to demote user
	user, err := adminService.DemoteToUser(identifier)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n✅ User '%s' demoted to regular user!\n", user.Username)
	fmt.Println("   Email:", user.Email)
	fmt.Println("   New Role: user")
}

func listAllUsers() {
	fmt.Println("\n=== All Users ===")
	fmt.Println()

	// Use admin service to get all users
	users, err := adminService.GetAllUsers()
	if err != nil {
		log.Fatal("Failed to fetch users:", err)
	}

	fmt.Printf("%-20s %-30s %-10s %-15s\n", "USERNAME", "EMAIL", "ROLE", "VERIFIED")
	fmt.Println(strings.Repeat("─", 80))

	if len(users) == 0 {
		fmt.Println("No users found.")
		return
	}

	for _, user := range users {
		verified := "No"
		if user.EmailVerified {
			verified = "Yes"
		}
		roleDisplay := user.Role
		if user.Role == "admin" {
			roleDisplay = "🔑 admin"
		}
		fmt.Printf("%-20s %-30s %-10s %-15s\n", user.Username, user.Email, roleDisplay, verified)
	}

	fmt.Printf("\nTotal users: %d\n", len(users))
}
