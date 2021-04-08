package mqtt

import (
	"fmt"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type Options struct {
	Host     string `json:"host"`
	Port     uint16 `json:"port"`
	Username string `json:"username"`
	Password string `json:"-"`
}

type Client struct {
	client *mqtt.Client
	topics TopicMap
}

func NewClient(o Options) (*Client, error) {
	opts := mqtt.NewClientOptions()

	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", o.Host, o.Port))
	opts.SetClientID(fmt.Sprintf("grafana_%d", rand.Int()))

	if o.Username != "" {
		opts.SetUsername(o.Username)
	}

	if o.Password != "" {
		opts.SetPassword(o.Password)
	}

	log.DefaultLogger.Info("MQTT Connecting")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("error connecting to MQTT broker: %s", token.Error())
	}

	return &Client{
		client: &client,
	}, nil
}

func (c *Client) IsConnected() bool {
	client := *c.client
	return client.IsConnectionOpen()
}

func (c *Client) GetMessages(topic string) ([]Message, bool) {
	return c.topics.Load(topic)
}

func (c *Client) HandleMessage(client mqtt.Client, msg mqtt.Message) {
	topic, ok := c.topics.Load(msg.Topic())

	if ok {
		message := Message{
			timestamp: time.Now(),
			value:     string(msg.Payload()),
		}
		topic = append(topic, message)
		c.topics.Store(msg.Topic(), topic)
	}

}

func (c *Client) Subscribe(topic string) {
	client := *c.client
	_, ok := c.topics.Load(topic)

	if !ok {
		var messages []Message
		c.topics.Store(topic, messages)
		client.Subscribe(topic, 2, c.HandleMessage)
	}
}

func (c *Client) Dispose() {
	client := *c.client
	log.DefaultLogger.Info("MQTT Disconnecting")
	client.Disconnect(250)
}
