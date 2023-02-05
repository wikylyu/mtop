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

install-mtop: mtop mctl
	mkdir -p $(PREFIX)/bin/
	mkdir -p $(PREFIX)/etc/mtop/
	install bin/mtop $(PREFIX)/bin/
	install bin/mctl $(PREFIX)/bin/
	install --mode=644 script/mtop.yaml $(PREFIX)/etc/mtop/

install-climber: climber
	mkdir -p $(PREFIX)/bin/
	mkdir -p $(PREFIX)/etc/climber/
	install bin/climber $(PREFIX)/bin/
	install --mode=644 script/climber.yaml $(PREFIX)/etc/climber/

uninstall-mtop:
	rm $(PREFIX)/bin/mtop
	rm $(PREFIX)/bin/mctl
	rm -r $(PREFIX)/etc/mtop/

uninstall-climber:
	rm $(PREFIX)/bin/climber
	rm -r $(PREFIX)/etc/climber/

install-mtop-systemd:
	cat script/mtop.service | MTOP_PATH=$(PREFIX)/bin/mtop envsubst > /etc/systemd/system/mtop.service
	systemctl daemon-reload


install-climber-systemd:
	cat script/climber.service | CLIMBER_PATH=$(PREFIX)/bin/climber envsubst > /etc/systemd/system/climber.service
	systemctl daemon-reload
