package main

import (
	"ffslicer/core"
	"flag"
	"os"
)

func main() {
	vidfile := flag.String("i", "", "Input video file")
	tcfile := flag.String("c", "", "Comma seperated timecodes")
	args := flag.String("args", "", "Additional arguments for FFmpeg (slices will be re-encoded)")
	verbose := flag.Bool("v", false, "Verbose output")

	flag.Parse()
	
	if *vidfile == "" || *tcfile == "" {
		flag.Usage()
		os.Exit(1)
	}

	
	timedeltas := core.ParseCSV(*tcfile, *vidfile)

	core.RunFFmpeg(timedeltas, *vidfile, *args, *verbose)

}
