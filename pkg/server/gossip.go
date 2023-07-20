package server

import (
	"net/http"

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

type CreateTopicRequest struct {
	Topic string `json:"topic" validate:"required"`
}

type CreateTopicResponse struct {
	Message string `json:"message"`
}

// CreateTopic godoc
//
//	@Summary		Create a topic
//	@Description	Create a topic
//	@Tags			Gossip
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateTopicRequest	true	"Create Topic Request"
//	@Success		201		{object}	CreateTopicResponse
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/v1/gossip [put]
func (r *GossipRouter) CreateTopic(c *gin.Context) {
	var request CreateTopicRequest
	if err := Decode(c.Request, &request); err != nil {
		LoggingRespondErrWithMsg(c, err, "invalid create topic request", http.StatusBadRequest)
		return
	}

	topics := r.service.GetGossipTopics()
	for _, topic := range topics {
		if topic == request.Topic {
			LoggingRespondErrMsg(c, "topic already exists", http.StatusBadRequest)
			return
		}
	}

	if err := r.service.StartGossiper(c, request.Topic); err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to start gossiper", http.StatusInternalServerError)
		return
	}

	Respond(c, CreateTopicResponse{Message: "topic created"}, http.StatusCreated)
}

type ListTopicsResponse struct {
	Topics []string `json:"topics"`
}

// ListTopics godoc
//
//	@Summary		List all topics
//	@Description	List all topics
//	@Tags			Gossip
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	ListTopicsResponse
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/v1/gossip [get]
func (r *GossipRouter) ListTopics(c *gin.Context) {
	topics := r.service.GetGossipTopics()
	Respond(c, ListTopicsResponse{Topics: topics}, http.StatusOK)
}

type PublishMessageRequest struct {
	Message map[string]any `json:"message" validate:"required"`
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
