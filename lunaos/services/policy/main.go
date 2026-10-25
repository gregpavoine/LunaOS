// Policy broker: no shell, root operations, sensors, payment or network adapters.
package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type request struct {
	GoalID  string `json:"goal_id"`
	Action  string `json:"action"`
	Privacy string `json:"privacy_class"`
	Grant   string `json:"grant,omitempty"`
}
type response struct {
	Allowed bool   `json:"allowed"`
	Code    string `json:"code"`
	Grant   string `json:"grant,omitempty"`
}
type grant struct {
	GoalID    string `json:"goal_id"`
	Action    string `json:"action"`
	Principal string `json:"principal"`
	Expires   int64  `json:"expires"`
}
type broker struct {
	key            []byte
	requirePeerUID bool
}

type peerUIDContextKey struct{}

func peerUIDAllowed(peerUID, brokerUID int) bool { return peerUID == brokerUID }

func (b *broker) sign(g grant) string {
	raw, _ := json.Marshal(g)
	mac := hmac.New(sha256.New, b.key)
	mac.Write(raw)
	return base64.RawURLEncoding.EncodeToString(raw) + "." + hex.EncodeToString(mac.Sum(nil))
}
func (b *broker) valid(s string, r request) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	sig, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, b.key)
	mac.Write(raw)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return false
	}
	var g grant
	if json.Unmarshal(raw, &g) != nil {
		return false
	}
	return g.Principal == "owner" && g.GoalID == r.GoalID && g.Action == r.Action && g.Expires > time.Now().Unix()
}
func (b *broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reply := func(status int, v response) { w.WriteHeader(status); _ = json.NewEncoder(w).Encode(v) }
	if b.requirePeerUID {
		peerUID, ok := r.Context().Value(peerUIDContextKey{}).(int)
		if !ok || !peerUIDAllowed(peerUID, os.Getuid()) {
			reply(http.StatusForbidden, response{Code: "SOCKET_PEER_DENIED"})
			return
		}
	}
	if !hmac.Equal([]byte(r.Header.Get("Authorization")), []byte("Bearer "+string(b.key))) {
		reply(401, response{Code: "OWNER_AUTH_REQUIRED"})
		return
	}
	if r.Method != "POST" {
		reply(405, response{Code: "METHOD_NOT_ALLOWED"})
		return
	}
	if r.URL.Path != "/v1/grants" && r.URL.Path != "/v1/policy/evaluate" {
		reply(404, response{Code: "NOT_FOUND"})
		return
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	dec.DisallowUnknownFields()
	var req request
	if dec.Decode(&req) != nil || dec.Decode(new(any)) != io.EOF || len(req.GoalID) < 1 || len(req.GoalID) > 128 || len(req.Action) > 80 {
		reply(400, response{Code: "INVALID_REQUEST"})
		return
	}
	// Hard gates run before any utility computation. Spending stays zero.
	if strings.HasPrefix(req.Action, "payment.") {
		reply(200, response{Code: "BUDGET_UNSET"})
		return
	}
	if req.Privacy != "general" {
		reply(200, response{Code: "PRIVACY_SCOPE_DENIED"})
		return
	}
	if req.Action != "local.checksum" {
		reply(200, response{Code: "ACTION_DENIED"})
		return
	}
	if r.URL.Path == "/v1/grants" {
		reply(200, response{Allowed: true, Code: "OWNER_SCOPED_GRANT", Grant: b.sign(grant{req.GoalID, req.Action, "owner", time.Now().Add(time.Hour).Unix()})})
		return
	}
	if !b.valid(req.Grant, req) {
		reply(200, response{Code: "GRANT_INVALID_OR_EXPIRED"})
		return
	}
	reply(200, response{Allowed: true, Code: "SCOPED_ACTION_ALLOWED"})
}
func private(path string, dir bool) error {
	st, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if st.Mode()&os.ModeSymlink != 0 || st.IsDir() != dir || st.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("private owner-only path required")
	}
	if !dir && !st.Mode().IsRegular() {
		return fmt.Errorf("regular key file required")
	}
	if s, ok := st.Sys().(*syscall.Stat_t); !ok || int(s.Uid) != os.Getuid() {
		return fmt.Errorf("path owner mismatch")
	}
	return nil
}
func run() error {
	socket := flag.String("socket", "", "private Unix socket path")
	keyFile := flag.String("key-file", "", "owner key file (0600)")
	flag.Parse()
	if *socket == "" || *keyFile == "" {
		return fmt.Errorf("socket and key-file required")
	}
	return serve(*socket, *keyFile)
}

func serve(socket, keyFile string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("policy broker requires Linux Unix-socket peer credentials")
	}
	if err := private(filepath.Dir(socket), true); err != nil {
		return err
	}
	if err := private(keyFile, false); err != nil {
		return err
	}
	key, err := os.ReadFile(keyFile)
	if err != nil {
		return err
	}
	key = []byte(strings.TrimSpace(string(key)))
	if len(key) != 64 {
		return fmt.Errorf("key must be 32 random bytes encoded as hex")
	}
	if _, err = hex.DecodeString(string(key)); err != nil {
		return fmt.Errorf("key must be hex")
	}
	if _, err = os.Lstat(socket); !os.IsNotExist(err) {
		return fmt.Errorf("socket path exists or cannot be inspected; refusing replacement")
	}
	syscall.Umask(0077)
	l, err := net.Listen("unix", socket)
	if err != nil {
		return err
	}
	defer l.Close()
	if err := os.Chmod(socket, 0600); err != nil {
		return err
	}
	server := &http.Server{
		Handler: &broker{key: key, requirePeerUID: true},
		ConnContext: func(ctx context.Context, conn net.Conn) context.Context {
			uid, peerErr := unixPeerUID(conn)
			if peerErr != nil {
				return context.WithValue(ctx, peerUIDContextKey{}, -1)
			}
			return context.WithValue(ctx, peerUIDContextKey{}, uid)
		},
		ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 3 * time.Second,
		WriteTimeout: 3 * time.Second, IdleTimeout: 3 * time.Second, MaxHeaderBytes: 8192,
	}
	return server.Serve(l)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
