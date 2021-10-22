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
	@docker-compose up -d db pgadmin emulator
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
	@go test
	@go test ./...

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
	@cd flow && { flow emulator & echo $$! > ../$@; }

.PHONY: lint
lint:
	@golangci-lint run
