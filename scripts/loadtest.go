// loadtest simulates a small office team: N users (default 10) connecting at
// the same time and doing real work — listing/searching contacts, opening
// profiles, and creating/editing/deleting timeline notes — while the server
// runs with the exact SQLite pragmas the standalone binary ships with.
//
// It only uses the Go standard library, so it runs anywhere with Go installed.
//
// Usage:
//
//	# 1) Start the app locally (the same binary the .exe is made from):
//	#    API_RATE_MIN_INTERVAL=0s .../clientes-crm-linux
//	#    (0s disables the per-IP throttle so one machine can simulate many; the
//	#    LAN deployment never needs this because every user has their own IP.)
//
//	# 2) Run the load test:
//	cd scripts && go run ./loadtest.go -users 10 -steps 3
//
// The test registers its own throwaway user (strong random password), then has
// every simulated user log in and hit the API concurrently. It reports
// per-endpoint p50/p95/p99 latencies and fails loudly if the database reports
// "database is locked" or errors exceed ~1%.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const strongChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#%^&*"

func randomPassword(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte(strongChars[rand.IntN(len(strongChars))])
	}
	return b.String()
}

func main() {
	var (
		base     = flag.String("url", getenv("LOAD_TEST_URL", "http://localhost:7300"), "app base URL")
		users    = flag.Int("users", getenvInt("LOAD_TEST_USERS", 10), "simulated concurrent users")
		steps    = flag.Int("steps", getenvInt("LOAD_TEST_STEPS", 3), "interaction sessions per user")
		username = flag.String("username", getenv("LOAD_TEST_USERNAME", ""), "user to use (created if missing)")
		email    = flag.String("email", getenv("LOAD_TEST_EMAIL", ""), "email for the throwaway user")
		password = flag.String("password", getenv("LOAD_TEST_PASSWORD", ""), "password (generated if empty)")
	)
	flag.Parse()

	b := strings.TrimRight(*base, "/")
	user := *username
	mail := *email
	if user == "" {
		user = fmt.Sprintf("loadtest_%d", time.Now().UnixNano()%1000000)
	}
	if mail == "" {
		mail = fmt.Sprintf("%s@example.com", user)
	}
	created := *password == ""
	pass := *password
	if pass == "" {
		pass = randomPassword(28)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	// Ensure the user exists.
	if err := ensureUser(client, b, user, mail, pass); err != nil {
		fmt.Printf("(!) register step: %v\n", err)
		created = false
	}

	cookie, err := login(client, b, user, pass)
	if err != nil {
		fmt.Fprintf(os.Stderr, "login failed for user %q: %v\n", user, err)
		os.Exit(1)
	}

	if created {
		fmt.Printf("Registered throwaway user %q (password shown once):\n  %s\n", user, pass)
	} else {
		fmt.Printf("Using existing user %q\n", user)
	}
	fmt.Printf("Target: %s  |  simulated users: %d  |  sessions per user: %d\n\n", b, *users, *steps)

	agg := newAggregator()

	var wg sync.WaitGroup
	start := time.Now()
	for u := 0; u < *users; u++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			s := &session{client: client, base: b, cookie: cookie, userID: fmt.Sprintf("user%02d", id)}
			for step := 0; step < *steps; step++ {
				s.runStep(agg)
			}
		}(u)
	}
	wg.Wait()
	elapsed := time.Since(start)

	agg.printReport(*users, *steps, elapsed)

	if agg.locked() {
		fmt.Println("\nRESULT: FAIL — the database reported 'database is locked'. Concurrency settings need work.")
		os.Exit(2)
	}
	if agg.errorRate() > 0.01 {
		fmt.Printf("\nRESULT: FAIL — error rate %.2f%% (limit 1%%).\n", agg.errorRate()*100)
		os.Exit(2)
	}
	fmt.Println("\nRESULT: PASS — no locks and error rate under 1%.")
}

func ensureUser(client *http.Client, base, user, email, password string) error {
	body, _ := json.Marshal(map[string]string{
		"username": user, "email": email, "password": password, "language": "es",
	})
	resp, err := client.Post(base+"/api/v1/register", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("POST register: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	b, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("register returned %s: %s", resp.Status, strings.TrimSpace(string(b)))
}

func login(client *http.Client, base, user, password string) (string, error) {
	body, _ := json.Marshal(map[string]string{"identifier": user, "password": password})
	resp, err := client.Post(base+"/api/v1/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login returned %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	for _, h := range resp.Header.Values("Set-Cookie") {
		if strings.HasPrefix(h, "auth_token=") {
			return strings.TrimSpace(h[:strings.Index(h, ";")]), nil
		}
	}
	return "", fmt.Errorf("login response had no auth_token cookie")
}

type opStat struct {
	count   int
	totalNs int64
	lat     []time.Duration
	failed  int
}

type aggregator struct {
	mu     sync.Mutex
	ops    map[string]*opStat
	locks  int
	other  int
	status map[int]int
}

func newAggregator() *aggregator {
	return &aggregator{ops: map[string]*opStat{}, status: map[int]int{}}
}

func (a *aggregator) record(op string, d time.Duration, code int, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.ops[op]
	if s == nil {
		s = &opStat{}
		a.ops[op] = s
	}
	s.count++
	s.totalNs += int64(d)
	s.lat = append(s.lat, d)
	a.status[code]++
	if err != nil {
		s.failed++
		msg := err.Error()
		switch {
		case strings.Contains(msg, "database is locked"):
			a.locks++
		default:
			a.other++
		}
	}
}

func (a *aggregator) locked() bool { a.mu.Lock(); defer a.mu.Unlock(); return a.locks > 0 }

func (a *aggregator) errorRate() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	var total, failed int
	for _, s := range a.ops {
		total += s.count
		failed += s.failed
	}
	if total == 0 {
		return 0
	}
	return float64(failed) / float64(total)
}

func pct(lat []time.Duration, p float64) time.Duration {
	if len(lat) == 0 {
		return 0
	}
	idx := int(float64(len(lat)-1) * p / 100)
	return lat[idx]
}

func (a *aggregator) printReport(users, steps int, elapsed time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	var all []time.Duration
	var total int
	fmt.Println("Endpoint                                        count    p50      p95      p99      max      failed")
	names := make([]string, 0, len(a.ops))
	for name := range a.ops {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		s := a.ops[name]
		sort.Slice(s.lat, func(i, j int) bool { return s.lat[i] < s.lat[j] })
		all = append(all, s.lat...)
		total += s.count
		fmt.Printf("  %-44s %6d %9s %9s %9s %9s %6d\n",
			name, s.count, pct(s.lat, 50), pct(s.lat, 95), pct(s.lat, 99), s.lat[len(s.lat)-1], s.failed)
	}
	sort.Slice(all, func(i, j int) bool { return all[i] < all[j] })
	fmt.Printf("\nTotal requests:              %d (%.0f req/s over %s)\n", total, float64(total)/elapsed.Seconds(), elapsed)
	fmt.Printf("Overall latency:            p50 %s | p95 %s | p99 %s\n", pct(all, 50), pct(all, 95), pct(all, 99))
	fmt.Printf("Errors:                     %d lock(s), %d other\n", a.locks, a.other)
	fmt.Printf("HTTP status distribution:  %v\n", sortedStatus(a.status))
}

func sortedStatus(m map[int]int) string {
	var keys []int
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%d=%d", k, m[k]))
	}
	return strings.Join(parts, " ")
}

// session simulates one user's interactions with the app.

type session struct {
	client *http.Client
	base   string
	cookie string
	userID string
}

func (s *session) runStep(a *aggregator) {
	s.do(a, "list_contacts", "GET", "/api/v1/contacts?page=1&limit=20", "")
	s.do(a, "search_contacts", "GET", "/api/v1/contacts?limit=5&search=Ana", "")
	s.do(a, "global_timeline", "GET", "/api/v1/notes?limit=25", "")

	// Ensure there is at least one contact to drive the write endpoints, so the
	// test always exercises real database write concurrency even for a fresh user.
	id := s.firstContactID(a)
	if id == 0 {
		id = s.do(a, "create_contact", "POST", "/api/v1/contacts",
			fmt.Sprintf(`{"firstname":"Carga","lastname":"%s"}`, s.userID))
	}
	if id == 0 {
		return
	}
	s.do(a, "contact_detail", "GET", fmt.Sprintf("/api/v1/contacts/%d", id), "")
	s.do(a, "contact_notes", "GET", fmt.Sprintf("/api/v1/contacts/%d/notes", id), "")
	s.do(a, "contact_relationships", "GET", fmt.Sprintf("/api/v1/contacts/%d/relationships", id), "")

	// Create -> read back -> update -> delete a note (real write concurrency).
	noteID := s.do(a, "create_note", "POST", fmt.Sprintf("/api/v1/contacts/%d/notes", id),
		fmt.Sprintf(`{"title":"load test","content":"note from %s","date":"%s","contact_id":%d}`,
			s.userID, time.Now().UTC().Format("2006-01-02T15:04:05Z"), id))
	if noteID > 0 {
		s.do(a, "update_note", "PUT", fmt.Sprintf("/api/v1/notes/%d", noteID),
			fmt.Sprintf(`{"title":"load test","content":"note from %s (edited)","date":"%s","contact_id":%d}`,
				s.userID, time.Now().UTC().Format("2006-01-02T15:04:05Z"), id))
		s.do(a, "delete_note", "DELETE", fmt.Sprintf("/api/v1/notes/%d", noteID), "")
	}
}

// firstContactID lists contacts and returns the first id found (0 if none).
func (s *session) firstContactID(a *aggregator) uint64 {
	var buf bytes.Buffer
	status, body := s.request(a, "list_contacts", "GET", "/api/v1/contacts?page=1&limit=20", "", &buf)
	if status < 200 || status >= 300 {
		return 0
	}
	var payload struct {
		Contacts []struct {
			ID uint64 `json:"id"`
		} `json:"contacts"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || len(payload.Contacts) == 0 {
		return 0
	}
	return payload.Contacts[0].ID
}

func (s *session) do(a *aggregator, op, method, path, jsonBody string) uint64 {
	var buf bytes.Buffer
	status, _ := s.request(a, op, method, path, jsonBody, &buf)
	if status < 200 || status >= 300 || jsonBody == "" {
		return 0
	}
	// Created objects are returned nested, e.g. {"contact":{"id":...}} and
	// {"note":{"id":...}}; plain updates return them at the top level.
	var payload struct {
		ID      uint64 `json:"id"`
		Contact struct {
			ID uint64 `json:"id"`
		} `json:"contact"`
		Note struct {
			ID uint64 `json:"id"`
		} `json:"note"`
	}
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		return 0
	}
	switch {
	case payload.ID > 0:
		return payload.ID
	case payload.Contact.ID > 0:
		return payload.Contact.ID
	case payload.Note.ID > 0:
		return payload.Note.ID
	}
	return 0
}

func (s *session) request(a *aggregator, op, method, path, jsonBody string, out *bytes.Buffer) (int, []byte) {
	var body io.Reader
	if jsonBody != "" {
		body = strings.NewReader(jsonBody)
	}
	req, err := http.NewRequest(method, s.base+path, body)
	if err != nil {
		a.record(op, 0, 0, err)
		return 0, nil
	}
	req.Header.Set("Cookie", s.cookie)
	if jsonBody != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	start := time.Now()
	resp, err := s.client.Do(req)
	d := time.Since(start)
	if err != nil {
		a.record(op, d, 0, err)
		return 0, nil
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if out != nil {
		out.Write(data)
	}
	if getenv("LOAD_TEST_VERBOSE", "") != "" && resp.StatusCode >= 300 {
		fmt.Printf("  [dbg] %s %s -> %d: %s\n", method, path, resp.StatusCode, strings.TrimSpace(string(data))[:min(200, len(strings.TrimSpace(string(data))))])
	}
	a.record(op, d, resp.StatusCode, errorFor(op, resp.StatusCode))
	return resp.StatusCode, data
}

func errorFor(op string, code int) error {
	if code >= 200 && code < 300 {
		return nil
	}
	return fmt.Errorf("%s: HTTP %d", op, code)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
