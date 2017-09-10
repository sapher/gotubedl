package main

import "testing"

const videoId = "dQw4w9WgXcQ" // Never gonna Go you up

func TestExtractVideoId(t *testing.T) {

	// Embed
	embedUrl := "https://www.youtube.com/embed/" + videoId
	if extractVideoId(embedUrl) != "dQw4w9WgXcQ" {
		t.Errorf("id not extracted")
	}

	// Watch
	watchUrl := "https://www.youtube.com/watch?v=" + videoId
	if extractVideoId(watchUrl) != "dQw4w9WgXcQ" {
		t.Errorf("id not extracted")
	}
}
