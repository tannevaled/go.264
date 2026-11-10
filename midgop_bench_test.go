package go264

import (
	"os"
	"testing"

	"github.com/oops1/go.264/internal/nal"
)

// BenchmarkDecodeMidGOP decodes the IDR picture and the 45 pictures that
// follow it from the Annex B file named by GO264_MIDGOP_CLIP (for example a
// 1280x720 x264 clip, GOP 60), the work a poster frame at 1.5 s into a GOP
// costs. The stream is fed one access unit per Decode call, as a container
// reader hands it over. Skipped when the variable is unset.
func BenchmarkDecodeMidGOP(b *testing.B) {
	path := os.Getenv("GO264_MIDGOP_CLIP")
	if path == "" {
		b.Skip("GO264_MIDGOP_CLIP not set")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatal(err)
	}
	var units [][]byte
	var cur []byte
	slices := 0
	for _, n := range nal.SplitAnnexB(data) {
		if len(n) == 0 {
			continue
		}
		typ := n[0] & 0x1f
		if typ == 1 || typ == 5 {
			if slices == 46 {
				break
			}
			slices++
			if slices > 1 {
				units = append(units, cur)
				cur = nil
			}
		}
		cur = append(cur, 0, 0, 0, 1)
		cur = append(cur, n...)
	}
	units = append(units, cur)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoderWithConfig(DecoderConfig{ForceSoftware: true})
		for _, u := range units {
			if _, err := dec.Decode(u); err != nil {
				b.Fatalf("Decode: %v", err)
			}
		}
		if _, err := dec.Flush(); err != nil {
			b.Fatalf("Flush: %v", err)
		}
		dec.Close()
	}
}
