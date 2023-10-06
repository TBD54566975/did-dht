package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mr-tron/base58"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

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
	Short: "Add an identity, accepting a json string of records",
	Long:  `Add an identity, accepting a json string of records such as [["foo", "bar"]].`,
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
		d, err := dht.NewDHT()
		if err != nil {
			logrus.WithError(err).Error("failed to create dht")
			return err
		}

		// generate put request
		putReq, err := dht.CreatePutRequest(pubKey, privKey, records)
		if err != nil {
			logrus.WithError(err).Error("failed to create put request")
			return err
		}

		// put the identity into the dht
		id, err := d.Put(context.Background(), pubKey, *putReq)
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
	Short: "Get an identity",
	Long:  `Get an identity by its id.`,
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
		d, err := dht.NewDHT()
		if err != nil {
			logrus.WithError(err).Error("failed to create dht")
			return err
		}

		// get the identity from the dht
		gotJSON, err := d.Get(context.Background(), id)
		if err != nil {
			logrus.WithError(err).Error("failed to get identity from dht")
			return err
		}

		fmt.Printf("Records: %s\n", gotJSON)
		return nil
	},
}
