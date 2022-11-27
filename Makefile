.PHONY: build clean install up

all: build

build: ## Build the project
	go mod download
	go build -o bin/ cmd/main.go

clean: ## Clean the project
	rm -rf bin/main

install: ## Install the project
	mkdir /usr/local/share/icmpfw
	cp -r ./bin/main /usr/local/share/icmpfw/icmpfw
	cp ./bin/config.yaml /usr/local/share/icmpfw/config.yaml
	cp ./icmpfw.service /etc/systemd/system/icmpfw.service
up:
	systemctl daemon-reload
	systemctl enable icmpfw.service
	systemctl start icmpfw.service