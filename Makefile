APP_NAME := google-calendar-app

.PHONY: build clean fmt vet

build:
	# Builds the application binary with size optimization flags
	# -ldflags="-w -s" removes debug information to reduce binary size
	# -o specifies the output filename
	go build -ldflags="-w -s" -o $(APP_NAME) main.go

clean:
	# Removes the compiled binary and cleans Go build cache
	rm -rf $(APP_NAME)
	go clean

fmt:
	# Formats all Go source files in the project according to Go standards
	go fmt ./...

vet:
	# Performs static analysis to find potential errors in the code
	go vet ./...
