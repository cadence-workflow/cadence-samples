package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// scenarios maps the -scenario flag to its driver.
var scenarios = map[string]func(){
	"lifecycle":     runLifecycle,
	"overlap":       runOverlap,
	"catchup":       runCatchUp,
	"pagination":    runPagination,
	"dataconverter": runDataConverter,
}

func scenarioNames() string {
	names := make([]string, 0, len(scenarios))
	for n := range scenarios {
		names = append(names, n)
	}
	return strings.Join(names, " | ")
}

func main() {
	var mode, scenario string
	flag.StringVar(&mode, "m", "worker", "Mode: worker | manage")
	flag.StringVar(&scenario, "scenario", "lifecycle", "manage scenario: "+scenarioNames())
	flag.Parse()

	switch mode {
	case "worker":
		StartWorker()
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT)
		fmt.Println("Cadence schedule worker started, press ctrl+c to terminate...")
		<-done
	case "manage":
		run, ok := scenarios[scenario]
		if !ok {
			logger := BuildLogger()
			logger.Sugar().Fatalf("Unknown scenario %q, valid: %s", scenario, scenarioNames())
		}
		run()
	default:
		logger := BuildLogger()
		logger.Sugar().Fatalf("Unknown mode %q, valid: worker | manage", mode)
	}
}
