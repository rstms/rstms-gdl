# go makefile

os_name := $(shell uname -s)
os_version := $(shell uname -r | tr -d .)
os_arch := $(shell uname -m)

binary := $(notdir $(shell pwd))
version := $(shell cat VERSION)
latest_release = $(shell gh release ls | awk '{print $$1;exit}')
ports = $(shell boxen ps -n | sed -n '/^port[0-9][0-9]/s/^port//p')

howdy:
	@echo os_name=$(os_name)
	@echo os_version=$(os_version)
	@echo os_arch=$(os_arch)
	@echo binary=$(binary)
	@echo latest_release=$(latest_release)
	@echo ports=$(ports)

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

dist: $(binary)
	$(foreach p,$(ports),scp gdl.go port$(p):. && ssh port$(p) go build gdl.go && scp port$(p):gdl dist/gdl$(p) && ssh port$(p) rm gdl && ssh port$(p) rm gdl.go;)


dist/$(release_binary): $(binary)
	mkdir -p dist
	scp $< $(dist_target)/$(release_binary)
	scp $< $(dist_target)/$(dist_binary)
	cp $< $@

release-upload: dist
	cd dist; gh release upload $(latest_release) $(release_binary) $(CLOBBER)

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
