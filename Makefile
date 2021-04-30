all: build
build:
	go build -mod vendor -o _output/bin/ushift cmd/main.go
clean:
	rm -f _output/bin/ushift