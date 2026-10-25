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
	path := flags.String("path", "/etc/lunaos/owner-token", "owner credential file path")
	reveal := flags.Bool("reveal-token", false, "print the generated credential once")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(errorsOutput, "usage: luna-owner-setup [--path FILE] [--reveal-token]")
		return 2
	}
	token, err := identity.ProvisionOwnerToken(*path)
	if err != nil {
		fmt.Fprintf(errorsOutput, "Owner credential provisioning failed: %v\n", err)
		return 1
	}
	fmt.Fprintf(output, "Owner credential provisioned at %s\n", filepath.Clean(*path))
	if *reveal {
		fmt.Fprintln(output, token)
	}
	return 0
}
