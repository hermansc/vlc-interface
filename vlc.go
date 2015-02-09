package main

import (
  "fmt"
  "strings"
  "os/exec"
  "os"
  "runtime"
  "errors"
  "regexp"
)

type VLC struct {
  Modules []Module
  Flags []string
}

type Module struct {
  name string
  values map[string]string
}

func NewPlayer() *VLC {
  return &VLC{}
}

func (player *VLC) AddModule(name string, values map[string]string) {
  // The module object to insert.
  m := Module{
    name: name,
    values: values,
  }

  // Check if module already exists
  for i, modules := range player.Modules {
    if name == modules.name {
      // It exists, so we rewrite it.
      player.Modules[i] = m
      return
    }
  }

  // Doesn't exist, so we append it.
  player.Modules = append(player.Modules, m)
}

func (player *VLC) SoutOpts() string {
  var opts []string
  for _, module := range player.Modules {
    var mopts []string
    for key, value := range module.values {
      mopts = append(mopts, fmt.Sprintf("%s=%s", key, value))
    }
    opts = append(opts,fmt.Sprintf("%s{%s}", module.name, strings.Join(mopts, ",")))
  }
  return fmt.Sprintf("#%s", strings.Join(opts, ":"))
}

func (player *VLC) Command(input string) (*exec.Cmd, error) {
  // Get the VLC binary to use.
  binary, err := findVLCBinary()
  if err != nil { return nil, err }

  // Array, holding all flags to pass to VLC.
  var opts_array []string

  // See if we have defined any specific VLC flags (apart from the sout modules)
  if len(player.Flags) > 0 {
    opts_array = append(opts_array, fmt.Sprintf("%s", strings.Join(player.Flags, " ")))
  }

  // Add the sout-modules flags.
  opts_array= append(opts_array, fmt.Sprintf("--sout '%s'", player.SoutOpts()))

  // Mangle together the command
  options := strings.Join(opts_array, " ")
  raw_command := fmt.Sprintf("%s %s '%s'" , binary, options, input)

  fmt.Println("Running: " + raw_command)

  // Run through bash, as this is by far the easiest way to run a long command in Go.
  return exec.Command("bash", "-c", raw_command), nil
}

func (player *VLC) AddFlag(flag string) {
  player.Flags = append(player.Flags, flag)
}

func (player *VLC) SetSoutOptions(opts string) {
  // Parse a '#transcode{heka=deka}:foo{bar=zoo}:std{loreum=ipsum}' string,
  // and add the modules as proper modules.
  sout_re := regexp.MustCompile(`(\w+){(.+?)}`)
  for _, match := range sout_re.FindAllStringSubmatch(opts, -1) {
    module := match[1]
    module_options := match[2]

    // Parse the inner values of the string, split by =
    mopts := make(map[string]string)
    for _, opt := range strings.Split(module_options, ",") {
      key_val := strings.Split(opt, "=")
      mopts[key_val[0]] = key_val[1]
    }

    player.AddModule(module, mopts)
  }
}

func findVLCBinary() (string, error) {
  // See if we have either cvlc or vlc on PATH.
  // Note that this will fail if a file named "vlc" or "cvlc" is in the exec directory.
  if _, err := os.Stat("vlc"); err == nil {
    return "vlc -I dummy", nil
  }
  if _, err := os.Stat("cvlc"); err == nil {
    return "cvlc", nil
  }

  // Alright, try some guesses based on OS.
  if runtime.GOOS == "darwin" {
    osx_default_path := "/Applications/VLC.app/Contents/MacOS/VLC"
    if _, err := os.Stat(osx_default_path); err == nil {
      return fmt.Sprintf("%s -I dummy", osx_default_path), nil
    }
  }
  return "", errors.New("Could not find VLC binary")
}

func main(){
  vlc := NewPlayer()

  vlc.AddModule("transcode", map[string]string{
    "vcodec": "mp4v",
  })

  vlc.AddModule("std", map[string]string{
    "access": "http",
    "mux": "ts",
    "dst": ":8080",
  })

  //vlc.AddFlag("--verbose 2")

  vlc.SetSoutOptions("#transcode{vcodec=h264}:std{access=http,mux=ts,dst=:8080}")

  // Get our command to run, based on input file.
  cmd, err := vlc.Command("http://nordond30a-f.akamaihd.net/i/wo/open/51/51f6610a9d94999eafb48058da88cc917ca24f22/44dafbc0-5ab2-4591-b47f-58d45015f3bf_,141,316,563,1266,2250,.mp4.csmil/index_3_av.m3u8?null=")

  if err != nil {
    fmt.Printf("Could not get command for VLC (%s). Aborting.\n", err.Error())
    os.Exit(1)
  }

  // Get all stdout and sderr in our console.
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr

  // Run VLC, and wait for it to exit.
  err = cmd.Run()

  // Run forever.
  for {
    err = cmd.Run()
    if err != nil {
      fmt.Printf("An error occured: %s\n", err.Error())
      os.Exit(1)
    }
  }

  fmt.Println("VLC is done. Exiting.")

}
