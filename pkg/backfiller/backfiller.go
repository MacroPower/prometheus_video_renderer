// Backfiller assists with backfilling via promtool.
package backfiller

import (
	"bytes"
	"fmt"
	"time"

	"github.com/youpy/go-wav"
)

func Help(metric string) string {
	return fmt.Sprintf("# HELP %s The metric.\n# TYPE %s gauge\n", metric, metric)
}

func WriteHelp(b *bytes.Buffer, metric ...string) {
	if b.Len() == 0 {
		for _, m := range metric {
			b.WriteString(fmt.Sprintf("# HELP %s The metric.\n# TYPE %s gauge\n", m, m))
		}
	}
}

func FromUnixMs(ms int64) time.Time {
	return time.Unix(ms/int64(1000), (ms%int64(1000))*int64(1000000))
}

func WriteWaveSample(b *bytes.Buffer, t time.Time, trackName string, sample wav.Sample, value int, format *wav.WavFormat) {
	WriteHelp(b, trackName)

	b.WriteString(
		fmt.Sprintf(
			`%s{audio_format="%d",bits_per_sample="%d",block_align="%d",byte_rate="%d",sample_rate="%d"} %d %d%s`,
			trackName,
			format.AudioFormat,
			format.BitsPerSample,
			format.BlockAlign,
			format.ByteRate,
			format.SampleRate,
			value,
			t.Unix(),
			"\n",
		),
	)
}

func WriteEnd(b *bytes.Buffer) {
	b.WriteString("# EOF")
}
