.PHONY: run build

build:
	go build -o out

run: build
	./out