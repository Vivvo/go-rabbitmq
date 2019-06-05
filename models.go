package rabbitmq

type Config struct {
	Hostname  string
	Port      string
	Username  string
	Password  string
	QueueName string
}

type RabbitMQ struct {
	Config   Config
	Handlers []MessageHandler
}

type MessageHandler struct {
	Type        string
	HandlerFunc func(dto MessageDto) error
}

type MessageDto struct {
	PublishChangesMessageType string      `json:"publishChangesMessageType"`
	Payload                   interface{} `json:"payload"`
	Metadata                  interface{} `json:"metadata"`
}