package dto

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
