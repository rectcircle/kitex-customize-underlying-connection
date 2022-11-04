package main

import (
	"log"

	api "github.com/rectcircle/kitex-customize-underlying-connection/kitex_gen/api/echo"
	"github.com/rectcircle/kitex-customize-underlying-connection/server"
)

func main() {
	svr := api.NewServer(new(server.EchoImpl))

	err := svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}
