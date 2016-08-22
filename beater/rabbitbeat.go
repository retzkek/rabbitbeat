package beater

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/streadway/amqp"

	"github.com/retzkek/rabbitbeat/config"
)

const (
	RetryFallback = 10 * time.Second
)

type Rabbitbeat struct {
	done         chan struct{}
	config       config.Config
	client       publisher.Client
	connection   *amqp.Connection
	connClosing  chan *amqp.Error
	connBlocking chan amqp.Blocking
	channel      *amqp.Channel
	chanClosing  chan *amqp.Error
	inbox        <-chan amqp.Delivery
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Rabbitbeat{
		done:   make(chan struct{}),
		config: config,
	}

	logp.Info("establishing AMQP connection")
	if err := bt.setupAMQP(); err != nil {
		logp.WTF("%s", err)
	}
	if err := bt.setupConsumer(); err != nil {
		logp.WTF("%s", err)
	}

	return bt, nil
}

func (bt *Rabbitbeat) Run(b *beat.Beat) error {
	logp.Info("rabbitbeat is running! Hit CTRL-C to stop it.")

	bt.client = b.Publisher.Connect()
	for {
		select {
		case <-bt.done:
			return nil
		case c := <-bt.connClosing:
			logp.Warn("AMQP connection closed (%s: %s)", c.Code, c.Reason)
			bt.setupAMQP()
		case c := <-bt.connBlocking:
			if c.Active {
				logp.Warn("AMQP connection blocked (%s)", c.Reason)
			} else {
				logp.Info("AMQP connection unblocked")
			}
		case c := <-bt.chanClosing:
			logp.Warn("AMQP channel closed (%s: %s)", c.Code, c.Reason)
			bt.setupConsumer()
		case r := <-bt.inbox:
			logp.Info("Event received")
			event := common.MapStr{
				"@timestamp": common.Time(time.Now()),
				"type":       b.Name,
			}
			if err := json.Unmarshal(r.Body, &event); err != nil {
				r.Nack(false, false)
				logp.Err("error unmarshalling message (%s)", err)
			} else {
				bt.client.PublishEvent(event)
				r.Ack(false)
				logp.Info("Event sent")
			}
		}
	}
}

func (bt *Rabbitbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}

// amqpURI generates the connection URI from the config parameters.
// If scrubbed is true, the password field is set to "xxx"
func (bt *Rabbitbeat) amqpURI(scrubbed bool) string {
	conf := bt.config
	if scrubbed {
		conf.Password = "xxx"
	}
	tpl := template.Must(template.New("uri").Parse(conf.URITemplate))
	var b bytes.Buffer
	if err := tpl.Execute(&b, conf); err != nil {
		logp.WTF("error generating URI from template \"%s\" (%s)", conf.URITemplate, err)
	}
	return b.String()
}

// retryDuration parses the Retry config setting duration. In case of error the
// RetryFallback duration is returned.
func (bt *Rabbitbeat) retryDuration() time.Duration {
	retry, err := time.ParseDuration(bt.config.Retry)
	if err != nil {
		logp.Warn("error parsing Retry duration, using \"%s\" (%s)", RetryFallback.String(), err)
		return RetryFallback
	}
	return retry
}

// setupAMQP initializes the connection to the AMQP broker
func (bt *Rabbitbeat) setupAMQP() error {
	if bt.connection != nil {
		bt.connection.Close()
	}

	logp.Info("connecting to AMQP broker at %s", bt.amqpURI(true))
	var err error
	connect := func() error {
		bt.connection, err = amqp.Dial(bt.amqpURI(false))
		return err
	}
	for err = connect(); err != nil; err = connect() {
		logp.Err("error connecting to AMQP broker. retry in %s (%s)", bt.config.Retry, err)
		time.Sleep(bt.retryDuration())
	}

	logp.Info("connection established to AMQP broker")

	// listen for close and blocking events
	bt.connClosing = bt.connection.NotifyClose(make(chan *amqp.Error))
	bt.connBlocking = bt.connection.NotifyBlocked(make(chan amqp.Blocking))

	return nil
}

// setupConsumer sets up the AMQP channel and initializes the consumer.
func (bt *Rabbitbeat) setupConsumer() error {
	if bt.connection == nil {
		return fmt.Errorf("AMQP connection not open")
	}

	// setup channel
	var err error
	bt.channel, err = bt.connection.Channel()
	if err != nil {
		return err
	}

	// setup exchange and queue
	if err := bt.channel.ExchangeDeclare(
		bt.config.Exchange,
		bt.config.ExchangeType,
		bt.config.Durable,
		bt.config.AutoDeleteExchange,
		false, //internal
		false, //noWait
		nil,   //args
	); err != nil {
		return fmt.Errorf("error declaring exchange (%s)", err)
	}
	if _, err := bt.channel.QueueDeclare(
		bt.config.Queue,
		bt.config.Durable,
		bt.config.AutoDeleteQueue,
		bt.config.Exclusive,
		false, //noWait
		nil,   //args
	); err != nil {
		return fmt.Errorf("error declaring queue (%s)", err)
	}
	if err := bt.channel.QueueBind(
		bt.config.Queue,
		bt.config.RoutingKey,
		bt.config.Exchange,
		false, //noWait
		nil,   //args
	); err != nil {
		return fmt.Errorf("error binding queue (%s)", err)
	}

	// start consumer
	if bt.inbox, err = bt.channel.Consume(
		bt.config.Queue,
		bt.config.ConsumerTag,
		false, //autoAck
		bt.config.Exclusive,
		true,  //noLocal
		false, //noWait
		nil,   //args
	); err != nil {
		return fmt.Errorf("error starting consumer (%s)", err)
	}

	return nil
}
