.PHONY: test bins clean
PROJECT_ROOT = github.com/uber-common/cadence-samples

export PATH := $(GOPATH)/bin:$(PATH)

# default target
default: test

PROGS = helloworld \
	branch \
	childworkflow \
	crossdomain \
	choice \
	dynamic \
	greetings \
	pickfirst \
	retryactivity \
	splitmerge \
	timer \
	localactivity \
	query \
	consistentquery \
	cron \
	tracing \
	dsl \
	fileprocessing \
	expense_dummy \
	expense \
	recovery \
	cancelactivity \
	ctxpropagation \
	pso \
	pageflow \
	signalcounter \
	sideeffect \

TEST_ARG ?= -race -v -timeout 5m
BUILD := ./build
SAMPLES_DIR=./cmd/samples

export PATH := $(GOPATH)/bin:$(PATH)

# Automatically gather all srcs
ALL_SRC := $(shell find ./cmd/samples/common -name "*.go")

# all directories with *_test.go files in them
TEST_DIRS=./cmd/samples/cron \
	./cmd/samples/dsl \
	./cmd/samples/expense \
	./cmd/samples/fileprocessing \
	./cmd/samples/recipes/branch \
	./cmd/samples/recipes/choice \
	./cmd/samples/recipes/greetings \
	./cmd/samples/recipes/helloworld \
	./cmd/samples/recipes/cancelactivity \
	./cmd/samples/recipes/pickfirst \
	./cmd/samples/recipes/mutex \
	./cmd/samples/recipes/retryactivity \
	./cmd/samples/recipes/splitmerge \
	./cmd/samples/recipes/timer \
	./cmd/samples/recipes/localactivity \
	./cmd/samples/recipes/query \
	./cmd/samples/recipes/consistentquery \
	./cmd/samples/recipes/ctxpropagation \
	./cmd/samples/recipes/searchattributes \
	./cmd/samples/recipes/sideeffect \
	./cmd/samples/recipes/signalcounter \
	./cmd/samples/recovery \
	./cmd/samples/pso \


cancelactivity:
	go build -i -o bin/cancelactivity cmd/samples/recipes/cancelactivity/*.go

helloworld:
	go build -i -o bin/helloworld cmd/samples/recipes/helloworld/*.go

branch:
	go build -i -o bin/branch cmd/samples/recipes/branch/*.go

childworkflow:
	go build -i -o bin/childworkflow cmd/samples/recipes/childworkflow/*.go

choice:
	go build -i -o bin/choice cmd/samples/recipes/choice/*.go

dynamic:
	go build -i -o bin/dynamic cmd/samples/recipes/dynamic/*.go

greetings:
	go build -i -o bin/greetings cmd/samples/recipes/greetings/*.go

pickfirst:
	go build -i -o bin/pickfirst cmd/samples/recipes/pickfirst/*.go

mutex:
	go build -i -o bin/mutex cmd/samples/recipes/mutex/*.go

retryactivity:
	go build -i -o bin/retryactivity cmd/samples/recipes/retryactivity/*.go

splitmerge:
	go build -i -o bin/splitmerge cmd/samples/recipes/splitmerge/*.go

searchattributes:
	go build -i -o bin/searchattributes cmd/samples/recipes/searchattributes/*.go

timer:
	go build -i -o bin/timer cmd/samples/recipes/timer/*.go

localactivity:
	go build -i -o bin/localactivity cmd/samples/recipes/localactivity/*.go

query:
	go build -i -o bin/query cmd/samples/recipes/query/*.go

consistentquery:
	go build -i -o bin/consistentquery cmd/samples/recipes/consistentquery/*.go

ctxpropagation:
	go build -i -o bin/ctxpropagation cmd/samples/recipes/ctxpropagation/*.go

tracing:
	go build -i -o bin/tracing cmd/samples/recipes/tracing/*.go

cron:
	go build -i -o bin/cron cmd/samples/cron/*.go

dsl:
	go build -i -o bin/dsl cmd/samples/dsl/*.go

fileprocessing:
	go build -i -o bin/fileprocessing cmd/samples/fileprocessing/*.go

expense_dummy:
	go build -i -o bin/expense_dummy cmd/samples/expense/server/*.go

expense:
	go build -i -o bin/expense cmd/samples/expense/*.go

recovery:
	go build -i -o bin/recovery cmd/samples/recovery/*.go

pso:
	go build -i -o bin/pso cmd/samples/pso/*.go

pageflow:
	go build -i -o bin/pageflow cmd/samples/pageflow/*.go

signalcounter:
	go build -i -o bin/signalcounter cmd/samples/recipes/signalcounter/*.go

crossdomain:
	go build -o bin/crossdomain cmd/samples/recipes/crossdomain/*.go

crossdomain-setup:
	# use the ..cadence-server --env development_xdc_cluster0 ... to set up three
	cadence --ad 127.0.0.1:7933 --env development --do domain0 domain register --ac cluster0 --gd true --clusters cluster0 cluster1 # global domain required
	cadence --ad 127.0.0.1:7933 --env development --do domain1 domain register --ac cluster1 --gd true --clusters cluster0 cluster1
	cadence --ad 127.0.0.1:7933 --env development --do domain2 domain register --ac cluster0 --gd true --clusters cluster0 cluster1

crossdomain-run: crossdomain
	tmux split-window -h './bin/crossdomain -m "worker0"' \; \
		split-window -v './bin/crossdomain -m "worker1"' \; \
		split-window -v './bin/crossdomain -m "worker2"'

sideeffect:
	go build -i -o bin/sideeffect cmd/samples/recipes/sideeffect/*.go

bins: helloworld \
	branch \
	crossdomain \
	childworkflow \
	choice \
	dynamic \
	greetings \
	pickfirst \
	mutex \
	cancelactivity \
	retryactivity \
	splitmerge \
	searchattributes \
	timer \
	cron \
	tracing \
	dsl \
	fileprocessing \
	expense_dummy \
	expense \
	localactivity \
	query \
	consistentquery \
	recovery \
	ctxpropagation \
	pso \
	pageflow \
	signalcounter \
	sideeffect \

test: bins
	@rm -f test
	@rm -f test.log
	@echo $(TEST_DIRS)
	@for dir in $(TEST_DIRS); do \
		go test -coverprofile=$@ "$$dir" | tee -a test.log; \
	done;

clean:
	rm -rf bin
	rm -Rf $(BUILD)
