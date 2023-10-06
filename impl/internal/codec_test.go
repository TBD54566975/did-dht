package internal

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecode(t *testing.T) {
	t.Run("uncompressable data", func(t *testing.T) {
		uncompressable := []any{[]int{3}, []string{"3"}}

		uncompressableBytes, err := json.Marshal(uncompressable)
		require.NoError(t, err)

		encoded, err := Encode(uncompressableBytes)
		require.NoError(t, err)
		assert.NotEmpty(t, encoded)

		assert.EqualValues(t, 0, encoded[0], "version 0")
		assert.EqualValues(t, "000b05805b5b335d2c5b2233225d5d03", Hex(encoded))

		decoded, err := Decode(encoded)
		require.NoError(t, err)
		assert.NotEmpty(t, decoded)

		assert.JSONEq(t, string(uncompressableBytes), string(decoded))
	})

	t.Run("compressable data", func(t *testing.T) {
		compressable := []any{
			[]any{"cloudflare.com.", 146, "IN", "A", "104.16.132.229"},
			[]any{"cloudflare.com.", 146, "IN", "A", "104.16.133.229"},
			[]any{},
		}

		compressableBytes, err := json.Marshal(compressable)
		require.NoError(t, err)

		encoded, err := Encode(compressableBytes)
		require.NoError(t, err)
		assert.NotEmpty(t, encoded)

		assert.EqualValues(t, 0, encoded[0], "version 0")
		assert.EqualValues(t, "001b6700f845e796faf3c704189bd16831144730d84cc2061c58e090278bc319042821b5379e63effc9b0712502389b9400c10da423445ddc4aca308b70a", Hex(encoded))

		decoded, err := Decode(encoded)
		require.NoError(t, err)
		assert.NotEmpty(t, decoded)

		assert.JSONEq(t, string(compressableBytes), string(decoded))
	})
}
