package cli

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/TBD54566975/did-dht-method/internal"
)

const (
	diddhtDir  = "/.diddht"
	diddhtPath = diddhtDir + "/diddht.json"
)

// Read reads the diddht file and returns the identities.
func Read() (internal.Identities, error) {
	homeDir, _ := os.UserHomeDir()
	diddhtFile := homeDir + diddhtPath
	if _, err := os.Stat(diddhtFile); os.IsNotExist(err) {
		logrus.WithError(err).Error("failed to find diddht file")
		return nil, nil
	}
	f, _ := os.Open(diddhtFile)
	defer f.Close()
	var identities internal.Identities
	if err := json.NewDecoder(f).Decode(&identities); err != nil {
		logrus.WithError(err).Error("failed to decode diddht file")
		return nil, err
	}
	return identities, nil
}

// Write writes the given identity to the diddht file.
func Write(id string, identity internal.Identity) error {
	homeDir, _ := os.UserHomeDir()
	diddhtFile := homeDir + diddhtPath
	var identities internal.Identities
	var err error
	if _, err = os.Stat(diddhtFile); os.IsNotExist(err) {
		if err = os.Mkdir(homeDir+diddhtDir, 0700); err != nil {
			logrus.WithError(err).Error("failed to create diddht directory")
			return err
		}
		if _, err = os.Create(homeDir + diddhtPath); err != nil {
			logrus.WithError(err).Error("failed to create diddht file")
			return err
		}
		identities = internal.Identities{id: identity}
	} else {
		identities, err = Read()
		if err != nil {
			logrus.WithError(err).Error("failed to read diddht file")
			return err
		}
		if _, ok := identities[identity.Base58PublicKey]; ok {
			logrus.WithError(err).Error("identity already exists")
			return err
		}
		identities[id] = identity
	}

	identitiesBytes, err := json.Marshal(identities)
	if err != nil {
		logrus.WithError(err).Error("failed to marshal identities")
		return err
	}

	f, _ := os.OpenFile(diddhtFile, os.O_WRONLY, os.ModeAppend)
	if _, err = f.WriteString(string(identitiesBytes)); err != nil {
		logrus.WithError(err).Error("failed to write identities to diddht file")
		return err
	}
	return f.Close()
}
