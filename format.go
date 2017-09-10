package main

import (
	"strings"
	"github.com/ryanuber/columnize"
	"fmt"
	"encoding/json"
	"strconv"
	"sort"
	"github.com/dustin/go-humanize"
)

// Video format
type Format struct {
	FormatId int `json:"format_id"`
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
	Asr int `json:"asr"`
	// ----
	Resolution string `json:"resolution"`
	Url string `json:"url"`
	Filesize uint64 `json:"filesize"`
	Tbr int `json:"tbr"`
}

// Formats
type Formats map[string]Format

// BasedFormats
var BaseFormats = Formats {
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


// Print available formats in the console
func PrintFormats(formats Formats) {

	// Headers
	headers := []string {
		"Format", "Extension", "Video", "Audio", "Resolution", "Size", "Note", "All",
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

		// Data
		formatId := strconv.Itoa(key)
		format := formats[formatId]

		// Audio sampling rate
		/**samplingRate := ""
		if format.Asr != 0 {
			samplingRate = strconv.Itoa(format.Asr) + " hz"
		}**/

		var others []string
		// FPS
		if format.Fps > 0 {
			others = append(others, fmt.Sprintf("@%dfps", format.Fps))
		}

		// Build string
		line := []string {
			formatId,
			format.Ext,
			format.Vcodec,
			format.Acodec,
			///samplingRate,
			format.Resolution,
			humanize.Bytes(format.Filesize),
			format.Format_note,
			strings.Join(others, " "),
		}
		lines = append(lines, strings.Join(line, " | "))
	}

	// Output in console or JSON
	fLines := columnize.SimpleFormat(lines)
	if !opts.Json {
		fmt.Println(fLines)
	} else if opts.Json {

		var _f []Format
		for _, v := range formats {
			_f = append(_f, v)
		}

		// Decide if output pretty json or not
		var jsonOutput []byte
		if opts.PrettyJson {
			jsonOutput, _ = json.MarshalIndent(_f, "", "\t")
		} else {
			jsonOutput, _ = json.Marshal(_f)
		}

		fmt.Println(string(jsonOutput))
	}
}