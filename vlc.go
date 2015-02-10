package vlcinterface

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

type VLC struct {
	Modules map[string]Module
	Flags   []string
}

type Module struct {
	name       string
	properties map[string]string
}

func NewPlayer() *VLC {
	return &VLC{Modules: make(map[string]Module)}
}

func (player *VLC) AddModule(name string, properties map[string]string) {
	// The module object to insert.
	m := Module{
		name:       name,
		properties: properties,
	}

	// Insert it / overwrite the existing one.
	player.Modules[name] = m
}

func (player *VLC) GetSoutOpts() string {
	var opts []string
	for _, module := range player.Modules {
		var mopts []string
		for key, value := range module.properties {
			mopts = append(mopts, fmt.Sprintf("%s=%s", key, value))
		}
		opts = append(opts, fmt.Sprintf("%s{%s}", module.name, strings.Join(mopts, ",")))
	}
	return fmt.Sprintf("#%s", strings.Join(opts, ":"))
}

func (player *VLC) Command(input string) (*exec.Cmd, error) {
	// Get the VLC binary to use.
	binary, err := findVLCBinary()
	if err != nil {
		return nil, err
	}

	// Array, holding all flags to pass to VLC.
	var opts_array []string

	// See if we have defined any specific VLC flags (apart from the sout modules)
	if len(player.Flags) > 0 {
		opts_array = append(opts_array, fmt.Sprintf("%s", strings.Join(player.Flags, " ")))
	}

	// Add the sout-modules flags.
	opts_array = append(opts_array, fmt.Sprintf("--sout '%s'", player.GetSoutOpts()))

	// Mangle together the command
	options := strings.Join(opts_array, " ")
	raw_command := fmt.Sprintf("%s %s '%s'", binary, options, input)

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
