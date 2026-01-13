package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/go-ctf-platform/backend/internal/config"
	"github.com/go-ctf-platform/backend/internal/database"
	"github.com/go-ctf-platform/backend/internal/models"
	"github.com/go-ctf-platform/backend/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	fmt.Println("🌱 RootAccess CTF Database Seeder")
	fmt.Println("================================")

	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	database.ConnectDB(cfg.MongoURI, cfg.DBName)
	fmt.Println("✅ Connected to MongoDB")

	// Seed data
	seedUsers()
	seedChallenges()
	seedTeams()

	fmt.Println("\n🎉 Database seeding completed!")
	fmt.Println("================================")
	fmt.Println("Admin credentials:")
	fmt.Println("  Username: admin")
	fmt.Println("  Password: admin123")
	fmt.Println("  Email: admin@rootaccess.ctf")
}

func seedUsers() {
	fmt.Println("\n📦 Seeding users...")

	usersCollection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if admin already exists
	var existingAdmin models.User
	err := usersCollection.FindOne(ctx, bson.M{"username": "admin"}).Decode(&existingAdmin)
	if err == nil {
		fmt.Println("  ⚠️  Admin user already exists, skipping...")
		return
	}

	// Hash passwords
	adminPassHash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	user1PassHash, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
	user2PassHash, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
	user3PassHash, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
	user4PassHash, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
	user5PassHash, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)

	users := []interface{}{
		models.User{
			ID:            primitive.NewObjectID(),
			Username:      "admin",
			Email:         "admin@rootaccess.ctf",
			PasswordHash:  string(adminPassHash),
			Role:          "admin",
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		models.User{
			ID:            primitive.NewObjectID(),
			Username:      "alice",
			Email:         "alice@example.com",
			PasswordHash:  string(user1PassHash),
			Role:          "user",
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		models.User{
			ID:            primitive.NewObjectID(),
			Username:      "bob",
			Email:         "bob@example.com",
			PasswordHash:  string(user2PassHash),
			Role:          "user",
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		models.User{
			ID:            primitive.NewObjectID(),
			Username:      "charlie",
			Email:         "charlie@example.com",
			PasswordHash:  string(user3PassHash),
			Role:          "user",
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		models.User{
			ID:            primitive.NewObjectID(),
			Username:      "diana",
			Email:         "diana@example.com",
			PasswordHash:  string(user4PassHash),
			Role:          "user",
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		models.User{
			ID:            primitive.NewObjectID(),
			Username:      "eve",
			Email:         "eve@example.com",
			PasswordHash:  string(user5PassHash),
			Role:          "user",
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	result, err := usersCollection.InsertMany(ctx, users)
	if err != nil {
		log.Printf("  ❌ Error seeding users: %v", err)
		return
	}

	fmt.Printf("  ✅ Created %d users\n", len(result.InsertedIDs))
}

func seedChallenges() {
	fmt.Println("\n📦 Seeding challenges...")

	challengesCollection := database.GetCollection("challenges")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if challenges already exist
	count, err := challengesCollection.CountDocuments(ctx, bson.M{})
	if err == nil && count > 0 {
		fmt.Printf("  ⚠️  %d challenges already exist, skipping...\n", count)
		return
	}

	challenges := []interface{}{
		// Web Exploitation
		models.Challenge{
			ID:          primitive.NewObjectID(),
			Title:       "SQL Injection 101",
			Description: "# SQL Injection 101\n\nWelcome to your first SQL injection challenge!\n\n## Objective\nFind the flag hidden in the database by exploiting a vulnerable login form.\n\n## Hints\n- Try common SQL injection payloads\n- The flag format is `CTF{...}`\n\n## Resources\n- [OWASP SQL Injection](https://owasp.org/www-community/attacks/SQL_Injection)",
			Category:    "web",
			Difficulty:  "easy",
			MaxPoints:   200,
			MinPoints:   100,
			Decay:       15,
			SolveCount:  0,
			FlagHash:    utils.HashFlag("CTF{sql_injection_is_fun}"),
			Files:       []string{},
		},
		models.Challenge{
			ID:          primitive.NewObjectID(),
			Title:       "XSS Adventure",
			Description: "# XSS Adventure\n\nExplore the world of Cross-Site Scripting!\n\n## Objective\nFind and exploit an XSS vulnerability to retrieve the admin's cookie containing the flag.\n\n## The Challenge\nA guestbook application allows users to leave messages. However, it seems like the developers forgot to sanitize user input...\n\n## Notes\n- The flag is stored in a cookie named `admin_flag`\n- You'll need to find a way to exfiltrate it",
			Category:    "web",
			Difficulty:  "medium",
			MaxPoints:   300,
			MinPoints:   100,
			Decay:       12,
			SolveCount:  0,
			FlagHash:    utils.HashFlag("CTF{xss_cookie_monster}"),
			Files:       []string{},
		},

		// Cryptography
		models.Challenge{
			ID:          primitive.NewObjectID(),
			Title:       "Caesar's Secret",
			Description: "# Caesar's Secret\n\nJulius Caesar used a simple cipher to communicate with his generals. Can you decode this message?\n\n## Encrypted Message\n```\nFWI{fdhvdu_flskhu_lv_hdvb}\n```\n\n## Hints\n- The shift might not be what you expect\n- Try different shift values",
			Category:    "crypto",
			Difficulty:  "easy",
			MaxPoints:   150,
			MinPoints:   100,
			Decay:       20,
			SolveCount:  0,
			FlagHash:    utils.HashFlag("CTF{caesar_cipher_is_easy}"),
			Files:       []string{},
		},
		models.Challenge{
			ID:          primitive.NewObjectID(),
			Title:       "RSA Breakdown",
			Description: "# RSA Breakdown\n\nWe intercepted an RSA-encrypted message, but the implementation seems weak...\n\n## Given\n```\nn = 3233\ne = 17\nc = 2790\n```\n\n## Objective\nDecrypt the ciphertext `c` to find the flag.\n\n## Notes\n- n = p * q where p and q are small primes\n- The flag is the decrypted plaintext converted to ASCII",
			Category:    "crypto",
			Difficulty:  "hard",
			MaxPoints:   500,
			MinPoints:   100,
			Decay:       8,
			SolveCount:  0,
			FlagHash:    utils.HashFlag("CTF{rsa_small_primes_bad}"),
			Files:       []string{},
		},

		// Forensics
		models.Challenge{
			ID:          primitive.NewObjectID(),
			Title:       "Memory Forensics",
			Description: "# Memory Forensics\n\nWe captured a memory dump from a compromised machine. Your task is to analyze it and find the hidden flag.\n\n## Tools You Might Need\n- Volatility Framework\n- Strings\n- Hex editors\n\n## Hints\n- Look for suspicious processes\n- Check the clipboard contents",
			Category:    "forensics",
			Difficulty:  "medium",
			MaxPoints:   350,
			MinPoints:   100,
			Decay:       10,
			SolveCount:  0,
			FlagHash:    utils.HashFlag("CTF{memory_never_forgets}"),
			Files:       []string{},
		},

		// Steganography
		models.Challenge{
			ID:          primitive.NewObjectID(),
			Title:       "Hidden in Plain Sight",
			Description: "# Hidden in Plain Sight\n\nThis image looks perfectly normal, but appearances can be deceiving...\n\n## Objective\nFind the flag hidden within the image.\n\n## Tools\n- steghide\n- zsteg\n- binwalk\n\n## Hints\n- Sometimes the simplest approach works\n- Check the LSB (Least Significant Bits)",
			Category:    "steganography",
			Difficulty:  "easy",
			MaxPoints:   200,
			MinPoints:   100,
			Decay:       15,
			SolveCount:  0,
			FlagHash:    utils.HashFlag("CTF{hidden_in_pixels}"),
			Files:       []string{},
		},

		// Reverse Engineering
		models.Challenge{
			ID:          primitive.NewObjectID(),
			Title:       "CrackMe Easy",
			Description: "# CrackMe Easy\n\nYour first reverse engineering challenge! Analyze this simple binary to find the correct password.\n\n## Objective\nFind the password that makes the program print 'Access Granted!'\n\n## Tools\n- Ghidra\n- IDA Free\n- radare2\n- ltrace/strace",
			Category:    "reverse",
			Difficulty:  "easy",
			MaxPoints:   250,
			MinPoints:   100,
			Decay:       12,
			SolveCount:  0,
			FlagHash:    utils.HashFlag("CTF{reverse_me_baby}"),
			Files:       []string{},
		},

		// Misc
		models.Challenge{
			ID:          primitive.NewObjectID(),
			Title:       "Sanity Check",
			Description: "# Sanity Check\n\nWelcome to RootAccess CTF! 🎉\n\nThis is your first challenge. The flag is:\n\n```\nCTF{welcome_to_rootaccess}\n```\n\nJust submit it to get your first points!",
			Category:    "misc",
			Difficulty:  "easy",
			MaxPoints:   50,
			MinPoints:   50,
			Decay:       100,
			SolveCount:  0,
			FlagHash:    utils.HashFlag("CTF{welcome_to_rootaccess}"),
			Files:       []string{},
		},
	}

	result, err := challengesCollection.InsertMany(ctx, challenges)
	if err != nil {
		log.Printf("  ❌ Error seeding challenges: %v", err)
		return
	}

	fmt.Printf("  ✅ Created %d challenges\n", len(result.InsertedIDs))
}

func seedTeams() {
	fmt.Println("\n📦 Seeding teams...")

	teamsCollection := database.GetCollection("teams")
	usersCollection := database.GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if teams already exist
	count, err := teamsCollection.CountDocuments(ctx, bson.M{})
	if err == nil && count > 0 {
		fmt.Printf("  ⚠️  %d teams already exist, skipping...\n", count)
		return
	}

	// Get user IDs for team members
	var alice, bob, charlie, diana, eve models.User
	usersCollection.FindOne(ctx, bson.M{"username": "alice"}).Decode(&alice)
	usersCollection.FindOne(ctx, bson.M{"username": "bob"}).Decode(&bob)
	usersCollection.FindOne(ctx, bson.M{"username": "charlie"}).Decode(&charlie)
	usersCollection.FindOne(ctx, bson.M{"username": "diana"}).Decode(&diana)
	usersCollection.FindOne(ctx, bson.M{"username": "eve"}).Decode(&eve)

	if alice.ID.IsZero() || bob.ID.IsZero() {
		fmt.Println("  ⚠️  Users not found, skipping team creation...")
		return
	}

	teams := []interface{}{
		models.Team{
			ID:          primitive.NewObjectID(),
			Name:        "Team Alpha",
			Description: "We are Team Alpha! First to solve, first to win.",
			LeaderID:    alice.ID,
			MemberIDs:   []primitive.ObjectID{alice.ID, bob.ID},
			InviteCode:  generateInviteCode(),
			Score:       0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		models.Team{
			ID:          primitive.NewObjectID(),
			Name:        "Team Beta",
			Description: "Beta testers of the CTF world. We find all the bugs!",
			LeaderID:    charlie.ID,
			MemberIDs:   []primitive.ObjectID{charlie.ID, diana.ID, eve.ID},
			InviteCode:  generateInviteCode(),
			Score:       0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	result, err := teamsCollection.InsertMany(ctx, teams)
	if err != nil {
		log.Printf("  ❌ Error seeding teams: %v", err)
		return
	}

	fmt.Printf("  ✅ Created %d teams\n", len(result.InsertedIDs))
}

func generateInviteCode() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
