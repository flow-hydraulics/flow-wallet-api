
current_dir = $(shell pwd)

cli_version = 0.24.0
cli_image=flow-cli

.PHONY: run
up:
	FLOW_CLI_VERSION=${cli_version} docker-compose up -d

.PHONY: stop
stop:
	FLOW_CLI_VERSION=${cli_version} docker-compose stop

.PHONY: down
down:
	FLOW_CLI_VERSION=${cli_version} docker-compose down

.PHONY: test
test:
	go test ./...

.PHONY: build-cli
build-cli:
	docker build \
		--network host \
		--build-arg VERSION=${cli_version} \
		-t ${cli_image}:${cli_version} \
		./docker/flow-cli

.PHONY: deploy
deploy:
	docker run --rm -it -v "${current_dir}:/app" --net host ${cli_image}:${cli_version} project deploy --update

.PHONY: shell
shell:
	docker exec -it flow-emulator ash

.PHONY: reset
reset:
	FLOW_CLI_VERSION=${cli_version} docker-compose down
	FLOW_CLI_VERSION=${cli_version} docker-compose up -d
