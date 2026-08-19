package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunStopsWhileWaitingForTheNextCheck(t *testing.T) {
	checked := make(chan struct{}, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		select {
		case checked <- struct{}{}:
		default:
		}
		w.WriteHeader(http.StatusNotFound) // not streaming, run goes straight to the wait
	}))
	defer ts.Close()

	saved := opts
	opts.CheckURL = ts.URL
	opts.CheckInterval = time.Hour
	opts.CheckTimeout = time.Second
	opts.SkipCheck = false

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	t.Cleanup(func() {
		cancel()
		waitDone(t, done)
		opts = saved
	})

	go func() {
		done <- run(ctx)
		close(done)
	}()

	select {
	case <-checked:
	case <-time.After(5 * time.Second):
		t.Fatal("run did not make the status check")
	}

	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("run did not return within 5s of cancellation, check interval is %v", opts.CheckInterval)
	}
}

func TestRunReportsKilledRetranslationAsCancellation(t *testing.T) {
	dir := t.TempDir()
	ready := filepath.Join(dir, "ready")
	ffmpeg := filepath.Join(dir, "ffmpeg")
	script := "#!/bin/sh\ntouch '" + ready + "'\nexec sleep 60\n" // exec, so the killed shell takes sleep with it
	if err := os.WriteFile(ffmpeg, []byte(script), 0o700); err != nil {
		t.Fatalf("can't write the fake ffmpeg: %v", err)
	}

	saved := opts
	opts.SkipCheck = true
	opts.FfmpegPath = ffmpeg

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	t.Cleanup(func() {
		cancel()
		waitDone(t, done)
		opts = saved
	})

	go func() {
		done <- run(ctx)
		close(done)
	}()

	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the fake ffmpeg did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}

	cancel()

	select {
	case err := <-done:
		// the killed process reports "signal: killed", main must not treat it as a failure
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("run did not return after the retranslation was cancelled")
	}
}

// waitDone waits for run to return, so the options it reads can be restored safely
func waitDone(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("run is still going after cancellation")
	}
}
