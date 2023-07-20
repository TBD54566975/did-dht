package server

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"did-dht/pkg/service/gossip"
)

type GossipRouter struct {
	service *gossip.Service
}

func NewGossipRouter(service *gossip.Service) (*GossipRouter, error) {
	if service == nil {
		return nil, errors.New("service cannot be nil")
	}
	return &GossipRouter{service: service}, nil
}

// CreateTopic creates a new topic
func (r *GossipRouter) CreateTopic(c *gin.Context) {

}

// ListTopics lists all topics
func (r *GossipRouter) ListTopics(c *gin.Context) {

}

// PublishMessageToTopic publishes a message to a given topic
func (r *GossipRouter) PublishMessageToTopic(c *gin.Context) {

}

// ListMessagesForTopic lists all messages for a given topic
func (r *GossipRouter) ListMessagesForTopic(c *gin.Context) {

}

// SearchTopic searches a topic for a given message
func (r *GossipRouter) SearchTopic(c *gin.Context) {

}
