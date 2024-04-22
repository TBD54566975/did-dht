package dht_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/internal/did"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
)

func TestNewRecord(t *testing.T) {
	// validate incorrect key length is rejected
	r, err := dht.NewBEP44Record([]byte("aaaaaaaaaaa"), nil, nil, 0)
	assert.EqualError(t, err, "incorrect key length for bep44 record")
	assert.Nil(t, r)

	// create a did doc as a packet to store
	sk, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
	require.NoError(t, err)
	require.NotEmpty(t, doc)

	packet, err := did.DHT(doc.ID).ToDNSPacket(*doc, nil, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, packet)

	putMsg, err := dht.CreateDNSPublishRequest(sk, *packet)
	require.NoError(t, err)
	require.NotEmpty(t, putMsg)

	r, err = dht.NewBEP44Record(putMsg.K[:], []byte(strings.Repeat("a", 1001)), putMsg.Sig[:], putMsg.Seq)
	assert.EqualError(t, err, "bep44 record value too long")
	assert.Nil(t, r)

	r, err = dht.NewBEP44Record(putMsg.K[:], putMsg.V.([]byte), []byte(strings.Repeat("a", 65)), putMsg.Seq)
	assert.EqualError(t, err, "incorrect sig length for bep44 record")
	assert.Nil(t, r)

	r, err = dht.NewBEP44Record(putMsg.K[:], putMsg.V.([]byte), putMsg.Sig[:], 0)
	assert.EqualError(t, err, "Key: 'BEP44Record.SequenceNumber' Error:Field validation for 'SequenceNumber' failed on the 'required' tag")
	assert.Nil(t, r)

	r, err = dht.NewBEP44Record(putMsg.K[:], putMsg.V.([]byte), putMsg.Sig[:], 1)
	assert.EqualError(t, err, "signature is invalid")
	assert.Nil(t, r)

	r, err = dht.NewBEP44Record(putMsg.K[:], putMsg.V.([]byte), putMsg.Sig[:], putMsg.Seq)
	assert.NoError(t, err)

	bep := r.Put()
	assert.Equal(t, putMsg.K, bep.K)
	assert.Equal(t, putMsg.V, bep.V)
	assert.Equal(t, putMsg.Sig, bep.Sig)
	assert.Equal(t, putMsg.Seq, bep.Seq)

	resp := r.Response()
	assert.Equal(t, r.Value, resp.V)
	assert.Equal(t, r.SequenceNumber, resp.Seq)
	assert.Equal(t, r.Signature, resp.Sig)

	r2 := dht.RecordFromBEP44(putMsg)
	assert.Equal(t, r.Key, r2.Key)
	assert.Equal(t, r.Value, r2.Value)
	assert.Equal(t, r.Signature, r2.Signature)
	assert.Equal(t, r.SequenceNumber, r2.SequenceNumber)
}
