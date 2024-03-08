package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type TimeFrame struct {
	startTime	float64
	duration	float64
}

func main() {
    filename := flag.String("i", "", "Input video file")
    tcfile := flag.String("c", "", "Comma seperated timecodes")
	
	flag.Parse()

	
	if *filename == "" || *tcfile == "" {
		flag.Usage()
		os.Exit(1)
	}
    
	fps := getVideoFps(*filename)

	file, err := os.Open(*tcfile) 
      
    if err != nil { 
        log.Fatal("Error while reading the file", err) 
    } 
  
    defer file.Close()
	
	r := csv.NewReader(file)

	rows, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	timeframes := make([]TimeFrame, len(rows))

	for i, row := range rows {
		start := parseTcToSecs(row[0], fps)
		end := parseTcToSecs(row[1], fps)
		duration := end - start + (1 / float64(fps))
		timeframe := TimeFrame{start, duration}
		timeframes[i] = timeframe
	}

	for i, tf := range timeframes {
		out := fmt.Sprintf("%s_%04d%s",strings.TrimSuffix(*filename, filepath.Ext(*filename)), i, filepath.Ext(*filename))
		arg := fmt.Sprintf("-y -i %s -ss %f -t %f -c:v copy -c:a copy %s", *filename, tf.startTime, tf.duration, out)
		args := strings.Split(arg, " ")
		cmd := exec.Command("ffmpeg", args...)

		var errb bytes.Buffer
		cmd.Stderr = &errb
		
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error running ffmpeg:", err)
			os.Exit(1)
		}

		fmt.Println(errb.String())

	}
	
	
}

func getVideoFps(file string) int {
	arg := "-v error -select_streams v:0 -show_entries stream=r_frame_rate -of default=noprint_wrappers=1:nokey=1 " + file
	args := strings.Split(arg, " ")
	cmd := exec.Command("ffprobe", args...)
	
	var out bytes.Buffer
    cmd.Stdout = &out
	
	err := cmd.Run()
    if err != nil {
        fmt.Println("Error running ffprobe:", err)
        os.Exit(1)
    }

    probe := strings.Split(out.String(), "/")[0]

	fps, err := strconv.Atoi(probe)
	if err != nil {
		fmt.Println("Error parsing file:", errors.New("file returned invalid framerate - it might be broken"))
        os.Exit(1)
	}
	
	return fps

}

func parseTcToSecs(tc string, tb int) float64 {
	tcsplit := strings.Split(tc, ":")
	if len(tcsplit) != 4 {
		fmt.Println("Error parsing timecode: ", tc, errors.New("invalid timecode format"))
		os.Exit(1)
	}
	hh, _ := strconv.ParseFloat((tcsplit[0]), 64)
	mm, _ := strconv.ParseFloat((tcsplit[1]), 64)
	ss, _ := strconv.ParseFloat((tcsplit[2]), 64)
	ff, _ := strconv.ParseFloat((tcsplit[3]), 64)

	return (hh * 60 * 60) + (mm * 60) + ss + (ff / float64(tb))

	
}