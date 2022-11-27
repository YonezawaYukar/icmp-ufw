.PHONY: build clean install up

all: build

build: ## Build the project
	go mod download
	go build -o bin/ cmd/main.go

clean: ## Clean the project
	rm -rf bin/main

install: ## Install the project
	mkdir -p /usr/local/share/icmpfw
	cp ./bin/main /usr/local/share/icmpfw/icmpfw
	cp -n ./bin/config.yaml /usr/local/share/icmpfw/config.yaml
	cp -n ./service/icmpfw.service /etc/systemd/system/icmpfw.service
up:
	systemctl daemon-reload
	systemctl enable icmpfw.service
	systemctl start icmpfw.service