PKGTITLE="ulog"
BUILDID:=$(shell ulidgen|tr '[:upper:]' '[:lower:]')
PKGVERSION="1.0.0.$(BUILDID)"
PKGID=io.micromdm.ulot
SERVER_URL="http://dev.micromdm.io:8080/log"

.PHONY: build

define SERVER_JSON
{
	"url": $(SERVER_URL)
}
endef

export SERVER_JSON

all: build

setup: clean
	mkdir -p ./pkgroot/etc/micromdm/ulog
	mkdir -p ./pkgroot/var/log/micromdm
	mkdir -p ./pkgroot/usr/local/micromdm/bin

clean:
	rm -f ./ulog*.{dmg,pkg}
	rm -f ./pkgroot/usr/local/micromdm/bin/*
	rm -f ./pkgroot/etc/micromdm/ulog/*
	rm -rf build/*

build:
	go build -i -o ./build/ulog ./cmd/ulog

pkg: setup build
	echo "$$SERVER_JSON" > pkgroot/etc/micromdm/ulog/server.json
	cp ./build/ulog ./pkgroot/usr/local/micromdm/bin/ulog
	pkgbuild --root pkgroot --identifier ${PKGID} --version ${PKGVERSION} --ownership recommended ./${PKGTITLE}-${PKGVERSION}.pkg
