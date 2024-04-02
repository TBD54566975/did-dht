package did

import (
	"bytes"
	"encoding/binary"
	"io"
	"net/http"
	"net/url"

	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/anacrolix/dht/v2/bep44"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
)

// GatewayClient is the client for the Gateway API
type GatewayClient struct {
	gatewayURL string
	client     *http.Client
}

// NewGatewayClient returns a new instance of the Gateway client
func NewGatewayClient(gatewayURL string) (*GatewayClient, error) {
	if _, err := url.Parse(gatewayURL); err != nil {
		return nil, err
	}
	return &GatewayClient{
		gatewayURL: gatewayURL,
		client:     http.DefaultClient,
	}, nil
}

// GetDIDDocument gets a DID document, its types, and authoritative gateways, from a did:dht Gateway
func (c *GatewayClient) GetDIDDocument(id string) (*did.Document, []TypeIndex, []AuthoritativeGateway, error) {
	d := DHT(id)
	if !d.IsValid() {
		return nil, nil, nil, errors.New("invalid did")
	}
	suffix, err := d.Suffix()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to get suffix")
	}
	resp, err := http.Get(c.gatewayURL + "/" + suffix)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to get did document")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil, nil, errors.Errorf("failed to get did document, status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to read response body")
	}
	msg := new(dns.Msg)
	if err = msg.Unpack(body[72:]); err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to unpack records")
	}
	return d.FromDNSPacket(msg)
}

// PutDocument puts a bep44.Put message to a did:dht Gateway
func (c *GatewayClient) PutDocument(id string, put bep44.Put) error {
	d := DHT(id)
	if !d.IsValid() {
		return errors.New("invalid did")
	}
	suffix, err := d.Suffix()
	if err != nil {
		return errors.Wrap(err, "failed to get suffix")
	}

	// prepare request as sig:seq:v
	var seqBuf [8]byte
	binary.BigEndian.PutUint64(seqBuf[:], uint64(put.Seq))
	reqBytes := append(put.Sig[:], append(seqBuf[:], put.V.([]byte)...)...)

	req, err := http.NewRequest(http.MethodPut, c.gatewayURL+"/"+suffix, bytes.NewReader(reqBytes))
	if err != nil {
		return errors.Wrap(err, "could not construct http put request")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "could not put document")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("unsuccessful")
	}
	return nil
}
