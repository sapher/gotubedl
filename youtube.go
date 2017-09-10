package main

import (
	"regexp"
	"net/url"
)

// Extract video id from youtube's url
func ExtractVideoId(videoUrl string) string {
	r := regexp.MustCompile(`(?:youtube\.com\/(?:[^\/]+\/.+\/|(?:v|e(?:mbed)?)\/|.*[?&]v=)|youtu\.be\/)([^"&?\/ ]{11})`)
	match := r.FindStringSubmatch(videoUrl)
	return match[1]
}

// Get video info from youtube
func GetVideoInfo(videoId string) (url.Values, error) {

	// QueryString
	query := url.Values{
		"video_id": { videoId },
		"el": { "info" },
		"ps": { "default" },
		"eurl": { "" },
		"gl": { "US" },
		"hl": { "en" },
	}

	infoVideoUrl := "http://www.youtube.com/get_video_info?" + query.Encode()

	// Download page
	rawContent, err := downloadPage(infoVideoUrl)
	if err != nil {
		return nil, err
	}

	// get_video_info return a huge querystring
	videoInfo, err := url.ParseQuery(rawContent)
	if err != nil {
		return nil, err
	}

	return videoInfo, nil
}