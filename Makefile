NAME=clash-core
BUILDDIR=build/releases
VERSION=$(shell git describe --tags || echo v0.0.0-test)
BUILDTIME=$(shell date -u)
GOBUILD=CGO_ENABLED=0 go generate && go build -trimpath -ldflags '-X "github.com/Dreamacro/clash/constant.Version=$(VERSION)" \
		-X "github.com/Dreamacro/clash/constant.BuildTime=$(BUILDTIME)" \
		-w -s -buildid='

LINUX_ARCH_LIST = \
	linux-386 \
	linux-amd64 \
	linux-armv7 \
	linux-armv8 \
	linux-mips-softfloat \
	linux-mips-hardfloat \
	linux-mipsle-softfloat \
	linux-mipsle-hardfloat \
	linux-mips64 \
	linux-mips64le

WINDOWS_ARCH_LIST = \
	windows-386 \
	windows-amd64 \
	windows-arm64 \
	windows-arm32v7

build_dep:
	go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest

all: linux-amd64 windows-amd64 # Most used

linux-386:
	GOARCH=386 GOOS=linux $(GOBUILD) -o $(BUILDDIR)/../$(NAME)-$@

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BUILDDIR)/../$(NAME)-$@

linux-armv7:
	GOARCH=arm GOOS=linux GOARM=7 $(GOBUILD) -o $(BUILDDIR)/../$(NAME)-$@

linux-armv8:
	GOARCH=arm64 GOOS=linux $(GOBUILD) -o $(BUILDDIR)/../$(NAME)-$@

linux-mips-softfloat:
	GOARCH=mips GOMIPS=softfloat GOOS=linux $(GOBUILD) -o $(BUILDDIR)/../$(NAME)-$@

linux-mips-hardfloat:
	GOARCH=mips GOMIPS=hardfloat GOOS=linux $(GOBUILD) -o $(BUILDDIR)/../$(NAME)-$@

linux-mipsle-softfloat:
	GOARCH=mipsle GOMIPS=softfloat GOOS=linux $(GOBUILD) -o $(BUILDDIR)/../$(NAME)-$@

linux-mipsle-hardfloat:
	GOARCH=mipsle GOMIPS=hardfloat GOOS=linux $(GOBUILD) -o $(BUILDDIR)/../$(NAME)-$@

linux-mips64:
	GOARCH=mips64 GOOS=linux $(GOBUILD) -o $(BUILDDIR)/../$(NAME)-$@

linux-mips64le:
	GOARCH=mips64le GOOS=linux $(GOBUILD) -o $(BUILDDIR)/../$(NAME)-$@

windows-386:
	GOARCH=386 GOOS=windows $(GOBUILD) -o $(BUILDDIR)/../$(NAME)-$@.exe

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(BUILDDIR)/../$(NAME)-$@.exe

windows-arm64:
	GOARCH=arm64 GOOS=windows $(GOBUILD) -o $(BUILDDIR)/../$(NAME)-$@.exe

windows-arm32v7:
	GOARCH=arm GOOS=windows GOARM=7 $(GOBUILD) -o $(BUILDDIR)/../$(NAME)-$@.exe

linux_zip_releases=$(addsuffix .zip, $(LINUX_ARCH_LIST))

$(linux_zip_releases): %.zip : %
	makecab $(BUILDDIR)/../$(NAME)-$(basename $@) $(BUILDDIR)/$(NAME)-$(basename $@)-$(VERSION).zip

windows_zip_releases=$(addsuffix .zip, $(WINDOWS_ARCH_LIST))

$(windows_zip_releases): %.zip : %
	makecab $(BUILDDIR)/../$(NAME)-$(basename $@).exe $(BUILDDIR)/$(NAME)-$(basename $@)-$(VERSION).zip

all-arch: $(LINUX_ARCH_LIST) $(WINDOWS_ARCH_LIST)

releases: $(linux_zip_releases) $(windows_zip_releases)

lint:
	GOOS=darwin golangci-lint run ./...
	GOOS=windows golangci-lint run ./...
	GOOS=linux golangci-lint run ./...
	GOOS=freebsd golangci-lint run ./...
	GOOS=openbsd golangci-lint run ./...

clean:
	rm $(BUILDDIR)/*
	rm $(BUILDDIR)/../*
