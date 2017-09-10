package main

import (
	"os"
	"fmt"
	"regexp"
	"net/http"
	"io/ioutil"
	"net/url"
	"log"
	"strings"
	"encoding/xml"
	"strconv"
	"sort"
	//"github.com/ryanuber/columnize"
	"github.com/dustin/go-humanize"
	"github.com/cavaliercoder/grab"
	"github.com/jessevdk/go-flags"
	"time"
	"encoding/json"
	"github.com/ryanuber/columnize"
)

// Video format
type Format struct {
	Ext string `json:"ext"`
	Width int `json:"width"`
	Height int `json:"height"`
	Acodec string `json:"acodec"`
	Vcodec string `json:"vcodec"`
	Format_note string `json:"format_note"`
	Preference int `json:"preference"`
	Container string `json:"container"`
	Fps int `json:"fps"`
	Abr int `json:"abr"`
	// ----
	Resolution string `json:"resolution"`
	Url string `json:"url"`
	Filesize uint64 `json:"filesize"`
	Tbr int `json:"tbr"`
}

type Dict struct {
	Format
	resolution int
	url string
}

// Formats
type Formats map[string]Format

// Video information
type VideoResult struct {
	video_id string
	title string
	url string
	duration string
	formats Formats
	view_count int
}

var formats = Formats {
	"5":  {Ext: "flv", Width: 400, Height: 240, Acodec: "mp3", Abr: 64, Vcodec: "h263"},
	"6":  {Ext: "flv", Width: 450, Height: 270, Acodec: "mp3", Abr: 64, Vcodec: "h263"},
	"13": {Ext: "3gp", Acodec: "aac", Vcodec: "mp4v"},
	"17": {Ext: "3gp", Width: 176, Height: 144, Acodec: "aac", Abr: 24, Vcodec: "mp4v"},
	"18": {Ext: "mp4", Width: 640, Height: 360, Acodec: "aac", Abr: 96, Vcodec: "h264"},
	"22": {Ext: "mp4", Width: 1280, Height: 720, Acodec: "aac", Abr: 192, Vcodec: "h264"},
	"34": {Ext: "flv", Width: 640, Height: 360, Acodec: "aac", Abr: 128, Vcodec: "h264"},
	"35": {Ext: "flv", Width: 854, Height: 480, Acodec: "aac", Abr: 128, Vcodec: "h264"},
	// itag 36 videos are either 320x180 (BaW_jenozKc) or 320x240 (__2ABJjxzNo), Abr varies as well
	"36": {Ext: "3gp", Width: 320, Acodec: "aac", Vcodec: "mp4v"},
	"37": {Ext: "mp4", Width: 1920, Height: 1080, Acodec: "aac", Abr: 192, Vcodec: "h264"},
	"38": {Ext: "mp4", Width: 4096, Height: 3072, Acodec: "aac", Abr: 192, Vcodec: "h264"},
	"43": {Ext: "webm", Width: 640, Height: 360, Acodec: "vorbis", Abr: 128, Vcodec: "vp8"},
	"44": {Ext: "webm", Width: 854, Height: 480, Acodec: "vorbis", Abr: 128, Vcodec: "vp8"},
	"45": {Ext: "webm", Width: 1280, Height: 720, Acodec: "vorbis", Abr: 192, Vcodec: "vp8"},
	"46": {Ext: "webm", Width: 1920, Height: 1080, Acodec: "vorbis", Abr: 192, Vcodec: "vp8"},
	"59": {Ext: "mp4", Width: 854, Height: 480, Acodec: "aac", Abr: 128, Vcodec: "h264"},
	"78": {Ext: "mp4", Width: 854, Height: 480, Acodec: "aac", Abr: 128, Vcodec: "h264"},

	// 3D videos
	"82": {Ext: "mp4", Height: 360, Format_note: "3D", Acodec: "aac", Abr: 128, Vcodec: "h264", Preference: -20},
	"83": {Ext: "mp4", Height: 480, Format_note: "3D", Acodec: "aac", Abr: 128, Vcodec: "h264", Preference: -20},
	"84": {Ext: "mp4", Height: 720, Format_note: "3D", Acodec: "aac", Abr: 192, Vcodec: "h264", Preference: -20},
	"85": {Ext: "mp4", Height: 1080, Format_note: "3D", Acodec: "aac", Abr: 192, Vcodec: "h264", Preference: -20},
	"100": {Ext: "webm", Height: 360, Format_note: "3D", Acodec: "vorbis", Abr: 128, Vcodec: "vp8", Preference: -20},
	"101": {Ext: "webm", Height: 480, Format_note: "3D", Acodec: "vorbis", Abr: 192, Vcodec: "vp8", Preference: -20},
	"102": {Ext: "webm", Height: 720, Format_note: "3D", Acodec: "vorbis", Abr: 192, Vcodec: "vp8", Preference: -20},

	// Apple HTTP Live Streaming
	"91": {Ext: "mp4", Height: 144, Format_note: "HLS", Acodec: "aac", Abr: 48, Vcodec: "h264", Preference: -10},
	"92": {Ext: "mp4", Height: 240, Format_note: "HLS", Acodec: "aac", Abr: 48, Vcodec: "h264", Preference: -10},
	"93": {Ext: "mp4", Height: 360, Format_note: "HLS", Acodec: "aac", Abr: 128, Vcodec: "h264", Preference: -10},
	"94": {Ext: "mp4", Height: 480, Format_note: "HLS", Acodec: "aac", Abr: 128, Vcodec: "h264", Preference: -10},
	"95": {Ext: "mp4", Height: 720, Format_note: "HLS", Acodec: "aac", Abr: 256, Vcodec: "h264", Preference: -10},
	"96": {Ext: "mp4", Height: 1080, Format_note: "HLS", Acodec: "aac", Abr: 256, Vcodec: "h264", Preference: -10},
	"132": {Ext: "mp4", Height: 240, Format_note: "HLS", Acodec: "aac", Abr: 48, Vcodec: "h264", Preference: -10},
	"151": {Ext: "mp4", Height: 72, Format_note: "HLS", Acodec: "aac", Abr: 24, Vcodec: "h264", Preference: -10},

	// DASH mp4 video
	"133": {Ext: "mp4", Height: 240, Format_note: "DASH video", Vcodec: "h264"},
	"134": {Ext: "mp4", Height: 360, Format_note: "DASH video", Vcodec: "h264"},
	"135": {Ext: "mp4", Height: 480, Format_note: "DASH video", Vcodec: "h264"},
	"136": {Ext: "mp4", Height: 720, Format_note: "DASH video", Vcodec: "h264"},
	"137": {Ext: "mp4", Height: 1080, Format_note: "DASH video", Vcodec: "h264"},
	"138": {Ext: "mp4", Format_note: "DASH video", Vcodec: "h264"},  // Height can vary (https://github.com/rg3/youtube-dl/issues/4559)
	"160": {Ext: "mp4", Height: 144, Format_note: "DASH video", Vcodec: "h264"},
	"212": {Ext: "mp4", Height: 480, Format_note: "DASH video", Vcodec: "h264"},
	"264": {Ext: "mp4", Height: 1440, Format_note: "DASH video", Vcodec: "h264"},
	"298": {Ext: "mp4", Height: 720, Format_note: "DASH video", Vcodec: "h264", Fps: 60},
	"299": {Ext: "mp4", Height: 1080, Format_note: "DASH video", Vcodec: "h264", Fps: 60},
	"266": {Ext: "mp4", Height: 2160, Format_note: "DASH video", Vcodec: "h264"},

	// Dash mp4 audio
	"139": {Ext: "m4a", Format_note: "DASH audio", Acodec: "aac", Abr: 48, Container: "m4a_dash"},
	"140": {Ext: "m4a", Format_note: "DASH audio", Acodec: "aac", Abr: 128, Container: "m4a_dash"},
	"141": {Ext: "m4a", Format_note: "DASH audio", Acodec: "aac", Abr: 256, Container: "m4a_dash"},
	"256": {Ext: "m4a", Format_note: "DASH audio", Acodec: "aac", Container: "m4a_dash"},
	"258": {Ext: "m4a", Format_note: "DASH audio", Acodec: "aac", Container: "m4a_dash"},
	"325": {Ext: "m4a", Format_note: "DASH audio", Acodec: "dtse", Container: "m4a_dash"},
	"328": {Ext: "m4a", Format_note: "DASH audio", Acodec: "ec-3", Container: "m4a_dash"},

	// Dash webm
	"167": {Ext: "webm", Height: 360, Width: 640, Format_note: "DASH video", Container: "webm", Vcodec: "vp8"},
	"168": {Ext: "webm", Height: 480, Width: 854, Format_note: "DASH video", Container: "webm", Vcodec: "vp8"},
	"169": {Ext: "webm", Height: 720, Width: 1280, Format_note: "DASH video", Container: "webm", Vcodec: "vp8"},
	"170": {Ext: "webm", Height: 1080, Width: 1920, Format_note: "DASH video", Container: "webm", Vcodec: "vp8"},
	"218": {Ext: "webm", Height: 480, Width: 854, Format_note: "DASH video", Container: "webm", Vcodec: "vp8"},
	"219": {Ext: "webm", Height: 480, Width: 854, Format_note: "DASH video", Container: "webm", Vcodec: "vp8"},
	"278": {Ext: "webm", Height: 144, Format_note: "DASH video", Container: "webm", Vcodec: "vp9"},
	"242": {Ext: "webm", Height: 240, Format_note: "DASH video", Vcodec: "vp9"},
	"243": {Ext: "webm", Height: 360, Format_note: "DASH video", Vcodec: "vp9"},
	"244": {Ext: "webm", Height: 480, Format_note: "DASH video", Vcodec: "vp9"},
	"245": {Ext: "webm", Height: 480, Format_note: "DASH video", Vcodec: "vp9"},
	"246": {Ext: "webm", Height: 480, Format_note: "DASH video", Vcodec: "vp9"},
	"247": {Ext: "webm", Height: 720, Format_note: "DASH video", Vcodec: "vp9"},
	"248": {Ext: "webm", Height: 1080, Format_note: "DASH video", Vcodec: "vp9"},
	"271": {Ext: "webm", Height: 1440, Format_note: "DASH video", Vcodec: "vp9"},
	// itag 272 videos are either 3840x2160 (e.g. RtoitU2A-3E) or 7680x4320 (sLprVF6d7Ug)
	"272": {Ext: "webm", Height: 2160, Format_note: "DASH video", Vcodec: "vp9"},
	"302": {Ext: "webm", Height: 720, Format_note: "DASH video", Vcodec: "vp9", Fps: 60},
	"303": {Ext: "webm", Height: 1080, Format_note: "DASH video", Vcodec: "vp9", Fps: 60},
	"308": {Ext: "webm", Height: 1440, Format_note: "DASH video", Vcodec: "vp9", Fps: 60},
	"313": {Ext: "webm", Height: 2160, Format_note: "DASH video", Vcodec: "vp9"},
	"315": {Ext: "webm", Height: 2160, Format_note: "DASH video", Vcodec: "vp9", Fps: 60},

	// Dash webm audio
	"171": {Ext: "webm", Acodec: "vorbis", Format_note: "DASH audio", Abr: 128},
	"172": {Ext: "webm", Acodec: "vorbis", Format_note: "DASH audio", Abr: 256},

	// Dash webm audio with opus inside
	"249": {Ext: "webm", Format_note: "DASH audio", Acodec: "opus", Abr: 50},
	"250": {Ext: "webm", Format_note: "DASH audio", Acodec: "opus", Abr: 70},
	"251": {Ext: "webm", Format_note: "DASH audio", Acodec: "opus", Abr: 160},
}

// Extract video id from youtube's url
func extractVideoId(videoUrl string) string {
	r := regexp.MustCompile(`(?:youtube\.com\/(?:[^\/]+\/.+\/|(?:v|e(?:mbed)?)\/|.*[?&]v=)|youtu\.be\/)([^"&?\/ ]{11})`)
	match := r.FindStringSubmatch(videoUrl)
	return match[1]
}

// Get video info from youtube
func getVideoInfo(videoId string) (url.Values, error) {

	//fmt.Println("Download video info", videoId)

	// QueryString
	query := url.Values{
		"video_id": { videoId },
		"el": { "info" },
	}

	// Full info url
	infoVideoUrl := "http://www.youtube.com/get_video_info?" + query.Encode()

	// Download page
	rawContent, err := downloadPage(infoVideoUrl)
	if err != nil {
		return nil, err
	}

	// Parse content as get_video_info provide a querystring
	videoInfo, err := url.ParseQuery(rawContent)
	if err != nil {
		return nil, err
	}

	return videoInfo, nil
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
	var fmtList = videoInfo.Get("fmt_list")
	for _, _fmt := range strings.Split(fmtList, ",") {
		specs := strings.Split(_fmt, "/")
		formatId := specs[0]
		//size := strings.Split(specs[2], "x")
		videoResult.formats[formatId] = Format{
			Resolution: specs[1],
			//height: itoa(size[0]),
			//..width: itoa(size[1]),
		}
	}

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
		newFormat := formats[formatId]
		newFormat.Url = videoUrl
		newFormat.Tbr = extractInt("bitrate")
		newFormat.Filesize = uint64(extractInt("clen"))
		newFormat.Format_note = urlData.Get("quality_label")
		newFormat.Resolution = fmt.Sprintf("%vx%v", newFormat.Width, newFormat.Height)
		videoResult.formats[formatId] = newFormat

		// Generals
		videoResult.view_count = extractInt("view_count")
	}
}

func getDashFormats(mpdUrl string, videoResult *VideoResult) {

	// Decrypt signature
	regSig := regexp.MustCompile(`/s/([a-fA-F0-9\.]+)`)
	match := regSig.FindStringSubmatch(mpdUrl)
	if len(match) > 1 {
		sig := match[1]
		mpdUrl = strings.Replace(mpdUrl, match[0], "/signature/" + sig, 1)
	}

	// Get manifest
	mpdContent, err := downloadPage(mpdUrl)
	if err != nil {
		log.Fatal(err)
	}

	// Parse MPD XML
	manifest := new(MPD)
	err = xml.Unmarshal([]byte(mpdContent), manifest)
	if err != nil {
		log.Fatal(err)
	}

	for _, periods := range manifest.Periods {
		for _, apSets := range periods.AdaptationSets {
			for _, reps := range apSets.Representations {

				id := strconv.Itoa(reps.Id)
				newFormat := formats[id]
				newFormat.Filesize = uint64(reps.BaseURL.ContentLength)
				newFormat.Tbr = reps.Bandwidth
				newFormat.Width = reps.Width
				newFormat.Height = reps.Height
				newFormat.Fps = reps.FrameRate
				newFormat.Resolution = fmt.Sprintf("%vx%v", newFormat.Width, newFormat.Height)
				videoResult.formats[id] = newFormat
			}
		}
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

/**func getPlayerConfig(videoId string) string {

	// QueryString
	query := url.Values{
		"v": { videoId },
		"gl": { "US" },
		"hl": { "en" },
		"has_verified": { "1" },
		"bpctr": { "9999999999" },
	}

	// Full info url
	videoPageUrl := "http://www.youtube.com/watch?" + query.Encode()

	videoContent, err := downloadPage(videoPageUrl)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(videoContent)

	return videoContent
}**/

// Main function that download a youtube video
func download(videoUrl string) (filename string, err error) {

	//fmt.Println("Download video with ID:", videoUrl)

	videoId := extractVideoId(videoUrl)

	// Get video info
	videoInfo, err := getVideoInfo(videoId)
	if err != nil {
		return "", err
	}

	videoResult := VideoResult {
		video_id: videoId,
		title: videoInfo.Get("title"),
		duration: videoInfo.Get("length_seconds"),
		formats: Formats{},
	}

	// Get formats
	getFormatSpecs(videoInfo, &videoResult)

	dashmpd := videoInfo.Get("dashmpd")
	//fmt.Println("dashmap", dashmpd)

	// Get DASH formats
	getDashFormats(dashmpd, &videoResult)

	if opts.FormatList {
		printFormats(videoResult.formats)
		return
	}

	// Build filename
	selectedFormat := strconv.Itoa(opts.Format)
	format := videoResult.formats[selectedFormat]
	filename = videoResult.title + "." + format.Ext;

	downloadVideo(format.Url, filename)

	return filename, nil
}

// Print available formats in the console
func printFormats(formats Formats) {

	// Headers
	headers := []string {
		"Format", "Extension", "Video", "Audio", "Resolution", "Size", "Note",
	}
	fHeaders := strings.Join(headers, " | ")
	lines := []string{ fHeaders, }

	// Sort for prettifying outputs
	keys := make([]int, 0, len(formats))
	for key, _ := range formats {
		ikey, _ := strconv.Atoi(key)
		keys = append(keys, ikey)
	}
	sort.Ints(keys)

	// Body
	for _, key := range keys {
		formatId := strconv.Itoa(key)
		format := formats[formatId]
		line := []string {
			formatId,
			format.Ext,
			format.Vcodec,
			format.Acodec,
			format.Resolution,
			humanize.Bytes(format.Filesize),
			format.Format_note,
		}
		lines = append(lines, strings.Join(line, " | "))
	}

	// Output in console or JSON
	fLines := columnize.SimpleFormat(lines)
	if opts.Verbose {
		fmt.Println(fLines)
	} else if opts.Json {

		var _f []Format
		for _, v := range formats {
			_f = append(_f, v)
		}

		js, err := json.Marshal(_f)
		//js, err := json.MarshalIndent(_f, "", "\t")
		if err != nil {
			log.Fatal(err)
			return
		}

		fmt.Println(string(js))
	}
}

// Program options
type Options struct {
	FormatList bool `short:"F" long:"list-formats" description:"List all available formats of requested videos"`
	Format int `short:"f" long:"format" description:"Select video by format" require:"true"`
	Json bool `long:"json" description:"Output only json, disable other console print"`
	Verbose bool `short:"v" long:"verbose" description:"Enable verbose mode"`
}

// Global program options
var opts Options

func main() {

	// Assume the rest is video urls
	args, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		log.Fatal(err)
		return
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

type MPD struct {
	Periods []Period `xml:"Period"`
}

type Period struct {
	Duration string `xml:"duration,attr"`
	AdaptationSets []AdaptationSet `xml:"AdaptationSet"`
}

type AdaptationSet struct {
	Id int `xml:"id,attr"`
	MimeType string `xml:"mimeType,attr"`
	SubsegmentAlignment bool `xml:"subsegmentAlignment,attr"`
	Representations []Representation `xml:"Representation"`
}

type Representation struct {
	Id int `xml:"id,attr"`
	Codecs string `xml:"codecs,attr"`
	AudioSamplingRate int `xml:"audioSamplingRate,attr"`
	Bandwidth int `xml:"bandwidth,attr"`
	BaseURL BaseURL `xml:"BaseURL"`
	Width int `xml:"width,attr"`
	Height int `xml:"height,attr"`
	FrameRate int `xml:"frameRate,attr"`
}

type BaseURL struct {
	ContentLength int `xml:"contentLength,attr"`
	Value string `xml:",chardata"`
}