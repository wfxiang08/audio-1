package aiff

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	expectations := []struct {
		input           string
		id              [4]byte
		size            uint32
		format          [4]byte
		commSize        uint32
		numChans        uint16
		numSampleFrames uint32
		sampleSize      uint16
		sampleRate      int
	}{
		{"fixtures/kick.aif", formID, 9642, aiffID,
			18, 1, 4484, 16, 22050},
	}

	for _, exp := range expectations {
		path, _ := filepath.Abs(exp.input)
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		c := NewDecoder(f, nil)
		err = c.Parse()
		if err != nil {
			t.Fatal(err)
		}
		if c.ID != exp.id {
			t.Fatalf("%s of %s didn't match %s, got %s", "ID", exp.input, exp.id, c.ID)
		}
		if c.Size != exp.size {
			t.Fatalf("%s of %s didn't match %d, got %d", "BlockSize", exp.input, exp.size, c.Size)
		}
		if c.Format != exp.format {
			t.Fatalf("%s of %s didn't match %q, got %q", "Format", exp.input, exp.format, c.Format)
		}
		// comm chunk
		if c.commSize != exp.commSize {
			t.Fatalf("%s of %s didn't match %d, got %d", "comm size", exp.input, exp.commSize, c.commSize)
		}
		if c.NumChans != exp.numChans {
			t.Fatalf("%s of %s didn't match %d, got %d", "NumChans", exp.input, exp.numChans, c.NumChans)
		}
		if c.NumSampleFrames != exp.numSampleFrames {
			t.Fatalf("%s of %s didn't match %d, got %d", "NumSampleFrames", exp.input, exp.numSampleFrames, c.NumSampleFrames)
		}
		if c.SampleSize != exp.sampleSize {
			t.Fatalf("%s of %s didn't match %d, got %d", "SampleSize", exp.input, exp.sampleSize, c.SampleSize)
		}
		if c.SampleRate != exp.sampleRate {
			t.Fatalf("%s of %s didn't match %d, got %d", "SampleRate", exp.input, exp.sampleRate, c.SampleRate)
		}
	}
}

func TestNewDecoder(t *testing.T) {
	path, _ := filepath.Abs("fixtures/kick.aif")
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ch := make(chan *Chunk)
	c := NewDecoder(f, ch)
	go func() {
		if err := c.Parse(); err != nil {
			panic(err)
		}
	}()

	for chunk := range ch {
		id := string(chunk.ID[:])
		t.Log(id, chunk.Size)
		if id != "SSND" {
			buf := make([]byte, chunk.Size)
			chunk.ReadBE(buf)
			t.Log(hex.Dump(buf))
		}
		chunk.Done()
	}

}

func TestReadFrames(t *testing.T) {
	path, _ := filepath.Abs("fixtures/kick.aif")
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	r := NewDecoder(f, nil)
	info, frames, err := r.Frames()
	if err != nil {
		t.Fatal(err)
	}
	if info.SampleRate != 22050 {
		t.Fatalf("unexpected sample rate: %d", info.SampleRate)
	}
	if info.BitDepth != 16 {
		t.Fatalf("unexpected sample size: %d", info.BitDepth)
	}
	if info.NumChannels != 1 {
		t.Fatalf("unexpected channel number: %d", info.NumChannels)
	}

	if totalFrames := len(frames); totalFrames != 4484 {
		t.Fatalf("unexpected total frames: %d", totalFrames)
	}
}

func TestDuration(t *testing.T) {
	expectations := []struct {
		input    string
		duration time.Duration
	}{
		{"fixtures/kick.aif", time.Duration(203356009)},
	}

	for _, exp := range expectations {
		path, _ := filepath.Abs(exp.input)
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		c := NewDecoder(f, nil)
		err = c.Parse()
		if err != nil {
			t.Fatal(err)
		}
		d, err := c.Duration()
		if err != nil {
			t.Fatal(err)
		}
		if d != exp.duration {
			t.Fatalf("duration of %s didn't match %d milliseconds, got %d", exp.input, exp.duration, d)
		}
	}
}

func ExampleDecoder_Duration() {
	path, _ := filepath.Abs("fixtures/kick.aif")
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	c := NewDecoder(f, nil)
	if err = c.Parse(); err != nil {
		panic(err)
	}
	d, _ := c.Duration()
	fmt.Printf("kick.aif has a duration of %f seconds\n", d.Seconds())
	// Output:
	// kick.aif has a duration of 0.203356 seconds
}