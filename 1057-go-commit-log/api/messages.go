package api

type Message struct {
	Offset    int64  `json:"offset"`
	Topic     string `json:"topic"`
	Value     []byte `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

type ProduceRequest struct {
	Topic string `json:"topic"`
	Value []byte `json:"value"`
}

type ProduceResponse struct {
	Offset int64 `json:"offset"`
}

type FetchRequest struct {
	Topic  string `json:"topic"`
	Offset int64  `json:"offset"`
	Max    int    `json:"max"`
}

type FetchResponse struct {
	Messages []Message `json:"messages"`
	Next     int64     `json:"next"`
}
