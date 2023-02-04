ifeq ($(PREFIX),)
    PREFIX := /usr/local
endif

.PHONY: mtop climber mctl

all: mtop climber mctl

GOFLAGS=-ldflags "-X github.com/wikylyu/mtop/config.PREFIX=$(PREFIX)"

mtop:
	CGO_ENABLED=0 go build $(GOFLAGS) -o bin/mtop

climber:
	CGO_ENABLED=0 go build $(GOFLAGS) -o bin/climber climber/main.go

mctl:
	CGO_ENABLED=0 go build $(GOFLAGS) -o bin/mctl mctl/main.go

dist: mtop climber mctl
	tar -C bin/ -jcvf dist/mtop_linux64.tar.bz2 .

install: all
	mkdir -p $(PREFIX)/bin/
	mkdir -p $(PREFIX)/etc/mtop/
	install bin/* $(PREFIX)/bin/
	install --mode=644 script/mtop.yaml $(PREFIX)/etc/mtop/
	install --mode=644 script/climber.yaml $(PREFIX)/etc/mtop/

uninstall:
	rm $(PREFIX)/bin/mtop
	rm $(PREFIX)/bin/mctl
	rm $(PREFIX)/bin/climber
	rm -r $(PREFIX)/etc/mtop/

install-systemd:
	cat script/mtop.service | MTOP_PATH=$(PREFIX)/bin/mtop envsubst > /etc/systemd/system/mtop.service
	systemctl daemon-reload
