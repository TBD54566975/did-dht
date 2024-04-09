package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/TBD54566975/did-dht-method/internal/did"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
)

var (
	ticker = time.NewTicker(time.Second * 30)
)

func main() {
	logrus.SetLevel(logrus.InfoLevel)
	if len(os.Args) < 2 {
		logrus.Fatal("must specify 1 argument (server URL)")
	}

	run(os.Args[1])
}

func run(server string) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	for {
		select {
		case <-ctx.Done():
			logrus.Info("shutting down")
			return

		case <-ticker.C:
			suffix, err := put(ctx, server)
			if err != nil {
				logrus.WithError(err).Error("error making PUT request")
				continue
			}

			if err = get(ctx, server, suffix); err != nil {
				logrus.WithError(err).Error("error making GET request")
				continue
			}
		}
	}
}

func put(ctx context.Context, server string) (string, error) {
	didID, reqData, err := generateDIDPutRequest()
	if err != nil {
		return "", err
	}

	suffix, err := did.DHT(didID).Suffix()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPut, server+"/"+suffix, bytes.NewReader(reqData))
	if err != nil {
		return "", err
	}

	if err = doRequest(ctx, req); err != nil {
		return "", err
	}

	return suffix, nil
}

func get(ctx context.Context, server string, suffix string) error {
	req, err := http.NewRequest(http.MethodGet, server+"/"+suffix, nil)
	if err != nil {
		return err
	}

	if err = doRequest(ctx, req); err != nil {
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

	bep44Put, err := dht.CreatePkarrPublishRequest(sk, *packet)
	if err != nil {
		return "", nil, err
	}

	// prepare request as sig:seq:v
	var seqBuf [8]byte
	binary.BigEndian.PutUint64(seqBuf[:], uint64(bep44Put.Seq))
	return doc.ID, append(bep44Put.Sig[:], append(seqBuf[:], bep44Put.V.([]byte)...)...), nil
}

func doRequest(ctx context.Context, req *http.Request) error {
	log := logrus.WithFields(logrus.Fields{
		"method": req.Method,
		"url":    req.URL,
	})

	ctx, done := context.WithTimeout(ctx, time.Second*10)
	defer done()

	req = req.WithContext(ctx)

	log.Debug("making request")

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.WithFields(logrus.Fields{
		"time":   time.Since(start).Round(time.Millisecond),
		"status": resp.Status,
	}).Info("finished making request")

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected %s: %s", resp.Status, string(body))
	}

	return nil
}
