APP_NAME := 快递体积重计算器

.PHONY: build clean dist

build:
	mkdir -p build
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
	CC=x86_64-w64-mingw32-gcc \
	go build -ldflags="-H windowsgui -s -w" -o build/$(APP_NAME).exe .

dist: build
	@echo "确保 build/sciter.dll 存在后运行:"
	@echo "  cp sciter-sdk/bin/64/sciter.dll build/"
	@echo "分发文件: build/$(APP_NAME).exe + build/sciter.dll (UI 已嵌入 exe)"

clean:
	rm -rf build/*
