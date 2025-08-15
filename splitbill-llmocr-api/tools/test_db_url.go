package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Println("❌ DATABASE_URL environment variable is not set")
		os.Exit(1)
	}

	fmt.Printf("🔍 Testing DATABASE_URL: %s\n", maskPassword(databaseURL))

	// Check prefix
	if !strings.HasPrefix(databaseURL, "postgresql://") && !strings.HasPrefix(databaseURL, "postgres://") {
		fmt.Println("❌ DATABASE_URL must start with 'postgresql://' or 'postgres://'")
		os.Exit(1)
	}

	// Parse URL
	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		fmt.Printf("❌ Failed to parse DATABASE_URL: %v\n", err)
		os.Exit(1)
	}

	// Validate components
	errors := []string{}

	if parsedURL.Hostname() == "" {
		errors = append(errors, "missing hostname")
	} else {
		fmt.Printf("✅ Hostname: %s\n", parsedURL.Hostname())
	}

	if parsedURL.Port() == "" {
		fmt.Printf("✅ Port: 5432 (default)\n")
	} else {
		fmt.Printf("✅ Port: %s\n", parsedURL.Port())
	}

	if parsedURL.User == nil || parsedURL.User.Username() == "" {
		errors = append(errors, "missing username")
	} else {
		fmt.Printf("✅ Username: %s\n", parsedURL.User.Username())
	}

	if parsedURL.User != nil {
		if _, ok := parsedURL.User.Password(); !ok {
			errors = append(errors, "missing password")
		} else {
			fmt.Println("✅ Password: [present]")
		}
	}

	dbName := strings.TrimPrefix(parsedURL.Path, "/")
	if dbName == "" {
		errors = append(errors, "missing database name")
	} else {
		fmt.Printf("✅ Database: %s\n", dbName)
	}

	// Check for SSL mode
	if strings.Contains(databaseURL, "sslmode=require") {
		fmt.Println("✅ SSL Mode: require")
	} else {
		fmt.Println("⚠️  SSL Mode: not specified (Supabase requires 'require')")
	}

	if len(errors) > 0 {
		fmt.Println("\n❌ Validation errors:")
		for _, err := range errors {
			fmt.Printf("   - %s\n", err)
		}
		os.Exit(1)
	}

	fmt.Println("\n✅ DATABASE_URL is valid!")
	fmt.Println("\n💡 If you're using Supabase, make sure your connection string looks like:")
	fmt.Println("   postgresql://postgres.[project-ref]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres?sslmode=require")
}

func maskPassword(databaseURL string) string {
	// Simple password masking for display
	if strings.Contains(databaseURL, "@") {
		parts := strings.Split(databaseURL, "@")
		if len(parts) == 2 {
			userPass := parts[0]
			if strings.Contains(userPass, ":") {
				userParts := strings.Split(userPass, ":")
				if len(userParts) >= 3 {
					// postgresql://username:password@host
					return fmt.Sprintf("%s:***@%s", userParts[0], parts[1])
				}
			}
		}
	}
	return databaseURL
}
