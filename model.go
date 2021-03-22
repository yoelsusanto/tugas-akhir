package main

type ForwardTraffic struct {
	Service    string           `json:"service"`
	Repetition int              `json:"repetition"`
	Forwards   []ForwardTraffic `json:"forwards"`
}

type ReceivedTraffic struct {
	Repetition int              `json:"repetition"`
	Forwards   []ForwardTraffic `json:"forwards"`
}

type SelfResponse struct {
	Result          string            `json:"result"`
	Repetition      int               `json:"repetition"`
	ForwardResponse []ForwardResponse `json:"forward_response"`
}

type ForwardResponse struct {
	Service         string            `json:"service"`
	Result          string            `json:"result"`
	Repetition      int               `json:"repetition"`
	ForwardResponse []ForwardResponse `json:"forward_response"`
}
