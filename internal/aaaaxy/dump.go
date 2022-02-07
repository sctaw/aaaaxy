// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aaaaxy

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/divVerent/aaaaxy/internal/atexit"
	"github.com/divVerent/aaaaxy/internal/audiowrap"
	"github.com/divVerent/aaaaxy/internal/demo"
	"github.com/divVerent/aaaaxy/internal/engine"
	"github.com/divVerent/aaaaxy/internal/flag"
	"github.com/divVerent/aaaaxy/internal/log"
	m "github.com/divVerent/aaaaxy/internal/math"
	"github.com/divVerent/aaaaxy/internal/namedpipe"
)

var (
	dumpVideo            = flag.String("dump_video", "", "filename prefix to dump game frames to")
	dumpVideoFpsDivisor  = flag.Int("dump_video_fps_divisor", 1, "frame rate divisor (try 2 for faster dumping)")
	dumpAudio            = flag.String("dump_audio", "", "filename to dump game audio to")
	dumpMedia            = flag.String("dump_media", "", "filename to dump game media to; exclusive with dump_video and dump_audio")
	cheatDumpSlowAndGood = flag.Bool("cheat_dump_slow_and_good", false, "non-realtime video dumping (slows down the game, thus considered a cheat))")
)

type WriteCloserAt interface {
	io.Writer
	io.WriterAt
	io.Closer
}

var (
	dumpFrameCount = int64(0)
	dumpVideoFile  WriteCloserAt
	dumpAudioFile  WriteCloserAt
	dumpVideoPipe  *namedpipe.Fifo
	dumpAudioPipe  *namedpipe.Fifo
	dumpMediaCmd   *exec.Cmd
)

const (
	dumpVideoFrameSize = engine.GameWidth * engine.GameHeight * 4
)

var (
	dumpVideoFrame int64 = 0
	dumpVideoWg    sync.WaitGroup
)

func initDumpingEarly() error {
	if *dumpMedia != "" {
		if *dumpVideo != "" || *dumpAudio != "" {
			return fmt.Errorf("-dump_media is mutually exclusive with -dump_video/-dump_audio")
		}
		var err error
		dumpAudioPipe, err = namedpipe.New("aaaaxy-audio", 120, 4*96000)
		if err != nil {
			return fmt.Errorf("could not create audio pipe: %v", err)
		}
		dumpVideoPipe, err = namedpipe.New("aaaaxy-video", 120, dumpVideoFrameSize)
		if err != nil {
			return fmt.Errorf("could not create video pipe: %v", err)
		}
		dumpAudioFile = namedpipe.NewWriteCloserAt(dumpAudioPipe)
		dumpVideoFile = namedpipe.NewWriteCloserAt(dumpVideoPipe)
		audiowrap.InitDumping()
	}

	if *dumpAudio != "" {
		var err error
		dumpAudioFile, err = os.Create(*dumpAudio)
		if err != nil {
			return fmt.Errorf("could not initialize audio dump: %v", err)
		}
		audiowrap.InitDumping()
	}

	if *dumpVideo != "" {
		var err error
		dumpVideoFile, err = os.Create(*dumpVideo)
		if err != nil {
			return fmt.Errorf("could not initialize video dump: %v", err)
		}
	}

	return nil
}

func initDumpingLate() error {
	if *dumpMedia != "" {
		cmdLine, _, err := ffmpegCommand(dumpAudioPipe.Path(), dumpVideoPipe.Path(), *dumpMedia, *screenFilter)
		if err != nil {
			return err
		}
		dumpMediaCmd := exec.Command(cmdLine[0], cmdLine[1:]...)
		dumpMediaCmd.Stdout = os.Stdout
		dumpMediaCmd.Stderr = os.Stderr
		err = dumpMediaCmd.Start()
		if err != nil {
			return fmt.Errorf("could not launch FFmpeg: %v", err)
		}
	}

	return nil
}

func dumping() bool {
	return dumpAudioFile != nil || dumpVideoFile != nil
}

func slowDumping() bool {
	return dumping() && (*cheatDumpSlowAndGood || demo.Playing())
}

func dumpFrameThenReturnTo(screen *ebiten.Image, to chan *ebiten.Image, frames int) {
	if !dumping() || frames == 0 {
		to <- screen
		return
	}
	if dumpVideoFile != nil {
		dumpVideoFrameBegin := dumpFrameCount / int64(*dumpVideoFpsDivisor)
		dumpFrameCount += int64(frames)
		dumpVideoFrameEnd := dumpFrameCount / int64(*dumpVideoFpsDivisor)
		cnt := dumpVideoFrameEnd - dumpVideoFrameBegin
		if cnt > 0 {
			if cnt > 1 {
				log.Infof("video dump: %v frames dropped", cnt-1)
			}
			dumpVideoWg.Add(1)
			dumpPixelsRGBA(screen, func(pix []byte, err error) {
				to <- screen
				if err == nil {
					for i := dumpVideoFrameBegin; i < dumpVideoFrameEnd; i++ {
						_, err = dumpVideoFile.WriteAt(pix, i*dumpVideoFrameSize)
						if err != nil {
							break
						}
					}
				}
				if err != nil {
					log.Errorf("failed to encode video - expect corruption: %v", err)
					// dumpVideoFile.Close()
					// dumpVideoFile = nil
				}
				dumpVideoWg.Done()
			})
		} else {
			// log.Infof("video dump: frame skipped")
			to <- screen
		}
	} else {
		to <- screen
	}
	if dumpAudioFile != nil {
		err := audiowrap.DumpFrame(dumpAudioFile, time.Duration(dumpFrameCount)*time.Second/engine.GameTPS)
		if err != nil {
			log.Errorf("failed to encode audio - expect corruption: %v", err)
			dumpAudioFile.Close()
			dumpAudioFile = nil
		}
	}
}

func ffmpegCommand(audio, video, output, screenFilter string) ([]string, string, error) {
	precmd := ""
	inputs := []string{}
	settings := []string{"-vsync", "vfr", "-y"}
	// Video first, so we can refer to the video stream as [0:v] for sure.
	if video != "" {
		fps := float64(engine.GameTPS) / (float64(*fpsDivisor) * float64(*dumpVideoFpsDivisor))
		inputs = append(inputs, "-f", "rawvideo", "-pixel_format", "rgba", "-video_size", fmt.Sprintf("%dx%d", engine.GameWidth, engine.GameHeight), "-r", fmt.Sprint(fps), "-i", video)
		filterComplex := "[0:v]premultiply=inplace=1,format=gbrp[lowres]; "
		switch screenFilter {
		case "linear":
			filterComplex += "[lowres]scale=1920:1080"
		case "simple", "linear2x":
			// Note: the two step upscale simulates the effect of the linear2xcrt shader.
			// "simple" does the same as "linear2x" if the screen res is exactly 1080p.
			filterComplex += "[lowres]scale=1280:720:flags=neighbor,scale=1920:1080"
		case "linear2xcrt":
			// For 3x scale, pattern is: 1 (1-2/3*f) 1.
			// darkened := m.Rint(255 * (1.0 - 2.0/3.0**screenFilterScanLines))
			// pnm := fmt.Sprintf("P2 1 3 255 %d 255 %d", darkened, darkened)
			// Then second scale is to 1920:1080.
			// But for the lens correction, we gotta do better.
			// For 6x scale, pattern is: (1-5/6*f) (1-3/6*f) (1-1/6*f) (1-1/6*f) (1-3/6*f) (1-5/6*f).
			pnm := fmt.Sprintf("P2\n1 6 255 %d %d %d %d %d %d",
				m.Rint(255*(1.0-5.0/6.0**screenFilterScanLines)),
				m.Rint(255*(1.0-3.0/6.0**screenFilterScanLines)),
				m.Rint(255*(1.0-1.0/6.0**screenFilterScanLines)),
				m.Rint(255*(1.0-1.0/6.0**screenFilterScanLines)),
				m.Rint(255*(1.0-3.0/6.0**screenFilterScanLines)),
				m.Rint(255*(1.0-5.0/6.0**screenFilterScanLines)))
			tempFile, err := ioutil.TempFile("", "aaaaxy-*")
			if err != nil {
				return nil, "", err
			}
			atexit.Delete(tempFile.Name())
			_, err = tempFile.Write([]byte(pnm))
			if err != nil {
				return nil, "", err
			}
			err = tempFile.Close()
			if err != nil {
				return nil, "", err
			}
			precmd = fmt.Sprintf("{ echo '%s'; echo '%s'; } > '%s'; ", pnm[:2], pnm[3:], tempFile.Name())
			filterComplex += fmt.Sprintf("[lowres]scale=1280:720:flags=neighbor,scale=3840:2160[scaled]; movie=filename=%v:loop=360,tile=1x360,scale=3840:2160:flags=neighbor,format=gbrp[scanlines]; [scaled][scanlines]blend=all_mode=multiply,lenscorrection=i=bilinear:k1=%f:k2=%f", tempFile.Name(), crtK1(), crtK2())
		case "nearest":
			filterComplex += "[lowres]scale=1920:1080:flags=neighbor"
		case "":
			filterComplex += "[lowres]copy"
		}
		// Note: using high quality, fast settings and many keyframes
		// as the assumption is that the output file will be further edited.
		// Note: disabling 8x8 DCT here as some older FFmpeg versions -
		// or even newer versions with decoding options changed for compatibility,
		// if the video file has also been losslessly cut -
		// have trouble decoding that.
		settings = append(settings, "-codec:v", "libx264", "-profile:v", "high444", "-preset:v", "fast", "-crf:v", "10", "-8x8dct:v", "0", "-keyint_min", "10", "-g", "60", "-filter_complex", filterComplex)
	}
	if audio != "" {
		inputs = append(inputs, "-f", "s16le", "-ac", "2", "-ar", fmt.Sprint(audiowrap.SampleRate()), "-i", audio)
		settings = append(settings, "-codec:a", "aac", "-b:a", "128k")
	}
	cmd := []string{"ffmpeg"}
	cmd = append(cmd, inputs...)
	cmd = append(cmd, settings...)
	cmd = append(cmd, output)
	return cmd, precmd, nil
}

func printCommand(cmd []string) string {
	r := []string{}
	for _, arg := range cmd {
		r = append(r, "'"+strings.ReplaceAll(arg, "'", "'\\''")+"'")
	}
	return strings.Join(r, " ")
}

func finishDumping() error {
	if !dumping() {
		return nil
	}
	if dumpVideoFile != nil {
		dumpVideoWg.Wait()
	}
	if dumpAudioFile != nil {
		err := dumpAudioFile.Close()
		if err != nil {
			return fmt.Errorf("failed to close audio - expect corruption: %v", err)
		}
		dumpAudioFile = nil
	}
	if dumpVideoFile != nil {
		err := dumpVideoFile.Close()
		if err != nil {
			return fmt.Errorf("failed to close video - expect corruption: %v", err)
		}
		dumpVideoFile = nil
	}
	if dumpMediaCmd != nil {
		err := dumpMediaCmd.Wait()
		if err != nil {
			return fmt.Errorf("failed to close FFmpeg - expect corruption: %v", err)
		}
	}
	log.Infof("media has been dumped")
	log.Infof("to create a preview file (DO NOT UPLOAD):")
	cmd, precmd, err := ffmpegCommand(*dumpAudio, *dumpVideo, "video-preview.mp4", "")
	if err != nil {
		return err
	}
	log.Infof("  %v%v", precmd, printCommand(cmd))
	if *dumpVideo != "" {
		if *screenFilter != "linear2xcrt" {
			log.Infof("with current settings (1080p, MEDIUM QUALITY):")
			cmd, precmd, err := ffmpegCommand(*dumpAudio, *dumpVideo, "video-medium.mp4", *screenFilter)
			if err != nil {
				return err
			}
			log.Infof("  %v%v", precmd, printCommand(cmd))
		}
		log.Infof("preferred for uploading (4K, GOOD QUALITY):")
		cmd, precmd, err := ffmpegCommand(*dumpAudio, *dumpVideo, "video-high.mp4", "linear2xcrt")
		if err != nil {
			return err
		}
		log.Infof("  %v%v", precmd, printCommand(cmd))
	}
	return nil
}
