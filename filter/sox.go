package filter

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/pkg/errors"
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

type SoXOptions struct {
	app                                    string
	inFmt, inCh, inRate, inBit, inEnc      SoXOption
	outFmt, outCh, outRate, outBit, outEnc SoXOption
	inByteOrder, outByteOrder              SoXOption
}

func NewSoXOptions(app string,
	inFmt, inCh, inRate, inBit, inEnc SoXOption, inByteOrder binary.ByteOrder,
	outFmt, outCh, outRate, outBit, outEnc SoXOption, outByteOrder binary.ByteOrder,
) *SoXOptions {
	var so SoXOptions
	so.app = app
	so.inFmt, so.inCh, so.inRate, so.inBit, so.inEnc = inFmt, inCh, inRate, inBit, inEnc
	so.outFmt, so.outCh, so.outRate, so.outBit, so.outEnc = outFmt, outCh, outRate, outBit, outEnc
	so.inByteOrder, so.outByteOrder = EndianLittle, EndianLittle
	if inByteOrder == binary.BigEndian {
		so.inByteOrder = EndianBig
	}
	if outByteOrder == binary.BigEndian {
		so.inByteOrder = EndianBig
	}
	return &so
}

type SoXEffect []string

func NewSoXGain(gain float64) SoXEffect {
	return []string{"gain", fmt.Sprintf("%.3f", gain)}
}

func NewSoXEqualizer(q float64, gain float64) SoXEffect {
	return []string{"equalizer", fmt.Sprintf("%.3f", q), fmt.Sprintf("%.3f", gain)}
}

type SoXFilter struct {
	initOnce  sync.Once
	mu        sync.Mutex
	options   *SoXOptions
	effects   []SoXEffect
	parsedCmd []string
}

func (f *SoXFilter) parseCmd() {
	f.mu.Lock()
	opts := f.options
	effs := f.effects
	f.mu.Unlock()

	// TODO: endian
	cmdIn := fmt.Sprintf("-t%s -b%s -r%s -c%s -e%s --endian %s -", opts.inFmt, opts.inBit, opts.inRate, opts.inCh, opts.inEnc, opts.inByteOrder)
	cmdOut := fmt.Sprintf("-t%s -b%s -r%s -c%s -e%s --endian %s -", opts.outFmt, opts.outBit, opts.outRate, opts.outCh, opts.outEnc, opts.outByteOrder)
	cmdStr := fmt.Sprintf("%s %s %s", opts.app, cmdIn, cmdOut)
	cmd := strings.Fields(cmdStr)
	for _, e := range effs {
		cmd = append(cmd, []string(e)...)
	}

	f.mu.Lock()
	f.parsedCmd = cmd
	f.mu.Unlock()
}

func (f *SoXFilter) SetOptions(options *SoXOptions) error {
	if !func(opts *SoXOptions) bool {
		// TODO: validation
		return true
	}(options) {
		return fmt.Errorf("invalid options %v", options)
	}

	f.mu.Lock()
	f.options = options
	f.mu.Unlock()

	f.parseCmd()
	return nil
}

func (f *SoXFilter) SetEffects(effects []SoXEffect) error {
	if !func(effs []SoXEffect) bool {
		// TODO: validation
		return true
	}(effects) {
		return fmt.Errorf("invalid effects %v", effects)
	}

	f.mu.Lock()
	f.effects = effects
	f.mu.Unlock()

	f.parseCmd()
	return nil
}

func (f *SoXFilter) Filter(reader io.Reader) (io.Reader, error) {
	f.initOnce.Do(func() {
		so := NewSoXOptions(string(AppNameDefault),
			FmtRAW, ChStereo, Rate48k, Bit16, EncSigned, binary.LittleEndian, // in
			FmtRAW, ChStereo, Rate48k, Bit16, EncSigned, binary.LittleEndian, // out
		)
		es := []SoXEffect{
			NewSoXGain(-3),
		}
		_ = f.SetOptions(so)
		_ = f.SetEffects(es)
	}) // safe

	f.mu.Lock()
	newCmd := f.parsedCmd
	f.mu.Unlock()

	cmd := exec.Command(newCmd[0], newCmd[1:]...)

	cmd.Stdin = reader
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	err := cmd.Run()
	if err != nil {
		return nil, errors.Wrap(err, "failed to run cmd")
	}

	return &outBuf, nil
}
