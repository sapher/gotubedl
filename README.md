# GotubeDL

[![Build Status](https://travis-ci.org/sapher/gotubedl.svg?branch=master)](https://travis-ci.org/sapher/gotubedl)

Youtube video downloader based on Go

Heavily based on Youtube-DL but without the awesomeness

:warning: Still new to go, this need a lot of Works

## Usage

`gotubedl https://www.youtube.com/watch?v=dQw4w9WgXcQ`

## Options

    gotubedl [OPTIONS]

    Application Options:
        -F, --list-formats   List all available formats of requested videos
        -f, --format=        Select video by format
            --json           Output only json, disable other console print
            --pretty-json    Prettify JSON output
        -s, --secure         Force HTTPS
        -i, --ignore-errors  Ignore errors
        -v, --verbose        Enable verbose mode

    Help Options:
        -h, --help           Show this help message

## XMas Lists

- [X] Output video formats as JSON
- [X] Select video format to download
- [ ] Solve slow video download
- [ ] Output template for filename
- [ ] Handle download of playlist(s)
- [ ] Download thumbnails
- [ ] Better progress bars
- [ ] Force HTTPS