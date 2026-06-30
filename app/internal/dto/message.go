package dto

import (
	"time"

	"github.com/google/uuid"
)

// Message base message entity
type Message struct {
	ID        uuid.UUID                `json:"id"`                 // ID message unique id
	SenderID  string                   `json:"sender_id"`          // SenderID sender id
	To        []string                 `json:"to"`                 // To array of recipients
	Meta      map[string]interface{}   `json:"metadata"`           // Meta some metadata for use specific options
	Status    string                   `json:"status"`             // Status message status
	Code      string                   `json:"code,omitempty"`     // Code template code
	Params    map[string]*MessageParam `json:"params,omitempty"`   // Params k->v data for set into template
	Transport string                   `json:"transport"`          // Transport transport code
	Subject   string                   `json:"subject"`            // Subject message subject
	Body      string                   `json:"body"`               // Body message body
	Retry     *Retry                   `json:"retry,omitempty"`    // Retry retry policy
	Schedule  *Schedule                `json:"schedule,omitempty"` // Schedule send message on planned time
}

type MessageParam struct {
	Value any `json:"value"` // Value param value
}

// Template message template with default config
type Template struct {
	Code        string                    `json:"code"`        // Code template code
	Name        string                    `json:"name"`        // Name template name
	Description string                    `json:"description"` // Description template description
	Params      map[string]*TemplateParam `json:"params"`      // Params Parameters in template
	Subject     string                    `json:"subject"`     // Subject Message subject
	Body        string                    `json:"body"`        // Body Message body
	Retry       *Retry                    `json:"retry"`       // Retry Default retry policy for template
}

type TemplateParam struct {
	Default string `json:"default"` // Default value
}

// Retry retry policy config
type Retry struct {
	Times    int           `json:"times,omitempty"`    // Times retry count
	Interval time.Duration `json:"interval,omitempty"` // Interval first interval
	Strategy string        `json:"strategy,omitempty"` // Strategy send interval strategy
}

// Schedule schedule policy config
type Schedule struct {
	When time.Time `json:"when,omitempty"` // When send message on specific type
}
