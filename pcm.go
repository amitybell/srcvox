package main

import (
	"github.com/amitybell/memio"
	"github.com/gopxl/beep"
)

func StreamToPCM(stream beep.Streamer, format beep.Format) *memio.File {
	out := &memio.File{}
	samples := make([][2]float64, 512)
	buffer := make([]byte, len(samples)*format.Width())
	for {
		n, ok := stream.Stream(samples)
		if !ok {
			break
		}
		buf := buffer
		for _, sample := range samples[:n] {
			buf = buf[format.EncodeSigned(buf, sample):]
		}
		out.Write(buffer[:n*format.Width()])
	}
	out.Seek(0, 0)
	return out
}
