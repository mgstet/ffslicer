# ffslicer
Cuts videos into slices based on timecode defined in a csv file

```
Usage of ffslicer:
  -c string
        CSV Timecode list
  -i string
        Input video file
```
Specify a video file and a CSV with timcodes, e.g.:
```
00:03:02:18,00:04:02:18
00:05:03:12,00:05:07:01
00:10:06:14,00:12:09:23
00:15:07:03,00:17:08:09
00:03:02:18,00:04:02:18
```
This will produce 5 new sections of your input video file.

For now only works with intra-frame codecs (eg. ProRes, DNxHR,...) which don't need to be re-encoded when cut, and only with zero-based timecodes.