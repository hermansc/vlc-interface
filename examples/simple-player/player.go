package main

import (
	vlc "github.com/hermansc/vlc-interface"
	"log"
	"os"
)

func main() {
	// Get a new VLC player object.
	player := vlc.NewPlayer()

	// Add a standard module, specifying that we want to mux the output to http.
	player.AddModule("std", map[string]string{
		"access": "http",
		"mux":    "ts",
		"dst":    ":8080",
	})

	// Specify URL to media we want to play.
	cmd, err := player.Command("http://example.com/mystream.m3u")
	if err != nil {
		log.Fatalf("Could not get command for VLC (%s). Aborting.\n", err.Error())
	}

	// Get all stdout and sderr in our console.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run VLC, and wait for it to exit.
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
