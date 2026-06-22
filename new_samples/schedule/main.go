package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var operations = map[string]func(){
	"create":   runCreate,
	"describe": runDescribe,
	"pause":    runPause,
	"unpause":  runUnpause,
	"backfill": runBackfill,
	"update":   runUpdate,
	"list":     runList,
	"delete":   runDelete,
}

func operationNames() string {
	names := make([]string, 0, len(operations))
	for n := range operations {
		names = append(names, n)
	}
	return strings.Join(names, " | ")
}

func main() {
	var mode, op string
	flag.StringVar(&mode, "m", "worker", "Mode: worker | manage")
	flag.StringVar(&op, "op", "", "Operation: "+operationNames())
	flag.Parse()

	switch mode {
	case "worker":
		StartWorker()
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT)
		fmt.Println("Cadence schedule worker started, press ctrl+c to terminate...")
		<-done
	case "manage":
		run, ok := operations[op]
		if !ok {
			logger := BuildLogger()
			logger.Sugar().Fatalf("Unknown operation %q, valid: %s", op, operationNames())
		}
		run()
	default:
		logger := BuildLogger()
		logger.Sugar().Fatalf("Unknown mode %q, valid: worker | manage", mode)
	}
}
