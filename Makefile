ifeq (, $(shell which flow))
$(error "No flow in PATH")
endif

ifeq (, $(shell which docker-compose))
$(error "No docker-compose in PATH")
endif

ifeq (, $(shell which go))
$(error "No go in PATH")
endif


dev = docker-compose -f docker-compose.dev.yml -p flow-wallet-api-dev
test-suite = docker-compose -f docker-compose.test-suite.yml -p flow-wallet-api-test

.PHONY: dev
dev:
	@$(dev) up --remove-orphans -d db pgadmin emulator redis

.PHONY: stop
stop:
	@$(dev) stop

.PHONY: down
down:
	@$(dev) down --remove-orphans

.PHONY: reset
reset: down dev

.PHONY: run-tests
run-tests:
	@go test ./... -p 1

.PHONY: test
test: start-emulator deploy run-tests

.PHONY: test-clean
test-clean: clean-test-cache test

.PHONY: clean-test-cache
clean-test-cache:
	@go clean -testcache

.PHONY: deploy
deploy:
	@cd flow && flow project deploy --update

.PHONY: start-emulator
start-emulator: emulator.pid
	@sleep 1

.PHONY: stop-emulator
stop-emulator: emulator.pid
	@kill `cat $<` && rm $<

emulator.pid:
	@cd flow && { flow emulator -b 100ms & echo $$! > ../$@; }

.PHONY: lint
lint:
	@golangci-lint run

.PHONY: run-test-suite
run-test-suite:
	@$(test-suite) build flow api
	@$(test-suite) up --remove-orphans -d db redis flow
	@$(test-suite) unpause \
	; echo "\nRunning tests, hang on...\n" \
	; $(test-suite) run --rm api go test ./... -p 1 \
	; echo "\nRunning linter, hang on...\n" \
	; $(test-suite) run --rm lint golangci-lint run \
	; $(test-suite) pause

.PHONY: stop-test-suite
stop-test-suite:
	@$(test-suite) down --remove-orphans

.PHONY: clean-test-suite
clean-test-suite:
	@$(test-suite) run --rm api go clean -testcache
