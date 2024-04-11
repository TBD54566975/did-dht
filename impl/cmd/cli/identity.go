package main

import (
	"context"
	"fmt"

	"github.com/goccy/go-json"

	"github.com/miekg/dns"
	"github.com/mr-tron/base58"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/internal"
	"github.com/TBD54566975/did-dht-method/internal/cli"
	"github.com/TBD54566975/did-dht-method/internal/util"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
)

func init() {
	rootCmd.AddCommand(identityCmd)
	identityCmd.AddCommand(identityAddCmd)
	identityCmd.AddCommand(identityGetCmd)
}

var identityCmd = &cobra.Command{
	Use:   "id",
	Short: "Manage identities",
	RunE: func(cmd *cobra.Command, args []string) error {
		identities, err := cli.Read()
		if err != nil {
			return err
		}
		if len(identities) == 0 {
			println("No identities found.")
			return nil
		}
		i := 0
		for id, identity := range identities {
			recordJSON, err := json.Marshal(identity.Records)
			if err != nil {
				logrus.WithError(err).Error("failed to marshal records")
				return err
			}
			fmt.Printf("%d: %s – %s\n", i+1, id, string(recordJSON))
			i++
		}
		return nil
	},
}

var identityAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an identity, accepting a json string of DNS TXT records",
	Long:  `Add an identity, accepting a json string of DNS TXT records such as [["_did", 7200, "did:example:1234"]].`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pubKey, privKey, err := util.GenerateKeypair()
		if err != nil {
			logrus.WithError(err).Error("failed to generate keypair")
			return err
		}
		var records [][]any
		if err := json.Unmarshal([]byte(args[0]), &records); err != nil {
			logrus.WithError(err).Error("failed to unmarshal records")
			return err
		}

		// start dht
		d, err := dht.NewDHT(config.GetDefaultBootstrapPeers())
		if err != nil {
			logrus.WithError(err).Error("failed to create dht")
			return err
		}

		var rrds []dns.RR
		for _, record := range records {
			if len(record) != 3 {
				logrus.WithError(err).Error("invalid record")
				return err
			}
			rr := dns.TXT{
				Hdr: dns.RR_Header{
					Name:   record[0].(string),
					Rrtype: dns.TypeTXT,
					Class:  dns.ClassINET,
					Ttl:    uint32(record[1].(float64)),
				},
				Txt: []string{
					record[2].(string),
				},
			}

			rrds = append(rrds, &rr)
		}
		msg := dns.Msg{
			MsgHdr: dns.MsgHdr{
				Id:            0,
				Response:      true,
				Authoritative: true,
			},
			Answer: rrds,
		}
		// generate put request
		putReq, err := dht.CreatePkarrPublishRequest(privKey, msg)
		if err != nil {
			logrus.WithError(err).Error("failed to create put request")
			return err
		}

		// put the identity into the dht
		id, err := d.Put(context.Background(), *putReq)
		if err != nil {
			logrus.WithError(err).Error("failed to put identity into dht")
			return err
		}

		// write the identity to the diddht file
		identity := internal.Identity{
			Base58PublicKey:  base58.Encode(pubKey),
			Base58PrivateKey: base58.Encode(privKey),
			Records:          records,
		}
		if err := cli.Write(id, identity); err != nil {
			logrus.WithError(err).Error("failed to write identity to diddht file")
			return err
		}

		fmt.Printf("Added identity: %s, with records: %s\n", id, args[0])
		return nil
	},
}

var identityGetCmd = &cobra.Command{
	Use:   "get",
	Short: "GetRecord an identity",
	Long:  `GetRecord an identity by its id.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		// first read the diddht file
		identities, err := cli.Read()
		if err == nil {
			// if the diddht file exists, look for the identity
			if identity, ok := identities[id]; ok {
				recordsBytes, err := json.Marshal(identity.Records)
				if err != nil {
					logrus.WithError(err).Error("failed to marshal records")
					return err
				}
				fmt.Printf("Records: %s\n", string(recordsBytes))
				return nil
			}
		}
		// fall back to dht if not found in diddht file

		// start dht
		d, err := dht.NewDHT(config.GetDefaultBootstrapPeers())
		if err != nil {
			logrus.WithError(err).Error("failed to create dht")
			return err
		}

		// get the identity from the dht
		gotResp, err := d.GetFull(context.Background(), id)
		if err != nil {
			logrus.WithError(err).Error("failed to get identity from dht")
			return err
		}

		msg, err := dht.ParsePkarrGetResponse(*gotResp)
		if err != nil {
			logrus.WithError(err).Error("failed to parse get response")
			return err
		}

		for _, rr := range msg.Answer {
			fmt.Printf("%s\n", rr.String())
		}

		return nil
	},
}
