.PHONY: mtop climber mctl

all: mtop climber mctl

mtop:
	CGO_ENABLED=0 go build -o bin/mtop

climber:
	CGO_ENABLED=0 go build -o bin/climber climber/main.go

mctl:
	CGO_ENABLED=0 go build -o bin/mctl mctl/main.go