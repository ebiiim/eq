package filter

import (
	"encoding/binary"
	"fmt"
)

type SoXOption string

const (
	AppNameDefault                        SoXOption = "sox"
	FmtRAW, FmtFLAC, FmtMP3, FmtWAV       SoXOption = "raw", "flac", "mp3", "wav"
	ChMono, ChStereo, Ch21, Ch51, Ch71    SoXOption = "1", "2", "3", "6", "8"
	Rate44100, Rate48k, Rate96k, Rate192k SoXOption = "44100", "48000", "96000", "192000"
	Bit16, Bit24, Bit32                   SoXOption = "16", "24", "32"
	EncSigned, EncUnsigned, EncFloat      SoXOption = "signed-integer", "unsigned-integer", "floating-point"
	EndianBig, EndianLittle               SoXOption = "big", "little"
)

func SoXCmd(execPath string, bufferSize int,
	inFmt, inCh, inRate, inBit, inEnc SoXOption, inByteOrder binary.ByteOrder,
	outFmt, outCh, outRate, outBit, outEnc SoXOption, outByteOrder binary.ByteOrder,
	effects ...SoXEffect,
) string {
	cmdInByteOrder := EndianLittle
	cmdOutByteOrder := EndianLittle
	if inByteOrder == binary.BigEndian {
		cmdInByteOrder = EndianBig
	}
	if outByteOrder == binary.BigEndian {
		cmdOutByteOrder = EndianBig
	}
	cmdIn := fmt.Sprintf("-t%s -b%s -r%s -c%s -e%s --endian %s -", inFmt, inBit, inRate, inCh, inEnc, cmdInByteOrder)
	cmdOut := fmt.Sprintf("-t%s -b%s -r%s -c%s -e%s --endian %s -", outFmt, outBit, outRate, outCh, outEnc, cmdOutByteOrder)
	cmdStr := fmt.Sprintf("%s %s %s --buffer %d -V0", execPath, cmdIn, cmdOut, bufferSize)

	for _, e := range effects {
		cmdStr += " " + string(e)
	}
	return cmdStr
}

type SoXEffect string

func NewSoXGain(gain float64) SoXEffect {
	return SoXEffect(fmt.Sprintf("gain %.3f", gain))
}

func NewSoXEqualizer(q float64, gain float64) SoXEffect {
	return SoXEffect(fmt.Sprintf("equalizer %.3f %.3f", q, gain))
}
