// pkg/notifier/client.go
package notifier

type Client interface {
	SendNotification(to, subject, message string) error
}

type Config struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
}

type emailClient struct {
	cfg Config
}

func NewClient(cfg Config) Client {
	return &emailClient{cfg: cfg}
}

func (c *emailClient) SendNotification(to, subject, message string) error {
	// 实现邮件发送逻辑
	return nil
}
