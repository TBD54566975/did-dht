package storage

import (
	"strconv"

	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/goccy/go-json"

	didint "github.com/TBD54566975/did-dht-method/internal/did"
)

const (
	gatewayNamespace = "dids"
	typesNamespace   = "types"
)

type GatewayRecord struct {
	// TODO(gabe) when historical document storage is supported, this should be a list of documents
	Document       did.Document       `json:"document" validate:"required"`
	Types          []didint.TypeIndex `json:"types,omitempty"`
	SequenceNumber int64              `json:"sequence_number" validate:"required"`
	RetentionProof string             `json:"retention_proof,omitempty"`
}

type TypeRecord struct {
	Types []string `json:"dids,omitempty"`
}

// WriteDID writes a DID to the storage and adds it to the type index(es) it is associated with
func (s *Storage) WriteDID(record GatewayRecord) error {
	// note current types for the DID to make sure we update the appropriate indexes
	gotDID, err := s.ReadDID(record.Document.ID)
	var currTypes []didint.TypeIndex
	if err == nil && gotDID != nil {
		currTypes = gotDID.Types
	}
	recordBytes, err := json.Marshal(record)
	if err != nil {
		return err
	}
	if err = s.Write(gatewayNamespace, record.Document.ID, recordBytes); err != nil {
		return err
	}
	return s.UpdateTypeIndexesForDID(record.Document.ID, currTypes, record.Types)
}

// ReadDID reads a DID from the storage by ID
func (s *Storage) ReadDID(id string) (*GatewayRecord, error) {
	recordBytes, err := s.Read(gatewayNamespace, id)
	if err != nil {
		return nil, err
	}
	if len(recordBytes) == 0 {
		return nil, nil
	}
	var record GatewayRecord
	if err = json.Unmarshal(recordBytes, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

// UpdateTypeIndexesForDID is an orchestration method that updates the type indexes for a DID
// It checks the existing type indexes for the DID and adds/removes the DID from the appropriate type indexes
func (s *Storage) UpdateTypeIndexesForDID(id string, currTypes, newTypes []didint.TypeIndex) error {
	currTypeMap := make(map[didint.TypeIndex]bool)
	for _, currType := range currTypes {
		currTypeMap[currType] = true
	}
	newTypeMap := make(map[didint.TypeIndex]bool)
	for _, newType := range newTypes {
		newTypeMap[newType] = true
	}

	// remove the DID from any type indexes it is no longer associated with
	for _, currType := range currTypes {
		if _, ok := newTypeMap[currType]; !ok {
			if err := s.RemoveDIDFromTypeIndex(id, currType); err != nil {
				return err
			}
		}
	}

	// add the DID to any type indexes it is now associated with
	for _, newType := range newTypes {
		if _, ok := currTypeMap[newType]; !ok {
			if err := s.AddDIDToTypeIndex(id, newType); err != nil {
				return err
			}
		}
	}

	return nil
}

// AddDIDToTypeIndex adds a DID to a type index by appending it to the list of DIDs for that type index
// If the type index does not exist, it is created and the DID is added to it
func (s *Storage) AddDIDToTypeIndex(id string, typeIndex didint.TypeIndex) error {
	t := strconv.Itoa(int(typeIndex))
	recordBytes, err := s.Read(typesNamespace, t)
	if err != nil {
		return err
	}
	if len(recordBytes) == 0 {
		record := TypeRecord{Types: []string{id}}
		recordBytes, err = json.Marshal(record)
		if err != nil {
			return err
		}
		return s.Write(typesNamespace, t, recordBytes)
	}
	var record TypeRecord
	if err = json.Unmarshal(recordBytes, &record); err != nil {
		return err
	}
	record.Types = append(record.Types, id)
	recordBytes, err = json.Marshal(record)
	if err != nil {
		return err
	}
	return s.Write(typesNamespace, t, recordBytes)
}

// RemoveDIDFromTypeIndex removes a DID from a type index by removing it from the list of DIDs for that type index
func (s *Storage) RemoveDIDFromTypeIndex(id string, typeIndex didint.TypeIndex) error {
	t := strconv.Itoa(int(typeIndex))
	recordBytes, err := s.Read(typesNamespace, t)
	if err != nil {
		return err
	}
	if len(recordBytes) == 0 {
		return nil
	}
	var record TypeRecord
	if err = json.Unmarshal(recordBytes, &record); err != nil {
		return err
	}
	for i, didID := range record.Types {
		if didID == id {
			record.Types = append(record.Types[:i], record.Types[i+1:]...)
			break
		}
	}
	recordBytes, err = json.Marshal(record)
	if err != nil {
		return err
	}
	return s.Write(typesNamespace, t, recordBytes)
}

// ListDIDsForType returns a list of DIDs for a given type index
func (s *Storage) ListDIDsForType(typeIndex didint.TypeIndex) ([]string, error) {
	t := strconv.Itoa(int(typeIndex))
	recordBytes, err := s.Read(typesNamespace, t)
	if err != nil {
		return nil, err
	}
	if len(recordBytes) == 0 {
		return nil, nil
	}
	var record TypeRecord
	if err = json.Unmarshal(recordBytes, &record); err != nil {
		return nil, err
	}
	return record.Types, nil
}
