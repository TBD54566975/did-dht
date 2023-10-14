package dht

import (
	"crypto/ed25519"
	"time"

	"github.com/anacrolix/dht/v2/bep44"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

// CreatePKARRPutRequest creates a put request for the given records. Requires a public/private keypair and the records to put.
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
func CreatePKARRPutRequest(publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey, msg dns.Msg) (*bep44.Put, error) {
	packed, err := msg.Pack()
	if err != nil {
		logrus.WithError(err).Error("failed to pack records")
		return nil, err
	}
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
func ParsePKARRGetResponse(response []byte) ([]dns.RR, error) {
	msg := new(dns.Msg)
	if err := msg.Unpack(response); err != nil {
		logrus.WithError(err).Error("failed to unpack records")
		return nil, err
	}
	return msg.Answer, nil
}
