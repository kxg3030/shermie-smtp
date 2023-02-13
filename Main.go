package main

import (
	"github.com/kxg3030/shermie-smtp/ShermieSmtp"
)

func main() {
	ShermieSmtp.NewServer(9091).Start()
}
