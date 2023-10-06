package internal

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	pkarrDir  = "/.pkarr"
	pkarrPath = pkarrDir + "/pkaar.json"
)

// Read reads the pkarr file and returns the identities.
func Read() (Identities, error) {
	homeDir, _ := os.UserHomeDir()
	pkarrFile := homeDir + pkarrPath
	if _, err := os.Stat(pkarrFile); os.IsNotExist(err) {
		logrus.WithError(err).Error("failed to find pkarr file")
		return nil, nil
	}
	f, _ := os.Open(pkarrFile)
	defer f.Close()
	var identities Identities
	if err := json.NewDecoder(f).Decode(&identities); err != nil {
		logrus.WithError(err).Error("failed to decode pkarr file")
		return nil, err
	}
	return identities, nil
}

// Write writes the given identity to the pkarr file.
func Write(id string, identity Identity) error {
	homeDir, _ := os.UserHomeDir()
	pkarrFile := homeDir + pkarrPath
	var identities Identities
	var err error
	if _, err = os.Stat(pkarrFile); os.IsNotExist(err) {
		if err = os.Mkdir(homeDir+pkarrDir, 0700); err != nil {
			logrus.WithError(err).Error("failed to create pkarr directory")
			return err
		}
		if _, err = os.Create(homeDir + pkarrPath); err != nil {
			logrus.WithError(err).Error("failed to create pkarr file")
			return err
		}
		identities = Identities{id: identity}
	} else {
		identities, err = Read()
		if err != nil {
			logrus.WithError(err).Error("failed to read pkarr file")
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

	f, _ := os.OpenFile(pkarrFile, os.O_WRONLY, os.ModeAppend)
	if _, err = f.WriteString(string(identitiesBytes)); err != nil {
		logrus.WithError(err).Error("failed to write identities to pkarr file")
		return err
	}
	return f.Close()
}
