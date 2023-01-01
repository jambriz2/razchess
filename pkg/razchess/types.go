package razchess

type Update struct {
	FEN       string    `json:"fen"`
	WhiteMove [2]string `json:"wm"`
	BlackMove [2]string `json:"bm"`
}
