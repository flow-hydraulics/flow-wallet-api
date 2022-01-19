ifeq (, $(shell which flow))
$(error "No flow in PATH")
endif

ifeq (, $(shell which docker-compose))
$(error "No docker-compose in PATH")
endif

ifeq (, $(shell which go))
$(error "No go in PATH")
endif

.PHONY: dev
dev:
	@docker-compose up -d db pgadmin emulator redis
	@docker-compose logs -f

.PHONY: stop
stop:
	@docker-compose stop

.PHONY: down
down:
	@docker-compose down

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
	@docker-compose -f docker-compose.test-suite.yml build flow test-api
	@docker-compose -f docker-compose.test-suite.yml up --remove-orphans -d db redis flow
	@echo "\nRunning tests, hang on...\n" \
	; docker-compose -f docker-compose.test-suite.yml run --rm test-api go test ./... -p 1 \
	; echo "\nRunning linter, hang on...\n" \
	; docker-compose -f docker-compose.test-suite.yml run --rm lint golangci-lint run \
	; echo "\nStopping services...\n" \
	; docker-compose -f docker-compose.test-suite.yml stop
