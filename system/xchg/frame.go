package xchg

type Frame struct {
	Src      string `json:"src"`
	Function string `json:"function"`
	Data     []byte `json:"data"`
}
