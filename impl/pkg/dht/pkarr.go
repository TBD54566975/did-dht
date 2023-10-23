package dht

import (
	"crypto/ed25519"
	"time"

	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/anacrolix/dht/v2/bep44"
	"github.com/anacrolix/dht/v2/exts/getput"
	"github.com/anacrolix/torrent/bencode"
	"github.com/miekg/dns"
)

// CreatePKARRPublishRequest creates a put request for the given records. Requires a public/private keypair and the records to put.
// The records are expected to be a DNS message packet, such as:
//
//	dns.Msg{
//			MsgHdr: dns.MsgHdr{
//				Id:            0,
//				Response:      true,
//				Authoritative: true,
//			},
//		   Answer: dns.RR{
//			dns.TXT{
//				Hdr: dns.RR_Header{
//					Name:   "_did.",
//					Rrtype: dns.TypeTXT,
//					Class:  dns.ClassINET,
//					Ttl:    7200,
//				},
//				Txt: []string{
//					"hello pkarr",
//				},
//		    }
//		}
func CreatePKARRPublishRequest(privateKey ed25519.PrivateKey, msg dns.Msg) (*bep44.Put, error) {
	packed, err := msg.Pack()
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to pack records")
	}
	publicKey := privateKey.Public().(ed25519.PublicKey)
	put := &bep44.Put{
		V:   packed,
		K:   (*[32]byte)(publicKey),
		Seq: time.Now().UnixMilli() / 1000,
	}
	put.Sign(privateKey)
	return put, nil
}

// ParsePKARRGetResponse parses the response from a get request.
// The response is expected to be a slice of DNS resource records.
func ParsePKARRGetResponse(response getput.GetResult) (*dns.Msg, error) {
	var payload string
	if err := bencode.Unmarshal(response.V, &payload); err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to unmarshal payload value")
	}
	msg := new(dns.Msg)
	if err := msg.Unpack([]byte(payload)); err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to unpack records")
	}
	return msg, nil
}
