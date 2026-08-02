package response

const (
	statusSuccess = "success"
	statusError   = "error"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Data   any    `json:"data,omitempty"`
}

func NewSuccess(data any) Response {
	return Response{
		Status: statusSuccess,
		Error:  "",
		Data:   data,
	}
}

func NewError(err string) Response {
	return Response{
		Status: statusError,
		Error:  err,
		Data:   nil,
	}
}
