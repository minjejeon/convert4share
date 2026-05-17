package concurrency

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestSem_AcquireRelease(t *testing.T) {
	s := New(2)
	ctx := context.Background()

	r1, err := s.Acquire(ctx)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	r2, err := s.Acquire(ctx)
	if err != nil {
		t.Fatalf("second acquire failed: %v", err)
	}

	// Third acquire should block; use a short-timeout ctx to verify.
	timeoutCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	if _, err := s.Acquire(timeoutCtx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded when sem is full, got %v", err)
	}

	// Release one and verify acquire succeeds.
	r1()
	r3, err := s.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire after release failed: %v", err)
	}
	r2()
	r3()
}

func TestSem_AcquireCancelledReturnsNoRelease(t *testing.T) {
	s := New(1)
	// Pre-fill.
	first, err := s.Acquire(context.Background())
	if err != nil {
		t.Fatalf("setup acquire failed: %v", err)
	}
	defer first()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	release, err := s.Acquire(ctx)
	if err == nil {
		t.Fatalf("expected context error, got nil")
	}
	if release != nil {
		t.Fatalf("expected nil release on cancellation, got non-nil")
	}
}

func TestSem_ResizeKeepsOutstandingReleasesValid(t *testing.T) {
	s := New(2)
	r1, err := s.Acquire(context.Background())
	if err != nil {
		t.Fatalf("acquire failed: %v", err)
	}

	// Resize while r1 is outstanding. The release must still work and
	// must not steal a slot from the new channel.
	s.Resize(4)
	if got := s.Cap(); got != 4 {
		t.Fatalf("Cap after Resize: want 4, got %d", got)
	}

	// Should not block / panic.
	r1()

	// New channel still has full capacity 4 available.
	ctx := context.Background()
	releases := make([]func(), 0, 4)
	for i := 0; i < 4; i++ {
		r, err := s.Acquire(ctx)
		if err != nil {
			t.Fatalf("acquire %d failed: %v", i, err)
		}
		releases = append(releases, r)
	}
	for _, r := range releases {
		r()
	}
}

func TestSem_ResizeBelowOneClampsToOne(t *testing.T) {
	s := New(3)
	s.Resize(0)
	if got := s.Cap(); got != 1 {
		t.Fatalf("Cap after Resize(0): want 1, got %d", got)
	}
}

func TestSem_NewBelowOneClampsToOne(t *testing.T) {
	s := New(0)
	if got := s.Cap(); got != 1 {
		t.Fatalf("Cap after New(0): want 1, got %d", got)
	}
}

func TestSem_Concurrent(t *testing.T) {
	const cap = 3
	const workers = 20
	s := New(cap)

	var (
		mu      sync.Mutex
		current int
		peak    int
	)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := s.Acquire(context.Background())
			if err != nil {
				t.Errorf("acquire: %v", err)
				return
			}
			mu.Lock()
			current++
			if current > peak {
				peak = current
			}
			mu.Unlock()

			time.Sleep(5 * time.Millisecond)

			mu.Lock()
			current--
			mu.Unlock()
			r()
		}()
	}
	wg.Wait()

	if peak > cap {
		t.Fatalf("peak concurrency %d exceeded capacity %d", peak, cap)
	}
}
