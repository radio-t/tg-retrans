package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/go-pkgz/lgr"
	"github.com/umputun/go-flags"
)

type Options struct {
	Check         bool          `long:"check" env:"CHECK" description:"Enable status check"`
	CheckURL      string        `long:"check-url" env:"CHECK_URL" default:"http://icecast:8000/status-json.xsl" description:"URL to check the stream status"`
	CheckInterval time.Duration `long:"check-interval" env:"CHECK_INTERVAL" default:"60s" description:"Interval for status checks (in seconds)"`
	StreamURL     string        `long:"stream-url" env:"STREAM_URL" default:"https://stream.radio-t.com" description:"Source stream URL"`
	TGServer      string        `long:"tg-server" env:"TG_SERVER" default:"dc4-1.rtmp.t.me" description:"Telegram server"`
	TGKey         string        `long:"tg-key" env:"TG_KEY" required:"true" description:"Telegram stream key"`
	Debug         bool          `long:"debug" env:"DEBUG" description:"Enable debug mode"`
}

var (
	opts Options
	log  = lgr.New(lgr.Msec)
)

func checkStatus() {
	status := false
	for !status {
		log.Logf("[DEBUG] Checking: %s", opts.CheckURL)
		resp, err := http.Get(opts.CheckURL)
		if err != nil {
			log.Logf("[DEBUG] FAIL: %v", err)
			time.Sleep(opts.CheckInterval)
			continue
		}
		defer resp.Body.Close()

		var data map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
			if icestats, ok := data["icestats"].(map[string]interface{}); ok {
				if sources, ok := icestats["source"]; ok && sources != nil {
					log.Logf("[DEBUG] SUCCESS")
					status = true
					continue
				}
			}
		}

		log.Logf("[DEBUG] FAIL: invalid data")
		log.Logf("[DEBUG] Next check in %s", opts.CheckInterval)
		time.Sleep(opts.CheckInterval)
	}
}

func startWork() {
	destURL := fmt.Sprintf("rtmps://%s/s/%s", opts.TGServer, opts.TGKey)

	log.Logf("[INFO] Start retranslation")
	log.Logf("[INFO] Source: %s", opts.StreamURL)
	log.Logf("[INFO] Destination: %s", destURL)

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

	log.Logf("[DEBUG] Run options: %v", runOpts)

	cmd := exec.Command("/usr/bin/ffmpeg", runOpts...)
	if opts.Debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
	} else {
		cmd.Stdout = nil
		cmd.Stderr = nil
	}

	if err := cmd.Run(); err != nil {
		log.Logf("[ERROR] Failed to run ffmpeg: %v", err)
	}

	log.Logf("[INFO] End retranslation")
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		log.Logf("[ERROR] Can't read params")
		os.Exit(1)
	}

	if opts.Debug {
		log = lgr.New(lgr.Msec, lgr.Debug, lgr.CallerFunc)
		log.Logf("[DEBUG] Debug mode enabled")
		log.Logf("Options: %+v\n", opts)
	}

	for {
		if opts.Check {
			checkStatus()
		}
		startWork()
	}
}
