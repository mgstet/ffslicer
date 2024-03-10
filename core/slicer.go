package core

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type TimeDelta struct {
	startTime float64
	duration  float64
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

func parseTcToSecs(tc string, tb int) (float64, error) {
	tcsplit := strings.Split(tc, ":")
	if len(tcsplit) != 4 {
		return 0, errors.New("invalid timecode format")
	}

	var errs []error
	hh, herr := strconv.ParseFloat((tcsplit[0]), 64)
	mm, merr := strconv.ParseFloat((tcsplit[1]), 64)
	ss, serr := strconv.ParseFloat((tcsplit[2]), 64)
	ff, ferr := strconv.ParseFloat((tcsplit[3]), 64)

	errs = append(errs, herr, merr, serr, ferr)

	if hh < 0 || mm < 0 || ss < 0 || ff < 0 || ff >= float64(tb) {
		errs = append(errs, errors.New("timecode cannot be negative"))
	}

	if err := errors.Join(errs...); err != nil {
		return 0, err
	}

	seconds := (hh * 60 * 60) + (mm * 60) + ss + (ff / float64(tb))

	return seconds, nil

}

func ParseCSV(tcFile string, videoFile string) []TimeDelta {

	fps := getVideoFps(videoFile)

	file, err := os.Open(tcFile)

	if err != nil {
		log.Fatal("Error while reading the file", err)
	}

	defer file.Close()

	r := csv.NewReader(file)

	rows, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	timedeltas := []TimeDelta{}

	for i, row := range rows {
		start, errs := parseTcToSecs(row[0], fps)
		end, erre := parseTcToSecs(row[1], fps)
		if err := errors.Join(errs, erre); err != nil {
			slog.Error("failed to parse timecodes", "tc", row, "row", i, "err", err)
			slog.Info("Skipping row", "row", row, "index", i)
			continue
		}
		if err := errors.New("start cannot be bigger than end"); start >= end {
			slog.Error("failed to parse timecodes", "tc", row, "row", i, "err", err)
			slog.Info("Skipping", "tc", row, "row", i)
			continue
		}
		duration := end - start + (1 / float64(fps))
		timedelta := TimeDelta{start, duration}
		timedeltas = append(timedeltas, timedelta)
	}

	return timedeltas
}

func RunFFmpeg(timedeltas []TimeDelta, videoFile string, verbose bool) {

	for i, tf := range timedeltas {
		out := fmt.Sprintf("%s_%04d%s", strings.TrimSuffix(videoFile, filepath.Ext(videoFile)), i, filepath.Ext(videoFile))
		arg := fmt.Sprintf("-y -i %s -ss %f -t %f -c:v copy -c:a copy %s", videoFile, tf.startTime, tf.duration, out)
		args := strings.Split(arg, " ")
		cmd := exec.Command("ffmpeg", args...)
		slog.Info("Processing slice #", strconv.Itoa(i), out)

		var errb bytes.Buffer
		cmd.Stderr = &errb

		err := cmd.Run()
		if err != nil {
			fmt.Println("Error running ffmpeg:", err)
			os.Exit(1)
		}

		if verbose {
			fmt.Println(errb.String())
		}

	}
}
