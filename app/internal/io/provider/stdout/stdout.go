package stdout

import (
	"fmt"
	"os"
	"strings"

	"github.com/devian2011/msgchute/pkg/shared/provider"
)

type Stdout struct{}

func NewStdout() *Stdout {
	return &Stdout{}
}

func (c *Stdout) SetParams(params []byte) error {
	return nil
}

func (c *Stdout) GetCode() string {
	return "stdout"
}

func (c *Stdout) Send(msg *provider.Message) error {
	tmpl := `
------------------------------------------------------
	To: %s
    Meta: %s

	Subject: %s

	Body: %s
------------------------------------------------------
`
	_, wErr := os.Stdout.WriteString(
		fmt.Sprintf(
			tmpl,
			strings.Join(msg.To, ","),
			msg.Meta,
			msg.Subject,
			msg.Body))

	return wErr
}
