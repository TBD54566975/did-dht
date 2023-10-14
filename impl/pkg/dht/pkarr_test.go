package dht

import (
	"context"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/internal/util"
)

func TestGetPutPKARRDHT(t *testing.T) {
	d, err := NewDHT()
	require.NoError(t, err)

	pubKey, privKey, err := util.GenerateKeypair()
	require.NoError(t, err)

	txtRecord := dns.TXT{
		Hdr: dns.RR_Header{
			Name:   "_did.",
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    7200,
		},
		Txt: []string{
			"hello pkarr",
		},
	}
	msg := dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:            0,
			Response:      true,
			Authoritative: true,
		},
		Answer: []dns.RR{&txtRecord},
	}
	put, err := CreatePKARRPutRequest(pubKey, privKey, msg)
	require.NoError(t, err)

	id, err := d.Put(context.Background(), pubKey, *put)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	got, err := d.Get(context.Background(), id)
	require.NoError(t, err)
	require.NotEmpty(t, got)

	gotRRs, err := ParsePKARRGetResponse(got)
	require.NoError(t, err)
	require.NotEmpty(t, gotRRs)

	assert.Equal(t, txtRecord.Txt, gotRRs[0].(*dns.TXT).Txt)
}
