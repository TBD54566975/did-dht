package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/TBD54566975/did-dht-method/internal/did"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	programstart := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		log := logrus.WithField("i", i)

		wg.Add(1)
		go func() {
			log.Info("starting request")
			start := time.Now()
			putdid()
			log.WithField("time", time.Since(start)).Info("request completed")
			wg.Done()
		}()
	}

	wg.Wait()

	logrus.WithField("time", time.Since(programstart)).Info("concurrency test completed")
}

func putdid() {
	didID, reqData, err := generateDIDPutRequest()
	if err != nil {
		logrus.WithError(err).Fatal("error generating DID for PUT request")
	}

	suffix, err := did.DHT(didID).Suffix()
	if err != nil {
		logrus.WithError(err).Fatal("error parsing generated did")
	}

	req, err := http.NewRequest(http.MethodPut, "http://diddht:8305/"+suffix, bytes.NewReader(reqData))
	if err != nil {
		logrus.WithError(err).Fatal("error preparing PUT request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.WithError(err).Fatal("error making request to server")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Fatal("error reading response from server")
	}

	if resp.StatusCode != 200 {
		logrus.WithFields(logrus.Fields{
			"status": resp.Status,
			"body":   string(body),
		}).Warn("unexpected non-200 response code from PUT request")
		return
	}
}

func generateDIDPutRequest() (string, []byte, error) {
	// generate a DID Document
	sk, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
	if err != nil {
		return "", nil, err
	}

	packet, err := did.DHT(doc.ID).ToDNSPacket(*doc, nil)
	if err != nil {
		return "", nil, err
	}

	bep44Put, err := dht.CreatePKARRPublishRequest(sk, *packet)
	if err != nil {
		return "", nil, err
	}

	// prepare request as sig:seq:v
	var seqBuf [8]byte
	binary.BigEndian.PutUint64(seqBuf[:], uint64(bep44Put.Seq))
	return doc.ID, append(bep44Put.Sig[:], append(seqBuf[:], bep44Put.V.([]byte)...)...), nil
}
