#! /usr/bin/nu

let CHECK = $env.CHECK? | default "false" | into bool
let CHECK_URL = $env.CHECK_URL? | default "http://icecast:8000/status-json.xsl"
let CHECK_INTERVAL = $env.CHECK_INTERVAL? | default "60" | append "sec" | str join "" | into duration

let STREAM_URL = $env.STREAM_URL? | default "https://stream.radio-t.com"
let TG_SERVER = $env.TG_SERVER? | default "dc4-1.rtmp.t.me"
let TG_KEY = $env.TG_KEY
let DEST_URL = $"rtmps://($TG_SERVER)/s/($TG_KEY)"

let DEBUG = $env.DEBUG? | default "false" | into bool

use std log

def log_ [log_string: string] {
  let now = date now | format date "[%Y-%m-%d %H:%M:%S]"
  print $"($now) ($log_string)"
}

def check [] {
  mut status = false
  while not $status {
    log debug $"Check: ($CHECK_URL)"
    let data = http get $CHECK_URL
    if ($data.icestats | columns | find source | is-not-empty)  {
      log debug "SUCCESS"
      $status = true
    } else {
      log debug "FAIL"
      log debug $"Next check in ($CHECK_INTERVAL)"
      sleep $CHECK_INTERVAL
    }
  }
}

def work [] {
  log info "!!!"
  log info "!!! Start retranslation."
  log info "!!!"
  log info $"!!! Source: ($STREAM_URL)"
  log info $"!!! Destination: ($DEST_URL)"
  log info "!!!"
  
  mut run_opts = []
  $run_opts = ($run_opts | append ['-v','verbose'])
  $run_opts = ($run_opts | append '-nostdin')
  $run_opts = ($run_opts | append '-nostats')
  $run_opts = ($run_opts | append '-hide_banner')
  $run_opts = ($run_opts | append ['-loop', '1'])
  $run_opts = ($run_opts | append ['-i', 'logo-dark.png'])
  $run_opts = ($run_opts | append ['-i', $STREAM_URL])
  $run_opts = ($run_opts | append ['-c:v', 'libx264'])
  $run_opts = ($run_opts | append ['-tune', 'stillimage'])
  $run_opts = ($run_opts | append ['-pix_fmt', 'yuv420p'])
  $run_opts = ($run_opts | append ['-c:a', 'aac'])
  $run_opts = ($run_opts | append ['-b:a','128k'])
  $run_opts = ($run_opts | append ['-ac', '1'])
  $run_opts = ($run_opts | append ['-ar', '44100'])
  $run_opts = ($run_opts | append ['-f','flv'])
  $run_opts = ($run_opts | append ['-rtmp_live','-1'])
  $run_opts = ($run_opts | append $DEST_URL)

  log debug $"Run options: ($run_opts)"
  
  run-external "/usr/bin/ffmpeg" ...($run_opts)
} 

def main [] {
  if $DEBUG { $env.NU_LOG_LEVEL = "DEBUG" }
  while true {
    if $CHECK { check }
    work
  }
}
