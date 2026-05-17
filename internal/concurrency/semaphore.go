// Package concurrency provides small concurrency primitives used by
// Convert4Share. The Sem helper wraps the existing buffered-channel
// semaphore pattern so callers can acquire with context cancellation and
// resize the capacity safely at runtime.
package concurrency

import (
	"context"
	"sync"
)

// Sem is a counting semaphore backed by a buffered channel. Acquire is
// context-aware: if the context is cancelled before a slot is taken, the
// caller receives the context's error and the returned release func is
// nil. Each successful Acquire returns a release closure bound to the
// channel slot it took, so Resize is safe even while jobs are in flight
// (the release closure drains the channel the slot was taken from).
type Sem struct {
	mu sync.Mutex
	ch chan struct{}
}

// New creates a Sem with the given capacity. Capacities below 1 are
// clamped to 1 so callers can never accidentally produce a zero-capacity
// (i.e. permanently blocking) semaphore.
func New(capacity int) *Sem {
	if capacity < 1 {
		capacity = 1
	}
	return &Sem{ch: make(chan struct{}, capacity)}
}

// Acquire takes a slot, blocking until one is available or ctx is done.
// On success it returns a release closure (which must be called exactly
// once, normally via defer) and a nil error. On context cancellation it
// returns (nil, ctx.Err()) — the caller MUST NOT invoke any release in
// that case (matching the BUG-001 contract).
func (s *Sem) Acquire(ctx context.Context) (release func(), err error) {
	s.mu.Lock()
	ch := s.ch
	s.mu.Unlock()

	select {
	case ch <- struct{}{}:
		// Bind release to the channel snapshot so a subsequent Resize
		// that swaps s.ch does not cause this Release to drain from a
		// different (possibly empty) channel.
		return func() { <-ch }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Resize replaces the underlying channel with a new one of the requested
// capacity. Outstanding release closures continue to drain the old
// channel, which is then garbage-collected once they are all done.
// Capacities below 1 are clamped to 1.
func (s *Sem) Resize(newCapacity int) {
	if newCapacity < 1 {
		newCapacity = 1
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if cap(s.ch) == newCapacity {
		return
	}
	s.ch = make(chan struct{}, newCapacity)
}

// Cap returns the current capacity.
func (s *Sem) Cap() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cap(s.ch)
}
