package server

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"did-dht/pkg/service"
)

type DHTRouter struct {
	service *service.DHTService
}

func NewDHTRouter(service *service.DHTService) (*DHTRouter, error) {
	if service == nil {
		return nil, errors.New("service cannot be nil")
	}
	return &DHTRouter{
		service: service,
	}, nil
}

type AddDHTRecord struct {
	RequesterID   string `json:"requesterId" validate:"required"`
	RequesterName string `json:"requesterName" validate:"required"`
	DID           string `json:"did" validate:"required"`
	Endpoint      string `json:"endpoint" validate:"required"`
}

func (r AddDHTRecord) toServiceRequest() service.DHTMessage {
	return service.DHTMessage{
		Requester: service.Requester{
			ID:   r.RequesterID,
			Name: r.RequesterName,
		},
		Record: service.Record{
			DID:      r.DID,
			Endpoint: r.Endpoint,
		},
	}
}

func (r *DHTRouter) AddDHTRecord(ctx context.Context, req AddDHTRecord) error {
	if err := validator.New().Struct(req); err != nil {
		return errors.WithMessage(err, "failed to validate request")
	}

	return r.service.Gossip(ctx, req.toServiceRequest())
}

type ReadDHTRecord struct {
	DID string `json:"did" validate:"required"`
}

type ReadDHTRecords struct{}

type RemoveDHTRecord struct {
	Requester string `json:"requester" validate:"required"`
	DID       string `json:"did" validate:"required"`
}
