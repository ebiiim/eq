package sox

import (
	"fmt"
	"sync"
)

type Option string

const (
	FmtRAW, FmtFLAC, FmtMP3, FmtWAV       Option = "raw", "flac", "mp3", "wav"
	Mono, Stereo, Ch21, Ch51, Ch71        Option = "1", "2", "3", "6", "8"
	Rate44100, Rate48k, Rate96k, Rate192k Option = "44100", "48000", "96000", "192000"
	Bit16, Bit24, Bit32                   Option = "16", "24", "32"
	EncSigned, EncUnsigned, EncFloat      Option = "signed", "unsigned", "floating"
	EndianBig, EndianLittle               Option = "-B", "-L"
)

type Command struct {
	initOnce                                                         sync.Once
	ExecPath                                                         string
	BufferSize                                                       int
	InFormat, InChannels, InRate, InBit, InEncode, InByteOrder       Option
	OutFormat, OutChannels, OutRate, OutBit, OutEncode, OutByteOrder Option
	Effects                                                          []Effect
}

func (s *Command) String() string {
	s.initOnce.Do(func() {
		if s.ExecPath == "" {
			s.ExecPath = "sox"
		}
		if s.BufferSize == 0 {
			s.BufferSize = 8192 // --buffer N (default: 8192)
		}
		if s.InFormat == "" {
			s.InFormat = FmtRAW
		}
		if s.InChannels == "" {
			s.InChannels = Stereo
		}
		if s.InRate == "" {
			s.InRate = Rate48k
		}
		if s.InBit == "" {
			s.InBit = Bit16
		}
		if s.InEncode == "" {
			s.InEncode = EncSigned
		}
		if s.InByteOrder == "" {
			s.InByteOrder = EndianLittle
		}
		if s.OutFormat == "" {
			s.OutFormat = FmtRAW
		}
		if s.OutChannels == "" {
			s.OutChannels = Stereo
		}
		if s.OutRate == "" {
			s.OutRate = Rate48k
		}
		if s.OutBit == "" {
			s.OutBit = Bit16
		}
		if s.OutEncode == "" {
			s.OutEncode = EncSigned
		}
		if s.OutByteOrder == "" {
			s.OutByteOrder = EndianLittle
		}
	})

	cmdIn := fmt.Sprintf("-t%s -b%s -r%s -c%s -e%s %s -", s.InFormat, s.InBit, s.InRate, s.InChannels, s.InEncode, s.InByteOrder)
	cmdOut := fmt.Sprintf("-t%s -b%s -r%s -c%s -e%s %s -", s.OutFormat, s.OutBit, s.OutRate, s.OutChannels, s.OutEncode, s.OutByteOrder)
	cmdStr := fmt.Sprintf("%s %s %s --buffer %d -V0", s.ExecPath, cmdIn, cmdOut, s.BufferSize)

	for _, e := range s.Effects {
		cmdStr += " " + string(e)
	}
	return cmdStr
}

type Effect string

func NewGain(gain float64) Effect {
	return Effect(fmt.Sprintf("gain %.3f", gain))
}

func NewEQ(freq uint, q float64, gain float64) Effect {
	return Effect(fmt.Sprintf("equalizer %d %.3f %.3f", freq, q, gain))
}