# go makefile

netboot_url := https://zippy.rstms.net:4443

binary := $(notdir $(shell pwd))
version := $(shell cat VERSION)

default: build

build: $(binary)

$(binary): fmt
	fix go build . ./...
	go build

fmt: go.sum
	fix go fmt . ./...

go.mod:
	go mod init

go.sum: go.mod
	go mod tidy

install: build
	go install

test: fmt
	go test -v -failfast . ./...

debug: fmt
	go test -v -failfast -count=1 -run $(test) . ./...

release: 
	$(gitclean)
	@$(if $(update),gh release delete -y v$(version),)
	gh release create v$(version) --notes "v$(version)"
	$(MAKE) build

netboot-upload: $(binary)
	./upload-netboot-binaries https://zippy.rstms.net:4443

clean:
	rm -f $(binary) *.core 
	go clean
	rm -rf dist && mkdir dist

sterile: clean
	-[ -e ~/go/bin/$(binary) ] && go clean -i 
	-go clean
	-go clean -modcache
	-go clean -cache
	rm -f go.mod go.sum
