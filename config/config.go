// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"time"
)

type Config struct {
	Host        string         `config:"host"`
	Port        string         `config:"port"`
	VHost       string         `config:"vhost"`
	User        string         `config:"user"`
	Password    string         `config:"password"`
	URITemplate string         `config:"uri"`
	Retry       time.Duration  `config:"retry" validate:"nonzero,min=0s"`
	Exchange    ExchangeConfig `config:"exchange"`
	Queue       QueueConfig    `config:"queue"`
	ConsumerTag string         `config:"consumer_tag"`
	Exclusive   bool           `config:"exclusive"`
}

type ExchangeConfig struct {
	Name       string `config:"name"`
	Type       string `config:"type"`
	Durable    bool   `config:"durable"`
	AutoDelete bool   `config:"auto_delete"`
}

type QueueConfig struct {
	Name       string `config:"name"`
	RoutingKey string `config:"routing_key"`
	Exclusive  bool   `config:"exclusive"`
	Durable    bool   `config:"durable"`
	AutoDelete bool   `config:"auto_delete"`
}

var DefaultConfig = Config{
	Host:        "localhost",
	Port:        "5672",
	VHost:       "",
	User:        "guest",
	Password:    "guest",
	URITemplate: "amqp://{{.User}}:{{.Password}}@{{.Host}}:{{.Port}}/{{.VHost}}",
	Retry:       10 * time.Second,
	Exchange: ExchangeConfig{
		Name:       "amqp.fanout",
		Type:       "fanout",
		Durable:    true,
		AutoDelete: false,
	},
	Queue: QueueConfig{
		Name:       "rabbitbeat",
		RoutingKey: "#",
		Exclusive:  false,
		Durable:    true,
		AutoDelete: false,
	},
	ConsumerTag: "",
	Exclusive:   false,
}
