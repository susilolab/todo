package services

type Response struct {
	Error   int         `json:"error"`
	Message string      `json:"message"`
	Items    interface{} `json:"items"`
}

func NewResponse(err int, msg string, data interface{}) *Response {
	return &Response{
		Error:   err,
		Message: msg,
		Items:    data,
	}
}

func (r *Response) SetErrMsg(msg string) {
	r.Error = 1
	r.Message = msg
	r.Items = nil
}

func (r *Response) SetData(data interface{}) {
	r.Error = 0
	r.Items = data
}
