package peer

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestUIDOfUnixSocketPairIsCurrentUser(t *testing.T) {
	// Short path: macOS limits Unix socket paths to about 104 bytes and t.TempDir() is longer.
	dir, err := os.MkdirTemp("/tmp", "peer")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	l, err := net.Listen("unix", filepath.Join(dir, "s.sock"))
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	accepted := make(chan net.Conn, 1)
	go func() {
		c, _ := l.Accept()
		accepted <- c
	}()
	client, err := net.Dial("unix", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	server := <-accepted
	if server == nil {
		t.Fatal("accept failed")
	}
	defer server.Close()

	uid, err := UID(server)
	if runtime.GOOS != "linux" {
		if !errors.Is(err, ErrUnsupported) || uid != -1 {
			t.Fatalf("UID = %d, %v; want -1, ErrUnsupported", uid, err)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if uid != os.Getuid() {
		t.Fatalf("UID = %d, want %d", uid, os.Getuid())
	}
}

func TestUIDRejectsNonUnixConnection(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	if uid, err := UID(a); err == nil || uid != -1 {
		t.Fatalf("UID = %d, %v; want -1 and an error", uid, err)
	}
}

func TestConnContextFailsClosedWithoutPeerCredentials(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	ctx := ConnContext(context.Background(), a)
	uid, ok := FromContext(ctx)
	if !ok || uid != -1 {
		t.Fatalf("FromContext = %d, %v; want -1, true", uid, ok)
	}
	if _, ok := FromContext(context.Background()); ok {
		t.Fatal("FromContext reported a UID on a bare context")
	}
}

func TestAllowedRejectsNegativeAllowedUID(t *testing.T) {
	if Allowed(WithUID(context.Background(), -1), -1) {
		t.Fatal("a failed peer read (-1) matched a -1 allowed UID")
	}
	if Allowed(WithUID(context.Background(), 5), -1) {
		t.Fatal("a negative allowed UID accepted a peer")
	}
}

func TestAllowedRejectsMissingOrFailedPeer(t *testing.T) {
	if Allowed(context.Background(), 5) {
		t.Fatal("a context with no peer UID was allowed")
	}
	if Allowed(WithUID(context.Background(), -1), 5) {
		t.Fatal("a failed peer read was allowed")
	}
}

func TestAllowedMatchesStoredUID(t *testing.T) {
	ctx := WithUID(context.Background(), 5)
	if !Allowed(ctx, 5) {
		t.Fatal("matching UID rejected")
	}
	if Allowed(ctx, 6) {
		t.Fatal("different UID accepted")
	}
}
