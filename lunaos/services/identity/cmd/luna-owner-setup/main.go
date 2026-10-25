package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"lunaos.local/bootstrap/services/identity"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, output, errorsOutput io.Writer) int {
	flags := flag.NewFlagSet("luna-owner-setup", flag.ContinueOnError)
	flags.SetOutput(errorsOutput)
	path := flags.String("path", "/var/lib/lunaos-identity/owner-token", "owner API credential file path")
	policyPath := flags.String("policy-path", "/var/lib/lunaos-identity/policy-token", "Core-to-Broker credential file path")
	reveal := flags.Bool("reveal-token", false, "print the generated credential once")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(errorsOutput, "usage: luna-owner-setup [--path FILE] [--policy-path FILE] [--reveal-token]")
		return 2
	}
	token, _, err := identity.ProvisionServiceCredentials(*path, *policyPath)
	if err != nil {
		fmt.Fprintf(errorsOutput, "Owner credential provisioning failed: %v\n", err)
		return 1
	}
	fmt.Fprintf(output, "Owner and policy credentials provisioned at %s and %s\n", filepath.Clean(*path), filepath.Clean(*policyPath))
	if *reveal {
		fmt.Fprintln(output, token)
	}
	return 0
}
