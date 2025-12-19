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
