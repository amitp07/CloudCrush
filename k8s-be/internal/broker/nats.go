package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/amitp07/CloudCrush/k8s-be/internal/dto"
	"github.com/nats-io/nats.go"
)

func ConnectNats(url string) (*nats.Conn, error) {
	options := nats.Options{
		Name:          "Image processor API",
		Url:           url,
		Timeout:       10 * time.Second,
		ReconnectWait: 2 * time.Second,
		DisconnectedErrCB: func(c *nats.Conn, err error) {
			log.Printf("Nats disconnected: %v\n", err)
		},
		ReconnectedCB: func(c *nats.Conn) {
			log.Printf("Nats reconnected to: %v\n", c.ConnectedUrl())
		},
	}

	nc, err := options.Connect()

	if err != nil {
		return nil, err
	}

	return nc, nil
}

func CreateStream(js nats.JetStreamContext) error {
	_, err := js.StreamInfo("IMAGES")

	if err != nil {
		fmt.Printf("Stream %q is not found creating it now...\n", "IMAGES")
		_, err := js.AddStream(&nats.StreamConfig{
			Name:     "IMAGES",
			Subjects: []string{"IMAGE.*"},
			Storage:  nats.FileStorage,
		})

		if err != nil {
			return err
		}
	}
	return nil
}

type NatsBroker struct {
	nc *nats.Conn
	Js nats.JetStreamContext
}

func NewNatsBroker(nc *nats.Conn) (*NatsBroker, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	return &NatsBroker{
		nc: nc,
		Js: js,
	}, nil
}

func (nb *NatsBroker) PublishImageJob(ctx context.Context, imageJob dto.ImageJob) error {
	imageBytes, _ := json.Marshal(imageJob)
	_, err := nb.Js.Publish("IMAGE.created", []byte(imageBytes))

	return err
}
