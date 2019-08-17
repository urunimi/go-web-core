package rest

type Response struct {
	Code    int         `json:"code"`
	Message *string     `json:"message,omitempty"`
	Result  interface{} `json:"result"`
}
