APP_NAME=monkey
DIR=main
BIN_PATH=/usr/local/bin/$(APP_NAME)

.PHONY: all build run clean tidy fmt vet install uninstall cross-build

all: fmt vet build

build: tidy
	cd $(DIR) && go build -ldflags "-s -w" -o $(APP_NAME) monkey.go

run: build
	./$(DIR)/$(APP_NAME)

fmt:
	@cd $(DIR) && go fmt ./...

vet:
	@cd $(DIR) && go vet ./...

tidy:
	@cd $(DIR) && go mod tidy

clean: 
	@rm -f $(DIR)/$(APP_NAME)
	@echo "cleaned."

install: build
	@sudo mv $(DIR)/$(APP_NAME) $(BIN_PATH)
	@echo "installed to $(BIN_PATH). run 'monkey' anywhere."

uninstall:
	@sudo rm -f $(BIN_PATH)
	@echo "uninstalled."

cross-build: tidy
	@cd $(DIR) && GOOS=windows GOARCH=amd64 go build -o $(APP_NAME).exe monkey.go
	@cd $(DIR) && GOOS=darwin GOARCH=arm64 go build -o $(APP_NAME)-mac monkey.go
	@cd $(DIR) && GOOS=linux GOARCH=amd64 go build -o $(APP_NAME)-linux monkey.go
	@echo "cross-compilation complete."
