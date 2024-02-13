package pkarr_test

import (
	"strings"
	"testing"

	"github.com/TBD54566975/did-dht-method/internal/did"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
	"github.com/TBD54566975/did-dht-method/pkg/pkarr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRecord(t *testing.T) {
	// validate incorrect key length is rejected
	r, err := pkarr.NewRecord([]byte("aaaaaaaaaaa"), nil, nil, 0)
	assert.EqualError(t, err, "incorrect key length for pkarr record")
	assert.Nil(t, r)

	// create a did doc as a packet to store
	sk, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
	require.NoError(t, err)
	require.NotEmpty(t, doc)

	packet, err := did.DHT(doc.ID).ToDNSPacket(*doc, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, packet)

	putMsg, err := dht.CreatePKARRPublishRequest(sk, *packet)
	require.NoError(t, err)
	require.NotEmpty(t, putMsg)

	r, err = pkarr.NewRecord(putMsg.K[:], []byte(strings.Repeat("a", 1001)), putMsg.Sig[:], putMsg.Seq)
	assert.EqualError(t, err, "pkarr record value too long")
	assert.Nil(t, r)

	r, err = pkarr.NewRecord(putMsg.K[:], putMsg.V.([]byte), []byte(strings.Repeat("a", 65)), putMsg.Seq)
	assert.EqualError(t, err, "incorrect sig length for pkarr record")
	assert.Nil(t, r)

	r, err = pkarr.NewRecord(putMsg.K[:], putMsg.V.([]byte), putMsg.Sig[:], 0)
	assert.EqualError(t, err, "Key: 'Record.SequenceNumber' Error:Field validation for 'SequenceNumber' failed on the 'required' tag")
	assert.Nil(t, r)

	r, err = pkarr.NewRecord(putMsg.K[:], putMsg.V.([]byte), putMsg.Sig[:], 1)
	assert.EqualError(t, err, "signature is invalid")
	assert.Nil(t, r)
}
