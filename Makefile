.PHONY: mtop

all: mtop

mtop:
	CGO_ENABLED=0 go build -o bin/mtop
