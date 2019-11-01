# RetroWave downloader

Easy download tracks from retrowave.ru radio

![](https://cdn.jsdelivr.net/gh/xpyct/retrowave-dl/term.svg)

## Build
```ssh
make build
```

## Usage

```ssh
xpyct@xpyct:$ ./retrowave-dl -h
Usage of ./retrowave-dl:
  -all
        get all possible tracks (ignoring --limit flag)
  -json
        download track list as JSON file
  -limit int
        tracks number for download (default 2)
  -out string
        directory for output
  -sync
        synchronize downloaded files

```

## Examples of usage

Download as JSON file
```ssh
retrowave-dl --json --all // output in ./downloads/soundtracks.json
```

Download only 10 tracks
```ssh
retrowave-dl --limit 10 // output in ./downloads/**/*.mp3
```
Download all tracks
```ssh
retrowave-dl --all // output in ./downloads/**/*.mp3
```

Download only new tracks
```ssh
retrowave-dl --all --sync // output in ./downloads/**/*.mp3
```
