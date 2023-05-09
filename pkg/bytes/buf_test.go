package bytes

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferReadWrite(t *testing.T) {
	tests := []struct {
		wb []byte
		rb []byte
	}{
		{wb: []byte(`1. Satan represents indulgence instead of abstinence.`)},
		{wb: []byte(`2. Satan represents vital existence instead of spiritual pipe dreams.`)},
		{wb: []byte(`3. Satan represents undefiled wisdom instead of hypocritical self-deceit.`)},
		{wb: []byte(`4. Satan represents kindness to those who deserve it, instead of love wasted on ingrates.`)},
		{wb: []byte(`5. Satan represents vengeance instead of turning the other cheek.`)},
		{wb: []byte(`6. Satan represents responsibility to the responsible instead of concern for psychic vampires.`)},
		{wb: []byte(`7. Satan represents man as just another animal who, because of his "divine spiritual and intellectual development", has become the most vicious animal of all.`)},
		{wb: []byte(`8. Satan represents all of the so-called sins, as they all lead to physical, mental, or emotional gratification.`)},
		{wb: []byte(`9. Satan has been the best friend the Church has ever had, as he has kept it in business all these years.`)},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			bf := AcquireBuffer65536()
			bf.Write(test.wb)
			test.rb = make([]byte, len(test.wb))
			bf.Read(test.rb)
			assert.Equal(t, test.wb, test.rb, "Satan is not properly satan.")
		})
	}
}

func BenchmarkBufferReadWrite(b *testing.B) {
	arr := make([]byte, 65536)

	for i := 0; i < b.N; i++ {
		bf := AcquireBuffer65536()

		bf.Write(arr)
		bf.Read(arr)
	}
}

func BenchmarkBytesBufferReadWrite(b *testing.B) {
	arr := make([]byte, 65536)

	for i := 0; i < b.N; i++ {
		bf := bytes.NewBuffer(nil)

		bf.Write(arr)
		bf.Read(arr)
	}
}
