
current_dir = $(shell pwd)

cli_version = 0.21.0
cli_image=flow-cli:${cli_version}

.PHONY: run
up:
	docker-compose up -d

.PHONY: stop
stop:
	docker-compose stop

.PHONY: down
down:
	docker-compose down

.PHONY: test
test:
	go test ./...

.PHONY: build-cli
build-cli:
	docker build --build-arg VERSION=${cli_version} -t ${cli_image} ./docker/flow-cli

.PHONY: deploy
deploy:
	docker run --rm -it -v "${current_dir}:/app" --net host ${cli_image} project deploy --update

.PHONY: shell
shell:
	docker exec -it flow-emulator ash
