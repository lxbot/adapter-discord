.PHONY: build

build:
	go build -buildmode=plugin -o adapter-discord.so adapter.go
