package main

import "github.com/An-Owlbear/homecloud/backend/internal/server"

func main() {
	e := server.CreateServer()
	e.Logger.Fatal(e.Start(":1323"))
}
