.PHONY: build run clean

build:
	CGO_ENABLED=0 go build -o scriptkiller

run:
	CGO_ENABLED=0 go run .

clean:
	rm -f scriptkiller
