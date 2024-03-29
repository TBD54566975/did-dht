package cli

import (
	"os"

	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/goccy/go-json"

	"github.com/TBD54566975/did-dht-method/internal"
)

const (
	didDHTDir  = "/.diddht"
	didDHTPath = didDHTDir + "/diddht.json"
)

// Read reads the diddht file and returns the identities.
func Read() (internal.Identities, error) {
	homeDir, _ := os.UserHomeDir()
	diddhtFile := homeDir + didDHTPath
	if _, err := os.Stat(diddhtFile); os.IsNotExist(err) {
		return nil, util.LoggingErrorMsg(err, "failed to find diddht file")
	}
	f, _ := os.Open(diddhtFile)
	defer f.Close()
	var identities internal.Identities
	if err := json.NewDecoder(f).Decode(&identities); err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to decode diddht file")
	}
	return identities, nil
}

// Write writes the given identity to the diddht file.
func Write(id string, identity internal.Identity) error {
	homeDir, _ := os.UserHomeDir()
	didDHTFile := homeDir + didDHTPath
	var identities internal.Identities
	var err error
	if _, err = os.Stat(didDHTFile); os.IsNotExist(err) {
		if err = os.Mkdir(homeDir+didDHTDir, 0700); err != nil {
			return util.LoggingErrorMsg(err, "failed to create diddht directory")
		}
		if _, err = os.Create(homeDir + didDHTPath); err != nil {
			return util.LoggingErrorMsg(err, "failed to create diddht file")
		}
		identities = internal.Identities{id: identity}
	} else {
		identities, err = Read()
		if err != nil {
			return util.LoggingErrorMsg(err, "failed to read diddht file")
		}
		if _, ok := identities[identity.Base58PublicKey]; ok {
			return util.LoggingErrorMsg(err, "identity already exists")
		}
		identities[id] = identity
	}

	identitiesBytes, err := json.Marshal(identities)
	if err != nil {
		return util.LoggingErrorMsg(err, "failed to marshal identities")
	}

	f, _ := os.OpenFile(didDHTFile, os.O_WRONLY, os.ModeAppend)
	if _, err = f.WriteString(string(identitiesBytes)); err != nil {
		return util.LoggingErrorMsg(err, "failed to write identities to diddht file")
	}
	return f.Close()
}
