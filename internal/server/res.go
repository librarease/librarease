package server

type Meta struct {
	Total  int  `json:"total"`
	Skip   int  `json:"skip"`
	Limit  int  `json:"limit"`
	Unread *int `json:"unread,omitempty"`
}

type Res struct {
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}
