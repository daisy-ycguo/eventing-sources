/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rabbitmq

import (
	"strconv"

	"github.com/Shopify/sarama"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/knative/eventing-sources/pkg/kncloudevents"
	"github.com/knative/pkg/logging"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

const (
	eventType = "dev.knative.rabbitmq.event"
)

type AdapterSASL struct {
	Enable   bool
	User     string
	Password string
}

// TODO we should have a system to use that in the futur
type AdapterTLS struct {
	Enable bool
}

type AdapterNet struct {
	SASL AdapterSASL
	TLS  AdapterTLS
}

type Adapter struct {

	// BootstrapServers is the RabbitMQ Broker URL that we're polling messages from
	BootstrapServers string

	// RabbitMQ Exchange name
	ExchangeName string

	// SinkURI is the URI messages will be forwarded on to.
	SinkURI string

	// Creds to connect to RabbitMQ
	Creds AdapterNet

	// Client sends cloudevents to the target.
	client client.Client
}

// --------------------------------------------------------------------

// Initialize cloudevent client
func (a *Adapter) initClient() error {
	if a.client == nil {
		var err error
		if a.client, err = kncloudevents.NewDefaultClient(a.SinkURI); err != nil {
			return err
		}
	}
	return nil
}

// --------------------------------------------------------------------

func (a *Adapter) Start(ctx context.Context, stopCh <-chan struct{}) error {
	logger := logging.FromContext(ctx)

	logger.Info("Starting with config: ", zap.Any("adapter", a))

	if err := a.initClient(); err != nil {
		logger.Error("Failed to create cloudevent client", zap.Error(err))
		return err
	}

	// Connect to RabbitMQ Brocker
	// TODO we should check if SASL is enable
	amqpConnection := a.BootstrapServers + a.Creds.SASL.User + ":" + a.Creds.SASL.Password + "@rabbitmq/"
	if conn, err := amqp.Dial(amqpConnection); err != nil {
		logger.Error("Failed to connect to RabbitMQ", zap.Error(err))
		return err
	}
	defer conn.Close()

	// Create a channel with the brocker
	if ch, err := conn.Channel(); err != nil {
		logger.Error("Failed to open a channel", zap.Error(err))
		return err
	}
	defer ch.Close()

	if err = ch.ExchangeDeclare(
		a.ExchangeName, // name
		"direct",       // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	); err != nil {
		logger.Error("Failed to declare an exchange", zap.Error(err))
		return err
	}

	if q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	); err != nil {
		logger.Error("Failed to declare an queue", zap.Error(err))
		return err
	}

	return a.pollLoop(ctx, ch, q, stopCh)
}

// pollLoop continuously polls from the given RabbitMQ queue until stopCh
// emits an element.  The
func (a *Adapter) pollLoop(ctx context.Context, ch *amqp.Channel, q *amqp.Queue, stopCh <-chan struct{}) error {

	logger := logging.FromContext(ctx)

	if msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	); err != nil {
		logger.Error("Failed to register a consumer", zap.Error(err))
		return err
	}

	for {
		select {
		case <-stopCh:
			logger.Info("Exiting")
			return nil
		default:
		}
		for m := range msgs {
			/*  Message structure form the queue
			Owner:         m.ReplyTo      		--> Producer URL if needed
			Body:          m.Body				--> Message Body
			CorrelationID: m.CorrelationId		--> Unique ID of the message if needed

			TODO We should also have a critical message system if there is a error
			*/
			a.receiveMessage(ctx, m)
		}
	}
}

func (a *Adapter) postMessage(ctx context.Context, logger *zap.SugaredLogger, msg *amqp.Delivery) error {
	// TODO We have to change that for CloudEvent
	event := cloudevents.Event{
		Context: cloudevents.EventContextV02{
			SpecVersion: cloudevents.CloudEventsVersionV02,
			Type:        eventType,
			ID:          m.CorrelationId,
			Time:        &types.Timestamp{Time: msg.Timestamp},
			Source:      *types.ParseURLRef(msg.Exchange), // TODO not very sure here
			ContentType: cloudevents.StringOfApplicationJSON()
		}.AsV02(),
		Data: a.jsonEncode(ctx, msg.Value),
	}

	_, err := a.client.Send(ctx, event)
	return err
}

// receiveMessage handles an incoming message from the RabbitMQ queue,
// and forwards it to a Sink, calling `Delivery.Ack()` when the forwarding is
// successful.
func (a *Adapter) receiveMessage(ctx context.Context, m *amqp.Delivery) {
	logger := logging.FromContext(ctx).With(zap.Any("eventID", m.MessageId)).With(zap.Any("sink", a.SinkURI))
	logger.Debugw("Received message from RabbitMQ:", zap.Any("message", string(m.Body)))

	err := a.postMessage(ctx, logger, m)
	if err != nil {
		logger.Infof("Event delivery failed: %s", err)
		m.Nack(false, true) // TODO Implement priority system for message that can be drop and system that need to be requeue
	} else {
		logger.Debug("Message successfully posted to Sink")
		m.Ack(false)
	}
}
