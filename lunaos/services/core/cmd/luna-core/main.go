package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"lunaos.local/bootstrap/services/core"
	"lunaos.local/bootstrap/services/intimacy"
)

func main() {
	if err := run(); err != nil {
		log.Printf("LUNA Core stopped: %v", err)
		os.Exit(1)
	}
}

func run() error {
	tokenPath := os.Getenv("LUNA_OWNER_TOKEN_FILE")
	if tokenPath == "" {
		return fmt.Errorf("LUNA_OWNER_TOKEN_FILE is required")
	}
	token, err := readOwnerToken(tokenPath)
	if err != nil {
		return err
	}
	database := os.Getenv("LUNA_DATABASE")
	if database == "" {
		return fmt.Errorf("LUNA_DATABASE is required")
	}
	policySocket := os.Getenv("LUNA_POLICY_SOCKET")
	if policySocket == "" {
		return fmt.Errorf("LUNA_POLICY_SOCKET is required")
	}
	address := os.Getenv("LUNA_LISTEN")
	if address == "" {
		address = "127.0.0.1:17766"
	}
	if err := requireLoopbackAddress(address); err != nil {
		return err
	}
	policy, err := core.NewUnixPolicyClient(policySocket, token)
	if err != nil {
		return err
	}
	engine, err := core.Open(database, token, policy, core.ChecksumRunner{})
	if err != nil {
		return err
	}
	defer engine.Close()
	var vault *intimacy.Vault
	vaultPath, keyPath := os.Getenv("LUNA_INTIMACY_DATABASE"), os.Getenv("LUNA_INTIMACY_KEY_FILE")
	if vaultPath != "" || keyPath != "" {
		if vaultPath == "" || keyPath == "" {
			return fmt.Errorf("LUNA_INTIMACY_DATABASE and LUNA_INTIMACY_KEY_FILE must be configured together")
		}
		key, keyErr := readVaultKey(keyPath)
		if keyErr != nil {
			return keyErr
		}
		vault, err = intimacy.Open(vaultPath, key)
		for i := range key {
			key[i] = 0
		}
		if err != nil {
			return fmt.Errorf("intimacy vault unavailable: %w", err)
		}
		defer vault.Close()
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	server := &http.Server{
		Handler:           core.NewHTTPHandler(engine, token, vault),
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       10 * time.Second,
		MaxHeaderBytes:    8192,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go runScheduler(ctx, engine)
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(listener)
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func readVaultKey(path string) ([]byte, error) {
	return intimacy.ReadOwnerKeyFile(path)
}

func runScheduler(ctx context.Context, engine *core.Engine) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := engine.Tick(ctx); err != nil {
				log.Printf("scheduler deferred work: %T", err)
			}
		}
	}
}

func readOwnerToken(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", fmt.Errorf("owner credential unavailable")
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return "", fmt.Errorf("owner credential must be a private regular file")
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok && uint64(stat.Uid) != uint64(os.Getuid()) {
		return "", fmt.Errorf("owner credential owner mismatch")
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("owner credential unavailable")
	}
	token := strings.TrimSpace(string(bytes))
	if len(token) != 64 {
		return "", fmt.Errorf("owner credential has invalid format")
	}
	for _, r := range token {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return "", fmt.Errorf("owner credential has invalid format")
		}
	}
	return token, nil
}

func requireLoopbackAddress(address string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid LUNA_LISTEN address")
	}
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("LUNA Core may bind only to a loopback address")
	}
	return nil
}
