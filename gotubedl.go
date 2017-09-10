package main

import (
	"os"
	"fmt"
	"net/http"
	"io/ioutil"
	"net/url"
	"log"
	"strings"
	"strconv"
	"github.com/cavaliercoder/grab"
	"github.com/jessevdk/go-flags"
	"time"
)

// Video information
type VideoResult struct {
	VideoId string `json:"video_id"`
	Title string `json:"title"`
	Author string `json:"title"`
	Duration string `json:"duration"`
	Formats Formats `json:"formats"`
	ViewCount int `json:"view_count"`
}

func itoa(a string) (i int) {
	i, err := strconv.Atoi(a)
	if err != nil {
		i = 0
	}
	return
}

func getFormatSpecs(videoInfo url.Values, videoResult *VideoResult) {

	// parse formats
	/**var fmtList = videoInfo.Get("fmt_list")
	for _, _fmt := range strings.Split(fmtList, ",") {
		specs := strings.Split(_fmt, "/")
		formatId := specs[0]
		size := strings.Split(specs[1], "x")
		videoResult.formats[formatId] = Format{
			FormatId: itoa(formatId),
			Resolution: specs[1],
			Width: itoa(size[0]),
			Height: itoa(size[1]),
		}
	}**/

	var fmtStreamMap = videoInfo.Get("url_encoded_fmt_stream_map")
	var fmtAdaptive = videoInfo.Get("adaptive_fmts")

	// Put all formats in the same string
	fmtStrs := strings.Join([]string{ fmtStreamMap, fmtAdaptive }, ",")

	// Iterate through all format strings
	for _, fmtStr := range strings.Split(fmtStrs, ",") {

		urlData, _ := url.ParseQuery(fmtStr)

		formatId := urlData.Get("itag") // format_id
		videoUrl := urlData.Get("url")

		// Nothing to do here, abort
		if formatId == "" || videoUrl == "" {
			continue
		}

		// Handle signature
		/**encSig := urlData.Get("s")
		if encSig != "" {
			fmt.Println("s", encSig)
		}

		sig := urlData.Get("sig")
		if sig != "" {
			videoUrl += "&signature=" + sig
		} /**else encSig != "" {

		}

		ratebypass := urlData.Get("ratebypass")
		if ratebypass == "" {
			videoUrl += "&ratebypass=yes"
		}**/

		// Extract int from urlData
		extractInt := func(key string) int {
			return itoa(urlData.Get(key))
		}

		// Format
		newFormat := BaseFormats[formatId]
		newFormat.FormatId = itoa(formatId)
		newFormat.Url = videoUrl
		newFormat.Tbr = extractInt("bitrate")
		newFormat.Filesize = uint64(extractInt("clen"))
		newFormat.Fps = extractInt("fps")

		// Size
		size := urlData.Get("size")
		if size != "" {
			_size := strings.Split(size, "x")
			newFormat.Width = itoa(_size[0])
			newFormat.Height = itoa(_size[1])
			newFormat.Resolution = size
		}

		// Quality
		quality := urlData.Get("quality_label")
		if quality == "" {
			quality = urlData.Get("quality")
		}
		newFormat.Format_note = quality

		// Type
		/**_type := urlData.Get("type")
		fmt.Println(_type)
		if _type != "" {
			typeSplit := strings.Split(_type, ";")
			kindExt := strings.Split(typeSplit[0], "/")
			if len(kindExt) == 2 {
				kind := kindExt[0]

				switch kind {
				case "audio":
					break;
				case "video":
					break;
				}
			}
		}**/

		// Set data
		videoResult.Formats[formatId] = newFormat

		// Generals
		videoResult.ViewCount = extractInt("view_count")
	}
}

// Download a page from internet regardless of its content
// Return a string representing the body
func downloadPage(url string) (string, error) {

	// Download url content
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}

	// Gives up in status != 200
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf(res.Status)
	}

	// Get data body
	body := res.Body
	defer body.Close()

	// Un-marshall body data
	raw, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

// Download video to file
func downloadVideo(url string, dest string) {

	// create client
	client := grab.NewClient()
	req, _ := grab.NewRequest(dest, url)

	// start download
	fmt.Printf("Downloading %v...\n", req.URL())
	resp := client.Do(req)
	fmt.Printf("  %v\n", resp.HTTPResponse.Status)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			fmt.Printf("  transferred %v / %v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Download saved to ./%v \n", resp.Filename)
}

// Main function that download a youtube video
func download(videoUrl string) (filename string, err error) {

	fmt.Println("Download video :", videoUrl)

	videoId := ExtractVideoId(videoUrl)

	// Get video info
	videoInfo, err := GetVideoInfo(videoId)
	if err != nil {
		return "", err
	}

	videoResult := VideoResult {
		VideoId: videoId,
		Title: videoInfo.Get("title"),
		Duration: videoInfo.Get("length_seconds"),
		Author: videoInfo.Get("author"),
		Formats: Formats{},
	}

	// Get formats
	getFormatSpecs(videoInfo, &videoResult)

	// Get DASH formats
	dashmpd := videoInfo.Get("dashmpd")
	if dashmpd != "" {
		err := ParseMPDManifest(dashmpd, &videoResult)
		if err != nil {
			fmt.Println("Unable to download MPD Manifest")
		}
	} else {
		fmt.Println("Skip download MPD Manifest")
	}

	if opts.FormatList {
		PrintFormats(videoResult.Formats)
		return
	}

	// Build filename
	selectedFormat := strconv.Itoa(opts.Format)
	format := videoResult.Formats[selectedFormat]
	filename = videoResult.Title + "." + format.Ext;

	downloadVideo(format.Url, filename)

	return filename, nil
}

// Program options
type Options struct {
	FormatList bool `short:"F" long:"list-formats" description:"List all available formats of requested videos"`
	Format int `short:"f" long:"format" description:"Select video by format" require:"true"`
	Json bool `long:"json" description:"Output only json, disable other console print"`
	PrettyJson bool `long:"pretty-json" description:"Prettify JSON output"`
	Secure bool `short:"s" long:"secure" description:"Force HTTPS"`
	IgnoreErrors bool `short:"i" long:"ignore-errors" description:"Ignore errors"`
	Verbose bool `short:"v" long:"verbose" description:"Enable verbose mode"`
}

// Global program options
var opts Options

func main() {

	// Assume the rest is video urls
	args, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		log.Fatal(err)
	}

	// No videos to process
	if len(args) <= 0 {
		log.Fatal("No video to process")
	}

	// Json option disable other print
	if opts.Json {
		opts.Verbose = false
	}

	// First param is app executable
	for _, videoUrl := range args[1:] {

		//filename, err := download(videoUrl)
		_, err := download(videoUrl)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// fmt.Println("Download to", filename)
	}
}