// THIS IS A GENERATED FILE
// PLEASE DO NOT EDIT

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	StartWorker()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT)
	fmt.Println("Cadence worker started, press ctrl+c to terminate...")
	<-done
}
