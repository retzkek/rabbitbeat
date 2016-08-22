// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type Config struct {
	User               string `config:"user"`
	Password           string `config:"password"`
	Host               string `config:"host"`
	VHost              string `config:"vhost"`
	Port               string `config:"port"`
	URITemplate        string `config:"uri_template"`
	Retry              string `config:"retry"`
	Exchange           string `config:"exchange"`
	ExchangeType       string `config:"exchange_type"`
	Queue              string `config:"queue"`
	RoutingKey         string `config:"routing_key"`
	Exclusive          bool   `config:"exclusive"`
	Durable            bool   `config:"durable"`
	AutoDeleteExchange bool   `config:"auto_delete_exchange"`
	AutoDeleteQueue    bool   `config:"auto_delete_queue"`
	ConsumerTag        string `config:"consumer_tag"`
}

var DefaultConfig = Config{
	User:               "guest",
	Password:           "guest",
	Host:               "localhost",
	VHost:              "",
	Port:               "5672",
	URITemplate:        "amqp://{{.User}}:{{.Password}}@{{.Host}}:{{.Port}}/{{.VHost}}",
	Retry:              "10s",
	Exchange:           "amqp.fanout",
	ExchangeType:       "fanout",
	Queue:              "rabbitbeat",
	RoutingKey:         "#",
	Exclusive:          false,
	Durable:            true,
	AutoDeleteExchange: false,
	AutoDeleteQueue:    false,
	ConsumerTag:        "",
}
