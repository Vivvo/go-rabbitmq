package rabbitmq

type Config struct {
	Hostname  string `json:"hostname"`
	Port      string `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	QueueName string `json:"queueName"`
}

type RabbitMQ struct {
	Config   Config
	Handlers []MessageHandler
}

type MessageHandler struct {
	Type          string
	ExpectedClass interface{}
	HandlerFunc   func(dto interface{}) error
}

type MessageDto struct {
	PublishChangesMessageType string      `json:"publishChangesMessageType"`
	Payload                   interface{} `json:"payload"`
	Metadata                  interface{} `json:"metadata"`
}
