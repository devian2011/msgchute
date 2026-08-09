package dto

type MessageDictionaries struct {
	SenderIDs  []string `json:"sender_ids"`
	Transports []string `json:"transports"`
	Templates  []string `json:"templates"`
}

type FullMessageInfo struct {
	Message Message    `json:"message"`
	Tasks   []FullTask `json:"tasks"`
}

type FullTask struct {
	Task    Task                  `json:"task"`
	Results []TaskExecutionResult `json:"results"`
}

type MessagePreview struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type AddMessageResponse struct {
	Message *Message `json:"message"`
	Task    *Task    `json:"task"`
}

type AddBatchMessageResponse struct {
	Message *Message `json:"message"`
	Task    *Task    `json:"task"`
	Err     error    `json:"error"`
}
