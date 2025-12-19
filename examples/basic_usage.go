package main

import (
	"fmt"
	"log"

	"github.com/licenseflow/go-sdk/pkg/licenseflow"
)

func main() {
	client := licenseflow.NewClient(licenseflow.Config{
		BaseURL: "https://api.test",
		APIKey:  "test-api-key",
	})

	fmt.Println("--- LicenseFlow Go Example ---")

	// 1. Activate
	fmt.Println("Activating license...")
	res, err := client.Activate("DEMO-KEY", "Go-Worker")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("Result: %v\n", res)

	// 2. Verify
	fmt.Println("Verifying license...")
	verify, err := client.Verify("DEMO-KEY")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("Is Valid: %v\n", verify["valid"])

	// 3. Deactivate
	fmt.Println("Deactivating license...")
	deactivation, err := client.Deactivate("DEMO-KEY")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("Result: %v\n", deactivation)
}
