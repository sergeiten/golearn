GO_PACKAGES=$(shell ls -d */ | grep -v vendor)

default: build

.PHONY: build
build:
	docker-compose --project-name golearn up --force-recreate --build --remove-orphans -d

.PHONY: quality
quality:
	go test -v -race ./...
	go vet ./...
	golint -set_exit_status $(go list ./...)
	megacheck ./...
	gocyclo -over 12 $(GO_PACKAGES)
