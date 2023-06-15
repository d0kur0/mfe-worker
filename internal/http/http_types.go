package http

type ResponseMeta struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total"`
}

type Response struct {
	Meta    ResponseMeta `json:"_meta"`
	Payload interface{}  `json:"payload"`
}
