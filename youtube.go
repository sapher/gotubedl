package main

import (
	"net/url"
	"regexp"
)

type Playlist struct {
	Title  string  `json:"title"`
	Videos []Video `json:"videos"`
}

type Video struct {
	VideoId   string  `json:"video_id"`
	Title     string  `json:"title"`
	Author    string  `json:"title"`
	Duration  string  `json:"duration"`
	Formats   Formats `json:"formats"`
	ViewCount int     `json:"view_count"`
}

// Extract video id from youtube's url
func ExtractVideoId(videoUrl string) string {
	r := regexp.MustCompile(`(?:youtube\.com\/(?:[^\/]+\/.+\/|(?:v|e(?:mbed)?)\/|.*[?&]v=)|youtu\.be\/)([^"&?\/ ]{11})`)
	match := r.FindStringSubmatch(videoUrl)
	return match[1]
}

// Get playlist info from youtube
func GetPlaylistInfo(playlistId string) (playlist Playlist) {

	playlistUrl := "https://www.youtube.com/playlist?list=" + playlistId

	playlistPage, err := downloadPage(playlistUrl)
	if err != nil {
		return
	}

	// Video Ids
	regVideos := regexp.MustCompile(`href="\s*/watch\?v=(?P<id>[0-9A-Za-z_-]{11})&amp;[^"]*?index=(?P<index>\d+)(?:[^>]+>(?P<title>[^<]+))?`)
	matchs := regVideos.FindAllStringSubmatch(playlistPage, -1)

	// Don't keep duplicates
	var videoIds []string
	for _, match := range matchs {
		found := false
		for _, videoId := range videoIds {
			if videoId == match[1] {
				found = true
			}
		}

		if !found {
			videoIds = append(videoIds, match[1])
		}
	}

	return
}

// Get video info from youtube
func GetVideoInfo(videoId string) (url.Values, error) {

	// QueryString
	query := url.Values{
		"video_id": {videoId},
		"el":       {"info"},
		"ps":       {"default"},
		"eurl":     {""},
		"gl":       {"US"},
		"hl":       {"en"},
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
