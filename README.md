# licenseflow-go

Official Go SDK for LicenseFlow.

## Installation

```bash
go get github.com/licenseflow/go-sdk
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"
	"github.com/licenseflow/go-sdk/pkg/licenseflow"
)

func main() {
	client := licenseflow.NewClient(licenseflow.Config{
		BaseURL: "https://your-project.supabase.co",
		APIKey:  "your-api-key",
	})

	// 1. Activate License
	res, err := client.Activate("XXXX-YYYY-ZZZZ-AAAA", "Production Server")
	if err != nil {
		log.Fatalf("Activation failed: %v", err)
	}
	fmt.Printf("Activated: %v\n", res["success"])

	// 2. Verify License (Uses internal cache)
	verify, err := client.Verify("XXXX-YYYY-ZZZZ-AAAA")
	if err != nil {
		log.Fatalf("Verification failed: %v", err)
	}
	fmt.Printf("Valid: %v\n", verify["valid"])
}
```

## Features

- **Standard Library**: Built using Go standard library for minimal dependencies.
- **Hardware ID**: Automatic hostname identification.
- **Thread-safe Caching**: Simple in-memory cache for verification.
- **Strongly Typed Errors**: Handle `RateLimitError` and `InvalidLicenseError` explicitly.

## Phase 5: Entitlements

Check feature access based on license tier:

```go
// Check boolean feature
if client.HasFeature(verification, "ai_features") {
    // Enable AI features
}

// Get raw entitlement value
if val := client.GetEntitlement(verification, "max_seats"); val != nil {
    fmt.Printf("Max seats: %v\n", val)
}
```

## Phase 5: Release Management

Check for updates and download new versions:

```go
// Check for updates
update, err := client.CheckForUpdates("prod_123", "v1.0.0", "stable")
if err != nil {
    log.Fatal(err)
}

if update != nil {
    fmt.Printf("New version: %s\n", update["version"])
    
    // Get download URL
    download, err := client.DownloadArtifact("XXXX-YYYY", update["id"].(string), "windows", "x64")
    if err == nil {
        fmt.Printf("Download URL: %s\n", download["url"])
    }
}
```

## Phase 5: Offline Licensing

Verify licenses without internet access:

```go
// Read license file
content, _ := os.ReadFile("license.lic") 
publicKeyHex := "YOUR_ORG_PUBLIC_KEY_HEX"

license, err := client.VerifyOfflineLicense(string(content), publicKeyHex)
if err != nil {
    log.Fatalf("Invalid offine license: %v", err)
}

fmt.Printf("Offline license valid! Customer: %v\n", license["customer_name"])
```
