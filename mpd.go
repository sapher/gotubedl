package main

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type MPD struct {
	Periods []Period `xml:"Period"`
}

type Period struct {
	Duration       string          `xml:"duration,attr"`
	AdaptationSets []AdaptationSet `xml:"AdaptationSet"`
}

type AdaptationSet struct {
	Id                  int              `xml:"id,attr"`
	MimeType            string           `xml:"mimeType,attr"`
	SubsegmentAlignment bool             `xml:"subsegmentAlignment,attr"`
	Representations     []Representation `xml:"Representation"`
}

type Representation struct {
	Id                int     `xml:"id,attr"`
	Codecs            string  `xml:"codecs,attr"`
	AudioSamplingRate int     `xml:"audioSamplingRate,attr"`
	Bandwidth         int     `xml:"bandwidth,attr"`
	BaseURL           BaseURL `xml:"BaseURL"`
	Width             int     `xml:"width,attr"`
	Height            int     `xml:"height,attr"`
	FrameRate         int     `xml:"frameRate,attr"`
}

type BaseURL struct {
	ContentLength int    `xml:"contentLength,attr"`
	Value         string `xml:",chardata"`
}

func ParseMPDManifest(mpdUrl string, video *Video) error {

	// Decrypt signature
	regSig := regexp.MustCompile(`/s/([a-fA-F0-9\.]+)`)
	match := regSig.FindStringSubmatch(mpdUrl)
	if len(match) > 1 {
		sig := match[1]
		mpdUrl = strings.Replace(mpdUrl, match[0], "/signature/"+sig, 1)
	}

	// Get manifest
	mpdContent, err := downloadPage(mpdUrl)
	if err != nil {
		return err
	}

	// Parse MPD XML
	manifest := new(MPD)
	err = xml.Unmarshal([]byte(mpdContent), manifest)
	if err != nil {
		return err
	}

	for _, periods := range manifest.Periods {
		for _, apSets := range periods.AdaptationSets {
			for _, reps := range apSets.Representations {

				if reps.BaseURL.Value == "" {
					continue
				}

				// Get correct format from const formats list
				formatId := strconv.Itoa(reps.Id)
				newFormat := BaseFormats[formatId]

				// Resolution
				if reps.Width != 0 && reps.Height != 0 {
					newFormat.Resolution = fmt.Sprintf("%dx%d", reps.Width, reps.Height)
				}

				// Set data
				newFormat.FormatId = reps.Id
				newFormat.Url = reps.BaseURL.Value
				newFormat.Filesize = uint64(reps.BaseURL.ContentLength)
				newFormat.Tbr = reps.Bandwidth
				newFormat.Width = reps.Width
				newFormat.Height = reps.Height
				newFormat.Fps = reps.FrameRate
				newFormat.Asr = reps.AudioSamplingRate
				video.Formats[formatId] = newFormat
			}
		}
	}

	return nil
}
