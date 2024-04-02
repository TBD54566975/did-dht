package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/TBD54566975/did-dht-method/internal/did"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
)

var (
	iterationsPerServer = 1000
	servers             = []string{"diddht-a", "diddht-b"}
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	programStart := time.Now()

	var wg sync.WaitGroup
	for _, server := range servers {
		for i := 0; i < iterationsPerServer; i++ {
			log := logrus.WithField("server", server).WithField("i", i)

			s := server
			wg.Add(1)
			go func() {
				putStart := time.Now()
				suffix, err := put(s)
				if err != nil {
					log = log.WithError(err)
				}
				log.WithField("time", time.Since(putStart)).Info("PUT request completed")
				if err != nil {
					return
				}

				getStart := time.Now()
				if err = get(s, suffix); err != nil {
					log = log.WithError(err)
				}
				log.WithField("time", time.Since(getStart)).Info("GET request completed")

				wg.Done()
			}()
		}
	}

	wg.Wait()

	logrus.WithField("time", time.Since(programStart)).Info("concurrency test completed")
}

func put(server string) (string, error) {
	didID, reqData, err := generateDIDPutRequest()
	if err != nil {
		return "", err
	}

	suffix, err := did.DHT(didID).Suffix()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPut, "http://"+server+":8305/"+suffix, bytes.NewReader(reqData))
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

func get(server, suffix string) error {
	resp, err := http.Get("http://" + server + ":8305/" + suffix)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err = io.ReadAll(resp.Body); err != nil {
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

	packet, err := did.DHT(doc.ID).ToDNSPacket(*doc, nil, nil)
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
