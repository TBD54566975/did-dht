package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	for i := 0; i < 1000; i++ {
		log := logrus.WithField("i", i)

		wg.Add(1)
		go func() {
			putStart := time.Now()
			suffix, err := put()
			if err != nil {
				log = log.WithError(err)
			}
			log.WithField("time", time.Since(putStart)).Info("PUT request completed")
			if err != nil {
				return
			}

			getStart := time.Now()
			err = get(suffix)
			if err != nil {
				log = log.WithError(err)
			}
			log.WithField("time", time.Since(getStart)).Info("GET request completed")

			wg.Done()
		}()
	}

	wg.Wait()

	logrus.WithField("time", time.Since(programstart)).Info("concurrency test completed")
}

func put() (string, error) {
	didID, reqData, err := generateDIDPutRequest()
	if err != nil {
		return "", err
	}

	suffix, err := did.DHT(didID).Suffix()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPut, "http://diddht:8305/"+suffix, bytes.NewReader(reqData))
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected %s: %s", resp.Status, string(body))
	}

	return suffix, nil
}

func get(suffix string) error {
	resp, err := http.Get("http://diddht:8305/" + suffix)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
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
