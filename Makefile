.PHONY: mtop climber

all: mtop climber

mtop:
	CGO_ENABLED=0 go build -o bin/mtop

climber:
	CGO_ENABLED=0 go build -o bin/climber climber/main.go