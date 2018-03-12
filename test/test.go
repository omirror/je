package main

import (
	"fmt"
	"log"

	"github.com/prologic/je/worker"
)

func main() {
	res, err := worker.Run("./hello.sh")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(
		"Logs: %v\nResponse: %v\nExit Code: %d\n",
		res.Logs(),
		res.Response(),
		res.Status(),
	)
}
