APP_NAME := 快递运费报价
RSRC := $(shell go env GOPATH)/bin/rsrc

.PHONY: build clean

build:
	mkdir -p build
	$(RSRC) -manifest app.manifest -o rsrc.syso
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
	go build -ldflags="-H windowsgui -s -w" -o build/$(APP_NAME).exe .
	rm -f rsrc.syso

clean:
	rm -rf build rsrc.syso
