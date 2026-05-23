package main

import (
	"fmt"
	"os"

	licenseflow "github.com/licenseflow/licenseflow-go/pkg"
)

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		fmt.Fprintf(os.Stderr, "missing env %s\n", k)
		os.Exit(2)
	}
	return v
}

func main() {
	c := licenseflow.NewClient(mustEnv("LICENSEFLOW_API_URL"), mustEnv("LICENSEFLOW_API_KEY"))
	lk := mustEnv("LICENSE_KEY")
	rk := mustEnv("REVOKED_LICENSE_KEY")

	if _, err := c.Activate(lk, "ci-go"); err != nil {
		panic(err)
	}
	if v, _ := c.Verify(lk); !v.Valid {
		panic("active must verify")
	}
	if v, _ := c.Verify(rk); v.Valid {
		panic("revoked must not verify")
	}
	_, _ = c.Deactivate(lk)
	fmt.Println("Go SDK E2E ✓")
}