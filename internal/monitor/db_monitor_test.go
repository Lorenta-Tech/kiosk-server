package monitor

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakePinger struct {
	mu      sync.Mutex
	err     error
	calls   int
	lastErr error
}

func (f *fakePinger) PingContext(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.lastErr = f.err
	return f.err
}

func (f *fakePinger) setErr(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.err = err
}

type sentMail struct {
	to      []string
	subject string
	body    string
}

type fakeMailer struct {
	mu       sync.Mutex
	sent     []sentMail
	failWith error
}

func (f *fakeMailer) Send(to []string, subject string, body string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failWith != nil {
		return f.failWith
	}
	f.sent = append(f.sent, sentMail{to: to, subject: subject, body: body})
	return nil
}

func (f *fakeMailer) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sent)
}

func (f *fakeMailer) last() sentMail {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.sent) == 0 {
		return sentMail{}
	}
	return f.sent[len(f.sent)-1]
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func testConfig() DBConfig {
	return DBConfig{
		Enabled:           true,
		Interval:          10 * time.Millisecond,
		Timeout:           time.Second,
		FailureThreshold:  2,
		RepeatInterval:    0,
		Recipients:        []string{"suhasdeveloper07@gmail.com"},
		NotifyOnRecovered: true,
	}
}

func newTestMonitor(p *fakePinger, m *fakeMailer, config DBConfig) *DBMonitor {
	return NewDBMonitorWithConfig(p, m, testLogger(), config)
}

func TestHealthyDatabaseSendsNoEmail(t *testing.T) {
	p := &fakePinger{}
	m := &fakeMailer{}
	mon := newTestMonitor(p, m, testConfig())

	for i := 0; i < 5; i++ {
		mon.probe(context.Background())
	}

	if m.count() != 0 {
		t.Fatalf("expected no emails while database is healthy, got %d", m.count())
	}
}

func TestSingleFailureBelowThresholdSendsNoEmail(t *testing.T) {
	p := &fakePinger{err: errors.New("connection refused")}
	m := &fakeMailer{}
	mon := newTestMonitor(p, m, testConfig())

	mon.probe(context.Background())

	if m.count() != 0 {
		t.Fatalf("expected no email below failure threshold, got %d", m.count())
	}
}

func TestDownAlertSentOnceAtThreshold(t *testing.T) {
	p := &fakePinger{err: errors.New("connection refused")}
	m := &fakeMailer{}
	mon := newTestMonitor(p, m, testConfig())

	mon.probe(context.Background())
	mon.probe(context.Background())

	if m.count() != 1 {
		t.Fatalf("expected exactly 1 down email, got %d", m.count())
	}

	sent := m.last()
	if !strings.Contains(sent.subject, "Database Down") {
		t.Errorf("expected down subject, got %q", sent.subject)
	}
	if !strings.Contains(sent.body, "connection refused") {
		t.Errorf("expected error detail in body, got %q", sent.body)
	}
	if len(sent.to) != 1 || sent.to[0] != "suhasdeveloper07@gmail.com" {
		t.Errorf("unexpected recipients: %v", sent.to)
	}

	for i := 0; i < 10; i++ {
		mon.probe(context.Background())
	}
	if m.count() != 1 {
		t.Fatalf("expected no repeat emails with RepeatInterval=0, got %d", m.count())
	}
}

func TestBriefFlappingNeverReachesThreshold(t *testing.T) {
	p := &fakePinger{err: errors.New("connection refused")}
	m := &fakeMailer{}
	config := testConfig()
	config.NotifyOnRecovered = false
	mon := newTestMonitor(p, m, config)

	for i := 0; i < 5; i++ {
		mon.probe(context.Background())
		p.setErr(nil)
		mon.probe(context.Background())
	}

	if m.count() != 0 {
		t.Fatalf("expected no emails for single-probe blips, got %d", m.count())
	}
}

func TestSeparateOutagesEachAlertOnce(t *testing.T) {
	p := &fakePinger{err: errors.New("connection refused")}
	m := &fakeMailer{}
	config := testConfig()
	config.NotifyOnRecovered = false
	mon := newTestMonitor(p, m, config)

	for i := 0; i < 3; i++ {
		p.setErr(errors.New("connection refused"))
		mon.probe(context.Background())
		mon.probe(context.Background())
		mon.probe(context.Background())

		p.setErr(nil)
		mon.probe(context.Background())
	}

	if m.count() != 3 {
		t.Fatalf("expected one email per distinct outage, got %d", m.count())
	}
}

func TestRecoveryEmailSentOnce(t *testing.T) {
	p := &fakePinger{err: errors.New("connection refused")}
	m := &fakeMailer{}
	mon := newTestMonitor(p, m, testConfig())

	mon.probe(context.Background())
	mon.probe(context.Background())

	p.setErr(nil)
	mon.probe(context.Background())

	if m.count() != 2 {
		t.Fatalf("expected down + recovered emails, got %d", m.count())
	}
	if !strings.Contains(m.last().subject, "Database Recovered") {
		t.Errorf("expected recovered subject, got %q", m.last().subject)
	}

	mon.probe(context.Background())
	mon.probe(context.Background())

	if m.count() != 2 {
		t.Fatalf("expected no further emails after recovery, got %d", m.count())
	}
}

func TestRecoveryEmailSkippedWhenDisabled(t *testing.T) {
	p := &fakePinger{err: errors.New("connection refused")}
	m := &fakeMailer{}
	config := testConfig()
	config.NotifyOnRecovered = false
	mon := newTestMonitor(p, m, config)

	mon.probe(context.Background())
	mon.probe(context.Background())
	p.setErr(nil)
	mon.probe(context.Background())

	if m.count() != 1 {
		t.Fatalf("expected only down email, got %d", m.count())
	}
}

func TestRepeatAlertWhileStillDown(t *testing.T) {
	p := &fakePinger{err: errors.New("connection refused")}
	m := &fakeMailer{}
	config := testConfig()
	config.RepeatInterval = 20 * time.Millisecond
	mon := newTestMonitor(p, m, config)

	mon.probe(context.Background())
	mon.probe(context.Background())
	mon.probe(context.Background())

	if m.count() != 1 {
		t.Fatalf("expected 1 email before repeat interval elapses, got %d", m.count())
	}

	time.Sleep(25 * time.Millisecond)
	mon.probe(context.Background())

	if m.count() != 2 {
		t.Fatalf("expected repeat email after interval, got %d", m.count())
	}
}

func TestDownAlertRetriedAfterMailFailure(t *testing.T) {
	p := &fakePinger{err: errors.New("connection refused")}
	m := &fakeMailer{failWith: errors.New("resend unreachable")}
	config := testConfig()
	config.RepeatInterval = time.Millisecond
	mon := newTestMonitor(p, m, config)

	mon.probe(context.Background())
	mon.probe(context.Background())
	if m.count() != 0 {
		t.Fatalf("expected no recorded sends while mailer fails, got %d", m.count())
	}

	time.Sleep(2 * time.Millisecond)
	mon.probe(context.Background())

	if m.count() != 0 {
		t.Fatalf("expected still no sends while mailer keeps failing, got %d", m.count())
	}
}

func TestStartRespectsDisabledConfig(t *testing.T) {
	p := &fakePinger{}
	m := &fakeMailer{}
	config := testConfig()
	config.Enabled = false
	mon := newTestMonitor(p, m, config)

	mon.Start(context.Background())

	time.Sleep(30 * time.Millisecond)

	p.mu.Lock()
	calls := p.calls
	p.mu.Unlock()

	if calls != 0 {
		t.Fatalf("expected no probes when monitor disabled, got %d", calls)
	}
}

func TestStartRunsLoopAndStopsOnContextCancel(t *testing.T) {
	p := &fakePinger{}
	m := &fakeMailer{}
	mon := newTestMonitor(p, m, testConfig())

	ctx, cancel := context.WithCancel(context.Background())
	mon.Start(ctx)

	time.Sleep(60 * time.Millisecond)
	cancel()
	time.Sleep(30 * time.Millisecond)

	p.mu.Lock()
	first := p.calls
	p.mu.Unlock()
	time.Sleep(40 * time.Millisecond)
	p.mu.Lock()
	second := p.calls
	p.mu.Unlock()

	if first == 0 {
		t.Fatal("expected monitor loop to probe the database")
	}
	if second != first {
		t.Fatalf("expected loop to stop after cancel, calls went from %d to %d", first, second)
	}
}

func TestPingRespectsTimeout(t *testing.T) {
	blocking := &blockingPinger{}
	m := &fakeMailer{}
	config := testConfig()
	config.Timeout = 20 * time.Millisecond
	mon := NewDBMonitorWithConfig(blocking, m, testLogger(), config)

	start := time.Now()
	err := mon.ping(context.Background())
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed > time.Second {
		t.Fatalf("ping exceeded timeout budget: %s", elapsed)
	}
}

type blockingPinger struct{}

func (b *blockingPinger) PingContext(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func TestHumanDuration(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{30 * time.Second, "30 seconds"},
		{5 * time.Minute, "5 minutes"},
		{2*time.Hour + 30*time.Minute, "2 hours 30 minutes"},
	}

	for _, c := range cases {
		if got := humanDuration(c.in); got != c.want {
			t.Errorf("humanDuration(%s) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("short", 10); got != "short" {
		t.Errorf("expected unchanged string, got %q", got)
	}
	got := truncate(strings.Repeat("a", 50), 10)
	if len(got) != 13 || !strings.HasSuffix(got, "...") {
		t.Errorf("expected truncated string with ellipsis, got %q (len %d)", got, len(got))
	}
}
