VERSION=0.1.0
PATH_BUILD=build/
FILE_ARCH=linux_arm64
FILE_COMMAND=psb-rotatord

clean:
	@rm -rf ./build

build: clean
	@$(GOPATH)/bin/goxc \
	  -bc="linux,amd64 linux,arm64" \
	  -pv=$(VERSION) \
	  -d=$(PATH_BUILD) \
	  -build-ldflags "-X main.VERSION=$(VERSION)"

version:
	@echo $(VERSION)

install:
	install $(PATH_BUILD)$(VERSION)/$(FILE_ARCH)/$(FILE_COMMAND) '/usr/bin/$(FILE_COMMAND)'