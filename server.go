package main

import "github.com/kxg3030/shermie-smtp/service"

func main() {
	service.NewServer(9091).StartTLS()
}
