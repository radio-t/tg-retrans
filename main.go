package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-pkgz/lgr"
	"github.com/umputun/go-flags"
)

var opts struct {
	CheckURL      string        `long:"check-url" env:"CHECK_URL" default:"http://icecast:8000/status-json.xsl" description:"URL to check the stream status"`
	CheckInterval time.Duration `long:"check-interval" env:"CHECK_INTERVAL" default:"60s" description:"Interval for status checks"`
	CheckTimeout  time.Duration `long:"check-timeout" env:"CHECK_TIMEOUT" default:"5s" description:"Timeout for status check"`

	StreamURL  string `long:"stream-url" env:"STREAM_URL" default:"https://stream.radio-t.com" description:"Source stream URL"`
	FfmpegPath string `long:"ffmpeg-path" env:"FFMPEG_PATH" default:"/usr/bin/ffmpeg" description:"Path to ffmpeg binary"`
	SkipCheck  bool   `long:"skip-check" env:"SKIP_CHECK" description:"Disable status check"`
	TGServer   string `long:"tg-server" env:"TG_SERVER" default:"dc4-1.rtmp.t.me" description:"Telegram server"`
	TGKey      string `long:"tg-key" env:"TG_KEY" required:"true" description:"Telegram stream key"`
	Debug      bool   `long:"debug" env:"DEBUG" description:"Enable debug mode"`
}

var revision = "unknown"

func main() {
	fmt.Printf("tg-retrans, %s\n", revision)
	if _, err := flags.Parse(&opts); err != nil {
		log.Printf("[ERROR] failed to parse flags: %v", err)
		os.Exit(2)
	}
	setupLog(opts.Debug)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatalf("[ERROR] run failed: %v", err)
	}
	log.Printf("[INFO] completed")
}

// run is the main function that starts the retranslation process
// it is one-shot if SkipCheck is set, otherwise it runs in a loop with a check interval
func run(ctx context.Context) error {
	if opts.SkipCheck {
		if err := startRetrans(ctx); err != nil {
			return fmt.Errorf("failed to start retranslation: %w", err)
		}
		if !checkStreamStatus(ctx) {
			return fmt.Errorf("stream is not available")
		}
		return nil
	}

	for {
		// cancel the context if the parent is done
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if checkStreamStatus(ctx) {
			log.Print("[INFO] Stream is available, start retranslation")
			if err := startRetrans(ctx); err != nil {
				log.Printf("[WARN] failed to start retranslation: %v", err)
			}
		} else {
			log.Printf("[DEBUG] Not streaming, next check in %v", opts.CheckInterval)
		}
		time.Sleep(opts.CheckInterval)
	}
}

// checkStreamStatus checks if the stream is available
func checkStreamStatus(ctx context.Context) bool {
	log.Printf("[DEBUG] Checking stream with %s", opts.CheckURL)
	client := http.Client{Timeout: opts.CheckTimeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, opts.CheckURL, http.NoBody)
	if err != nil {
		log.Printf("[WARN] Can't make request to %s: %v", opts.CheckURL, err)
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[WARN] Can't get response from %s: %v", opts.CheckURL, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[DEBUG] Invalid status code for %s: %d", opts.CheckURL, resp.StatusCode)
		return false
	}

	data := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Printf("[WARN] Failed to decode response: %v", err)
		return false
	}

	icestats, ok := data["icestats"].(map[string]any)
	if !ok {
		log.Printf("[WARN] Missing icestats key in response")
		return false
	}

	if sources, ok := icestats["source"]; !ok || sources == nil {
		log.Printf("[WARN] Missing or empty source in icestats response")
		return false
	}

	log.Printf("[DEBUG] Status check passed")
	return true
}

func startRetrans(ctx context.Context) error {
	// spawnFFmpeg creates and runs ffmpeg process
	spawnFFmpeg := func(destURL string) error {
		runOpts := []string{
			"-v", "verbose",
			"-nostdin",
			"-nostats",
			"-hide_banner",
			"-loop", "1",
			"-i", "logo-dark.png",
			"-i", opts.StreamURL,
			"-c:v", "libx264",
			"-tune", "stillimage",
			"-pix_fmt", "yuv420p",
			"-c:a", "aac",
			"-b:a", "128k",
			"-ac", "1",
			"-ar", "44100",
			"-f", "flv",
			"-rtmp_live", "-1",
			destURL,
		}

		log.Printf("[DEBUG] Run options: %v", runOpts)
		cmd := exec.CommandContext(ctx, opts.FfmpegPath, runOpts...)

		if opts.Debug {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout
		}

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run ffmpeg: %w", err)
		}
		return nil
	}

	destURL := fmt.Sprintf("rtmps://%s/s/%s", opts.TGServer, opts.TGKey)
	log.Printf("[INFO] Start retranslation from %s to %s", opts.StreamURL, destURL)
	start := time.Now()
	if err := spawnFFmpeg(destURL); err != nil {
		return fmt.Errorf("failed to start retranslation: %w", err)
	}

	log.Printf("[INFO] End retranslation in %v", time.Since(start))
	return nil
}

func setupLog(dbg bool) {
	logOpts := []lgr.Option{lgr.Msec, lgr.LevelBraces, lgr.StackTraceOnError}
	if dbg {
		logOpts = []lgr.Option{lgr.Debug, lgr.CallerFile, lgr.CallerFunc, lgr.Msec, lgr.LevelBraces, lgr.StackTraceOnError}
	}
	lgr.SetupStdLogger(logOpts...)
}
