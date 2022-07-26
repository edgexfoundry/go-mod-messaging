package pkg

import (
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"
)

type NoopClient struct{}

func (n NoopClient) Connect() error {
	panic("implement me")
}

func (n NoopClient) Publish(message types.MessageEnvelope, topic string) error {
	panic("implement me")
}

func (n NoopClient) Subscribe(topics []types.TopicChannel, messageErrors chan error) error {
	panic("implement me")
}

func (n NoopClient) Disconnect() error {
	panic("implement me")
}
