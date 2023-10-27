package dht

import (
	"context"
	"testing"

	"github.com/TBD54566975/ssi-sdk/crypto"
	"github.com/TBD54566975/ssi-sdk/crypto/jwx"
	didsdk "github.com/TBD54566975/ssi-sdk/did"
	"github.com/TBD54566975/ssi-sdk/did/ion"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/internal/did"
	"github.com/TBD54566975/did-dht-method/internal/util"
)

func TestGetPutPKARRDHT(t *testing.T) {
	d, err := NewDHT(config.GetDefaultBootstrapPeers())
	require.NoError(t, err)

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
	put, err := CreatePKARRPublishRequest(privKey, msg)
	require.NoError(t, err)

	id, err := d.Put(context.Background(), *put)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	got, err := d.Get(context.Background(), id)
	require.NoError(t, err)
	require.NotEmpty(t, got)

	gotMsg, err := ParsePKARRGetResponse(*got)
	require.NoError(t, err)
	require.NotEmpty(t, gotMsg.Answer)

	assert.Equal(t, txtRecord.Txt, gotMsg.Answer[0].(*dns.TXT).Txt)
}

func TestGetPutDIDDHT(t *testing.T) {
	dht, err := NewDHT(config.GetDefaultBootstrapPeers())
	require.NoError(t, err)

	pubKey, _, err := crypto.GenerateSECP256k1Key()
	require.NoError(t, err)
	pubKeyJWK, err := jwx.PublicKeyToPublicKeyJWK("key1", pubKey)
	require.NoError(t, err)

	opts := did.CreateDIDDHTOpts{
		VerificationMethods: []did.VerificationMethod{
			{
				VerificationMethod: didsdk.VerificationMethod{
					ID:           "key1",
					Type:         "JsonWebKey2020",
					Controller:   "did:dht:123456789abcdefghi",
					PublicKeyJWK: pubKeyJWK,
				},
				Purposes: []ion.PublicKeyPurpose{ion.AssertionMethod, ion.CapabilityInvocation},
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
	didDocPacket, err := didID.ToDNSPacket(*doc, nil)
	require.NoError(t, err)

	putReq, err := CreatePKARRPublishRequest(privKey, *didDocPacket)
	require.NoError(t, err)

	gotID, err := dht.Put(context.Background(), *putReq)
	require.NoError(t, err)
	require.NotEmpty(t, gotID)

	got, err := dht.Get(context.Background(), gotID)
	require.NoError(t, err)
	require.NotEmpty(t, got)

	gotMsg, err := ParsePKARRGetResponse(*got)
	require.NoError(t, err)
	require.NotEmpty(t, gotMsg.Answer)

	d := did.DHT("did:dht:" + gotID)
	gotDoc, _, err := d.FromDNSPacket(gotMsg)
	require.NoError(t, err)
	require.NotEmpty(t, gotDoc)
}
