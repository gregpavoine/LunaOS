package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"lunaos.local/bootstrap/services/intimacy"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: luna-vault <backup|restore> [options]")
	}
	switch args[0] {
	case "keygen":
		return runKeygen(args[1:], output)
	case "backup":
		return runBackup(args[1:], output)
	case "restore":
		return runRestore(args[1:], output)
	case "-h", "--help", "help":
		_, err := io.WriteString(output, "usage: luna-vault <keygen|backup|restore> [options]\n")
		return err
	default:
		return errors.New("unknown command; use backup or restore")
	}
}

func runKeygen(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("keygen", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	destination := flags.String("destination", "", "new owner-only 32-byte key file")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *destination == "" || flags.NArg() != 0 {
		return errors.New("keygen requires --destination")
	}
	if err := intimacy.GenerateOwnerKeyFile(*destination); err != nil {
		return err
	}
	_, err := io.WriteString(output, "owner-only key file created\n")
	return err
}

func runBackup(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("backup", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	database := flags.String("database", "", "path to the private vault database")
	vaultKeyFile := flags.String("vault-key-file", "", "owner-only vault key file")
	backupKeyFile := flags.String("backup-key-file", "", "independent owner-only backup key file")
	destination := flags.String("destination", "", "new encrypted backup path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *database == "" || *vaultKeyFile == "" || *backupKeyFile == "" || *destination == "" || flags.NArg() != 0 {
		return errors.New("backup requires --database, --vault-key-file, --backup-key-file, and --destination")
	}
	vaultKey, err := intimacy.ReadOwnerKeyFile(*vaultKeyFile)
	if err != nil {
		return err
	}
	defer clearBytes(vaultKey)
	backupKey, err := intimacy.ReadOwnerKeyFile(*backupKeyFile)
	if err != nil {
		return err
	}
	defer clearBytes(backupKey)
	vault, err := intimacy.Open(*database, vaultKey)
	if err != nil {
		return err
	}
	defer vault.Close()
	if err = vault.CreateEncryptedBackup(context.Background(), *destination, backupKey); err != nil {
		return err
	}
	_, err = io.WriteString(output, "encrypted vault backup created\n")
	return err
}

func runRestore(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("restore", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	backupFile := flags.String("backup-file", "", "encrypted backup file")
	backupKeyFile := flags.String("backup-key-file", "", "owner-only backup key file")
	vaultKeyFile := flags.String("vault-key-file", "", "owner-only vault key file")
	destination := flags.String("destination", "", "new restored database path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *backupFile == "" || *backupKeyFile == "" || *vaultKeyFile == "" || *destination == "" || flags.NArg() != 0 {
		return errors.New("restore requires --backup-file, --backup-key-file, --vault-key-file, and --destination")
	}
	backupKey, err := intimacy.ReadOwnerKeyFile(*backupKeyFile)
	if err != nil {
		return err
	}
	defer clearBytes(backupKey)
	vaultKey, err := intimacy.ReadOwnerKeyFile(*vaultKeyFile)
	if err != nil {
		return err
	}
	defer clearBytes(vaultKey)
	if err = intimacy.RestoreEncryptedBackup(context.Background(), *backupFile, *destination, backupKey, vaultKey); err != nil {
		return err
	}
	_, err = io.WriteString(output, "encrypted vault restored to new database\n")
	return err
}

func clearBytes(value []byte) {
	for index := range value {
		value[index] = 0
	}
}
