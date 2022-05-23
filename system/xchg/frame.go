package xchg

type Frame struct {
	Src         string `json:"src"`
	Function    string `json:"function"`
	Transaction string `json:"transaction"`
	Data        []byte `json:"data"`
}
