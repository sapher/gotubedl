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
	//"github.com/dustin/go-humanize"
	"github.com/ryanuber/columnize"
	"github.com/dustin/go-humanize"
	"github.com/cavaliercoder/grab"
	"time"
)

// Video format
type Format struct {
	ext string
	width int
	height int
	acodec string
	vcodec string
	format_note string
	preference int
	container string
	fps int
	abr int
	// ----
	resolution string
	url string
	filesize uint64
	tbr int
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
	"5":  {ext: "flv", width: 400, height: 240, acodec: "mp3", abr: 64, vcodec: "h263"},
	"6":  {ext: "flv", width: 450, height: 270, acodec: "mp3", abr: 64, vcodec: "h263"},
	"13": {ext: "3gp", acodec: "aac", vcodec: "mp4v"},
	"17": {ext: "3gp", width: 176, height: 144, acodec: "aac", abr: 24, vcodec: "mp4v"},
	"18": {ext: "mp4", width: 640, height: 360, acodec: "aac", abr: 96, vcodec: "h264"},
	"22": {ext: "mp4", width: 1280, height: 720, acodec: "aac", abr: 192, vcodec: "h264"},
	"34": {ext: "flv", width: 640, height: 360, acodec: "aac", abr: 128, vcodec: "h264"},
	"35": {ext: "flv", width: 854, height: 480, acodec: "aac", abr: 128, vcodec: "h264"},
	// itag 36 videos are either 320x180 (BaW_jenozKc) or 320x240 (__2ABJjxzNo), abr varies as well
	"36": {ext: "3gp", width: 320, acodec: "aac", vcodec: "mp4v"},
	"37": {ext: "mp4", width: 1920, height: 1080, acodec: "aac", abr: 192, vcodec: "h264"},
	"38": {ext: "mp4", width: 4096, height: 3072, acodec: "aac", abr: 192, vcodec: "h264"},
	"43": {ext: "webm", width: 640, height: 360, acodec: "vorbis", abr: 128, vcodec: "vp8"},
	"44": {ext: "webm", width: 854, height: 480, acodec: "vorbis", abr: 128, vcodec: "vp8"},
	"45": {ext: "webm", width: 1280, height: 720, acodec: "vorbis", abr: 192, vcodec: "vp8"},
	"46": {ext: "webm", width: 1920, height: 1080, acodec: "vorbis", abr: 192, vcodec: "vp8"},
	"59": {ext: "mp4", width: 854, height: 480, acodec: "aac", abr: 128, vcodec: "h264"},
	"78": {ext: "mp4", width: 854, height: 480, acodec: "aac", abr: 128, vcodec: "h264"},

	// 3D videos
	"82": {ext: "mp4", height: 360, format_note: "3D", acodec: "aac", abr: 128, vcodec: "h264", preference: -20},
	"83": {ext: "mp4", height: 480, format_note: "3D", acodec: "aac", abr: 128, vcodec: "h264", preference: -20},
	"84": {ext: "mp4", height: 720, format_note: "3D", acodec: "aac", abr: 192, vcodec: "h264", preference: -20},
	"85": {ext: "mp4", height: 1080, format_note: "3D", acodec: "aac", abr: 192, vcodec: "h264", preference: -20},
	"100": {ext: "webm", height: 360, format_note: "3D", acodec: "vorbis", abr: 128, vcodec: "vp8", preference: -20},
	"101": {ext: "webm", height: 480, format_note: "3D", acodec: "vorbis", abr: 192, vcodec: "vp8", preference: -20},
	"102": {ext: "webm", height: 720, format_note: "3D", acodec: "vorbis", abr: 192, vcodec: "vp8", preference: -20},

	// Apple HTTP Live Streaming
	"91": {ext: "mp4", height: 144, format_note: "HLS", acodec: "aac", abr: 48, vcodec: "h264", preference: -10},
	"92": {ext: "mp4", height: 240, format_note: "HLS", acodec: "aac", abr: 48, vcodec: "h264", preference: -10},
	"93": {ext: "mp4", height: 360, format_note: "HLS", acodec: "aac", abr: 128, vcodec: "h264", preference: -10},
	"94": {ext: "mp4", height: 480, format_note: "HLS", acodec: "aac", abr: 128, vcodec: "h264", preference: -10},
	"95": {ext: "mp4", height: 720, format_note: "HLS", acodec: "aac", abr: 256, vcodec: "h264", preference: -10},
	"96": {ext: "mp4", height: 1080, format_note: "HLS", acodec: "aac", abr: 256, vcodec: "h264", preference: -10},
	"132": {ext: "mp4", height: 240, format_note: "HLS", acodec: "aac", abr: 48, vcodec: "h264", preference: -10},
	"151": {ext: "mp4", height: 72, format_note: "HLS", acodec: "aac", abr: 24, vcodec: "h264", preference: -10},

	// DASH mp4 video
	"133": {ext: "mp4", height: 240, format_note: "DASH video", vcodec: "h264"},
	"134": {ext: "mp4", height: 360, format_note: "DASH video", vcodec: "h264"},
	"135": {ext: "mp4", height: 480, format_note: "DASH video", vcodec: "h264"},
	"136": {ext: "mp4", height: 720, format_note: "DASH video", vcodec: "h264"},
	"137": {ext: "mp4", height: 1080, format_note: "DASH video", vcodec: "h264"},
	"138": {ext: "mp4", format_note: "DASH video", vcodec: "h264"},  // Height can vary (https://github.com/rg3/youtube-dl/issues/4559)
	"160": {ext: "mp4", height: 144, format_note: "DASH video", vcodec: "h264"},
	"212": {ext: "mp4", height: 480, format_note: "DASH video", vcodec: "h264"},
	"264": {ext: "mp4", height: 1440, format_note: "DASH video", vcodec: "h264"},
	"298": {ext: "mp4", height: 720, format_note: "DASH video", vcodec: "h264", fps: 60},
	"299": {ext: "mp4", height: 1080, format_note: "DASH video", vcodec: "h264", fps: 60},
	"266": {ext: "mp4", height: 2160, format_note: "DASH video", vcodec: "h264"},

	// Dash mp4 audio
	"139": {ext: "m4a", format_note: "DASH audio", acodec: "aac", abr: 48, container: "m4a_dash"},
	"140": {ext: "m4a", format_note: "DASH audio", acodec: "aac", abr: 128, container: "m4a_dash"},
	"141": {ext: "m4a", format_note: "DASH audio", acodec: "aac", abr: 256, container: "m4a_dash"},
	"256": {ext: "m4a", format_note: "DASH audio", acodec: "aac", container: "m4a_dash"},
	"258": {ext: "m4a", format_note: "DASH audio", acodec: "aac", container: "m4a_dash"},
	"325": {ext: "m4a", format_note: "DASH audio", acodec: "dtse", container: "m4a_dash"},
	"328": {ext: "m4a", format_note: "DASH audio", acodec: "ec-3", container: "m4a_dash"},

	// Dash webm
	"167": {ext: "webm", height: 360, width: 640, format_note: "DASH video", container: "webm", vcodec: "vp8"},
	"168": {ext: "webm", height: 480, width: 854, format_note: "DASH video", container: "webm", vcodec: "vp8"},
	"169": {ext: "webm", height: 720, width: 1280, format_note: "DASH video", container: "webm", vcodec: "vp8"},
	"170": {ext: "webm", height: 1080, width: 1920, format_note: "DASH video", container: "webm", vcodec: "vp8"},
	"218": {ext: "webm", height: 480, width: 854, format_note: "DASH video", container: "webm", vcodec: "vp8"},
	"219": {ext: "webm", height: 480, width: 854, format_note: "DASH video", container: "webm", vcodec: "vp8"},
	"278": {ext: "webm", height: 144, format_note: "DASH video", container: "webm", vcodec: "vp9"},
	"242": {ext: "webm", height: 240, format_note: "DASH video", vcodec: "vp9"},
	"243": {ext: "webm", height: 360, format_note: "DASH video", vcodec: "vp9"},
	"244": {ext: "webm", height: 480, format_note: "DASH video", vcodec: "vp9"},
	"245": {ext: "webm", height: 480, format_note: "DASH video", vcodec: "vp9"},
	"246": {ext: "webm", height: 480, format_note: "DASH video", vcodec: "vp9"},
	"247": {ext: "webm", height: 720, format_note: "DASH video", vcodec: "vp9"},
	"248": {ext: "webm", height: 1080, format_note: "DASH video", vcodec: "vp9"},
	"271": {ext: "webm", height: 1440, format_note: "DASH video", vcodec: "vp9"},
	// itag 272 videos are either 3840x2160 (e.g. RtoitU2A-3E) or 7680x4320 (sLprVF6d7Ug)
	"272": {ext: "webm", height: 2160, format_note: "DASH video", vcodec: "vp9"},
	"302": {ext: "webm", height: 720, format_note: "DASH video", vcodec: "vp9", fps: 60},
	"303": {ext: "webm", height: 1080, format_note: "DASH video", vcodec: "vp9", fps: 60},
	"308": {ext: "webm", height: 1440, format_note: "DASH video", vcodec: "vp9", fps: 60},
	"313": {ext: "webm", height: 2160, format_note: "DASH video", vcodec: "vp9"},
	"315": {ext: "webm", height: 2160, format_note: "DASH video", vcodec: "vp9", fps: 60},

	// Dash webm audio
	"171": {ext: "webm", acodec: "vorbis", format_note: "DASH audio", abr: 128},
	"172": {ext: "webm", acodec: "vorbis", format_note: "DASH audio", abr: 256},

	// Dash webm audio with opus inside
	"249": {ext: "webm", format_note: "DASH audio", acodec: "opus", abr: 50},
	"250": {ext: "webm", format_note: "DASH audio", acodec: "opus", abr: 70},
	"251": {ext: "webm", format_note: "DASH audio", acodec: "opus", abr: 160},
}

// Extract video id from youtube's url
func extractVideoId(videoUrl string) string {
	r := regexp.MustCompile(`(?:youtube\.com\/(?:[^\/]+\/.+\/|(?:v|e(?:mbed)?)\/|.*[?&]v=)|youtu\.be\/)([^"&?\/ ]{11})`)
	match := r.FindStringSubmatch(videoUrl)
	return match[1]
}

// Get video info from youtube
func getVideoInfo(videoId string) (url.Values, error) {

	fmt.Println("Download video info", videoId)

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
			resolution: specs[1],
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
		newFormat.url = videoUrl
		newFormat.tbr = extractInt("bitrate")
		newFormat.filesize = uint64(extractInt("clen"))
		newFormat.format_note = urlData.Get("quality_label")
		newFormat.resolution = fmt.Sprintf("%vx%v", newFormat.width, newFormat.height)
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
				newFormat.filesize = uint64(reps.BaseURL.ContentLength)
				newFormat.tbr = reps.Bandwidth
				newFormat.width = reps.Width
				newFormat.height = reps.Height
				newFormat.fps = reps.FrameRate
				newFormat.resolution = fmt.Sprintf("%vx%v", newFormat.width, newFormat.height)
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

	fmt.Println("Download video with ID:", videoUrl)

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

	printFormats(videoResult.formats)

	// Build filename
	format := videoResult.formats["249"]
	filename = videoResult.title + "." + format.ext;

	downloadVideo(format.url, filename)

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
			format.ext,
			format.vcodec,
			format.acodec,
			format.resolution,
			humanize.Bytes(format.filesize),
			format.format_note,
		}
		lines = append(lines, strings.Join(line, " | "))
	}

	// Build output
	fLines := columnize.SimpleFormat(lines)
	fmt.Println(fLines)
}

func main() {

	videoUrls := os.Args[1:]

	for _, videoUrl := range videoUrls {
		filename, err := download(videoUrl)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("Download to", filename)
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