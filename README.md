# LicenseFlow Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/licenseflow/go-sdk)](https://pkg.go.dev/github.com/licenseflow/go-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**Stop Building Licensing Infrastructure. Start Shipping Software.**

The official Go SDK for [LicenseFlow](https://licenseflow.dev). Protect your intellectual property, enforce entitlements, and manage software distribution with minimal dependencies.

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
        BaseURL:      "https://api.licenseflow.dev",
        APIKey:       "lf_live_xxxxxxxxxxxx",
        CacheTTL:     5 * time.Minute,   // Cache TTL (default: 5m)
        GracePeriod:  72 * time.Hour,    // Offline grace (default: 72h)
    })

    res, err := client.Activate("XXXX-YYYY-ZZZZ-AAAA", "Production Server", "")
    if err != nil {
        log.Fatalf("Activation failed: %v", err)
    }
    fmt.Printf("Activated: %v\n", res["success"])

    verify, err := client.Verify("XXXX-YYYY-ZZZZ-AAAA", "")
    if err != nil {
        log.Fatalf("Verification failed: %v", err)
    }
    fmt.Printf("Valid: %v\n", verify["valid"])
}
```

---

## Entitlement Caching

The SDK's `EntitlementCache` (built into the client) uses an in-memory TTL cache with an optional offline grace period. Calls to `Verify` return cached results within the TTL window, avoiding redundant network requests.

```go
client := licenseflow.NewClient(licenseflow.Config{
    BaseURL:     "https://api.licenseflow.dev",
    APIKey:      "lf_live_xxxxxxxxxxxx",
    CacheTTL:    5 * time.Minute,   // Serve from cache for 5 minutes
    GracePeriod: 72 * time.Hour,    // Use stale cache for up to 72h when API is down
})

// First call fetches from API and populates cache
result, _ := client.Verify("XXXX-YYYY-ZZZZ-AAAA", "")

// Subsequent calls served from memory (zero latency)
result, _ = client.Verify("XXXX-YYYY-ZZZZ-AAAA", "")

// During outage: cache used within grace window, ErrOfflineGraceExpired after
```

| Scenario | Result |
|---|---|
| Cache hit within TTL | Instant cached response |
| Cache miss / TTL expired | Live API call, refreshes cache |
| API down, within grace | Stale cache returned |
| API down, grace expired | `ErrOfflineGraceExpired` error |

---

## API Reference

### Core Methods

| Method | Description |
|--------|-------------|
| `Activate(licenseKey, deviceName, envID)` | Activate on a device |
| `Verify(licenseKey, envID)` | Verify license (cached) |
| `Deactivate(licenseKey, envID)` | Deactivate from a device |
| `RecordUsage(licenseKey, metricName, value)` | Track usage metrics |
| `GetHardwareID()` | Get hostname-based device ID |

### Entitlements

```go
if client.HasFeature(verification, "ai_features") {
    enableAI()
}

val := client.GetEntitlement(verification, "max_seats")
fmt.Printf("Max seats: %v\n", val)
```

### Floating Licenses (Leases)

```go
lease, err := client.CheckoutLicense("XXXX-YYYY", 3600, "ci-runner-1", "ci_runner")
fmt.Printf("Lease key: %s\n", lease["lease_key"])

err = client.CheckinLicense(lease["lease_key"].(string))

status, err := client.GetLeaseStatus(lease["lease_key"].(string))
```

### Credits

```go
result, err := client.ConsumeCredits(100, "AI generation", "", "")
fmt.Printf("Remaining: %v\n", result["remaining"])

balance, err := client.GetCreditsBalance("", "")
```

### Release Management

```go
update, err := client.CheckForUpdates("prod_123", "v1.0.0", "stable")
if update != nil {
    download, _ := client.DownloadArtifact("XXXX-YYYY", update["id"].(string), "linux", "amd64")
    fmt.Printf("Download: %s\n", download["url"])
}
```

### Offline Licensing

```go
content, _ := os.ReadFile("license.lic")
license, err := client.VerifyOfflineLicense(string(content), "ORG_PUBLIC_KEY_HEX")
```

### Heartbeat

```go
client.StartHeartbeat("XXXX-YYYY", 60) // seconds
defer client.StopHeartbeat()
```

---

## Error Handling

```go
import "github.com/licenseflow/go-sdk/pkg/licenseflow"

_, err := client.Activate(key, name, "")
if err != nil {
    switch err.(type) {
    case *licenseflow.RateLimitError:
        log.Println("Rate limit exceeded")
    case *licenseflow.InvalidLicenseError:
        log.Println("Invalid license")
    default:
        log.Printf("Error: %v", err)
    }
}
```

## Features

- **Standard Library** — Minimal dependencies, built on `net/http`
- **Thread-safe Caching** — Concurrent-safe in-memory verification cache with configurable TTL
- **Offline Grace Period** — Continue operating for up to 72h without connectivity
- **Environment Scoping** — Isolate licenses per deployment environment
- **Ed25519 Offline** — Cryptographic offline license verification
- **Goroutine-safe** — All cache operations protected by `sync.RWMutex`

## License

MIT

## Links

- 📖 [Documentation](https://docs.licenseflow.dev)
- 🐛 [Issues](https://github.com/LicenseFlow/sdk-go/issues)
- 🏠 [Homepage](https://licenseflow.dev)
