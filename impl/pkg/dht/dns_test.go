package dht

import (
	"context"
	"testing"

	"github.com/TBD54566975/ssi-sdk/crypto"
	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	"github.com/TBD54566975/ssi-sdk/cryptosuite"
	didsdk "github.com/TBD54566975/ssi-sdk/did"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/internal/did"
	"github.com/TBD54566975/did-dht-method/internal/util"
)

func TestGetPutDNSDHT(t *testing.T) {
	dht := NewTestDHT(t)
	defer dht.Close()

	_, privKey, err := util.GenerateKeypair()
	require.NoError(t, err)

	txtRecord := dns.TXT{
		Hdr: dns.RR_Header{
			Name:   "_did.",
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    7200,
		},
		Txt: []string{
			"hello mainline",
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
	put, err := CreateDNSPublishRequest(privKey, msg)
	require.NoError(t, err)

	id, err := dht.Put(context.Background(), *put)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	got, err := dht.GetFull(context.Background(), id)
	require.NoError(t, err)
	require.NotEmpty(t, got)

	gotMsg, err := ParseDNSGetResponse(*got)
	require.NoError(t, err)
	require.NotEmpty(t, gotMsg.Answer)

	assert.Equal(t, txtRecord.Txt, gotMsg.Answer[0].(*dns.TXT).Txt)
}

func TestGetPutDIDDHT(t *testing.T) {
	dht := NewTestDHT(t)
	defer dht.Close()

	pubKey, _, err := crypto.GenerateSECP256k1Key()
	require.NoError(t, err)
	pubKeyJWK, err := jwx.PublicKeyToPublicKeyJWK(nil, pubKey)
	require.NoError(t, err)

	opts := did.CreateDIDDHTOpts{
		VerificationMethods: []did.VerificationMethod{
			{
				VerificationMethod: didsdk.VerificationMethod{
					ID:           "key1",
					Type:         cryptosuite.JSONWebKeyType,
					Controller:   "did:dht:123456789abcdefghi",
					PublicKeyJWK: pubKeyJWK,
				},
				Purposes: []didsdk.PublicKeyPurpose{didsdk.AssertionMethod, didsdk.CapabilityInvocation},
			},
		},
		Services: []didsdk.Service{
			{
				ID:              "vcs",
				Type:            "VerifiableCredentialService",
				ServiceEndpoint: "https://example.com/vc/",
			},
			{
				ID:              "hub",
				Type:            "MessagingService",
				ServiceEndpoint: "https://example.com/hub/",
			},
		},
	}
	privKey, doc, err := did.GenerateDIDDHT(opts)
	require.NoError(t, err)
	require.NotEmpty(t, privKey)
	require.NotEmpty(t, doc)

	didID := did.DHT(doc.ID)
	didDocPacket, err := didID.ToDNSPacket(*doc, nil, nil, nil)
	require.NoError(t, err)

	putReq, err := CreateDNSPublishRequest(privKey, *didDocPacket)
	require.NoError(t, err)

	gotID, err := dht.Put(context.Background(), *putReq)
	require.NoError(t, err)
	require.NotEmpty(t, gotID)

	got, err := dht.GetFull(context.Background(), gotID)
	require.NoError(t, err)
	require.NotEmpty(t, got)

	gotMsg, err := ParseDNSGetResponse(*got)
	require.NoError(t, err)
	require.NotEmpty(t, gotMsg.Answer)

	d := did.DHT("did:dht:" + gotID)
	gotDoc, err := d.FromDNSPacket(gotMsg)
	require.NoError(t, err)
	require.NotEmpty(t, gotDoc)
}
