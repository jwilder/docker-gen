.SILENT :
.PHONY : docker-gen clean fmt

TAG:=`git describe --tags`
LDFLAGS:=-X main.buildVersion=$(TAG)

all: docker-gen

docker-gen:
	echo "Building docker-gen"
	go build -ldflags "$(LDFLAGS)" ./cmd/docker-gen

dist-clean:
	rm -rf dist
	rm -f docker-gen-alpine-linux-*.tar.gz
	rm -f docker-gen-linux-*.tar.gz
	rm -f docker-gen-darwin-*.tar.gz
	rm -f docker-gen-windows-*.tar.gz

dist: dist-clean
	mkdir -p dist/alpine-linux/amd64 && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -a -tags netgo -installsuffix netgo -o dist/alpine-linux/amd64/docker-gen ./cmd/docker-gen
	mkdir -p dist/alpine-linux/arm64 && GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -a -tags netgo -installsuffix netgo -o dist/alpine-linux/arm64/docker-gen ./cmd/docker-gen
	mkdir -p dist/alpine-linux/armhf && GOOS=linux GOARCH=arm GOARM=6 go build -ldflags "$(LDFLAGS)" -a -tags netgo -installsuffix netgo -o dist/alpine-linux/armhf/docker-gen ./cmd/docker-gen
	mkdir -p dist/linux/amd64 && GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/linux/amd64/docker-gen ./cmd/docker-gen
	mkdir -p dist/linux/arm64 && GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/linux/arm64/docker-gen ./cmd/docker-gen
	mkdir -p dist/linux/i386  && GOOS=linux GOARCH=386 go build -ldflags "$(LDFLAGS)" -o dist/linux/i386/docker-gen ./cmd/docker-gen
	mkdir -p dist/linux/armel  && GOOS=linux GOARCH=arm GOARM=5 go build -ldflags "$(LDFLAGS)" -o dist/linux/armel/docker-gen ./cmd/docker-gen
	mkdir -p dist/linux/armhf  && GOOS=linux GOARCH=arm GOARM=6 go build -ldflags "$(LDFLAGS)" -o dist/linux/armhf/docker-gen ./cmd/docker-gen
	mkdir -p dist/darwin/amd64 && GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/darwin/amd64/docker-gen ./cmd/docker-gen
	mkdir -p dist/darwin/i386  && GOOS=darwin GOARCH=386 go build -ldflags "$(LDFLAGS)" -o dist/darwin/i386/docker-gen ./cmd/docker-gen
	mkdir -p dist/windows/i386  && GOOS=windows GOARCH=386 go build -ldflags "$(LDFLAGS)" -o dist/windows/i386/docker-gen ./cmd/docker-gen
	mkdir -p dist/windows/amd64  && GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/windows/amd64/docker-gen ./cmd/docker-gen


release: dist
	go mod tidy
	tar -cvzf docker-gen-alpine-linux-amd64-$(TAG).tar.gz -C dist/alpine-linux/amd64 docker-gen
	tar -cvzf docker-gen-alpine-linux-arm64-$(TAG).tar.gz -C dist/alpine-linux/arm64 docker-gen
	tar -cvzf docker-gen-alpine-linux-armhf-$(TAG).tar.gz -C dist/alpine-linux/armhf docker-gen
	tar -cvzf docker-gen-linux-amd64-$(TAG).tar.gz -C dist/linux/amd64 docker-gen
	tar -cvzf docker-gen-linux-arm64-$(TAG).tar.gz -C dist/linux/arm64 docker-gen
	tar -cvzf docker-gen-linux-i386-$(TAG).tar.gz -C dist/linux/i386 docker-gen
	tar -cvzf docker-gen-linux-armel-$(TAG).tar.gz -C dist/linux/armel docker-gen
	tar -cvzf docker-gen-linux-armhf-$(TAG).tar.gz -C dist/linux/armhf docker-gen
	tar -cvzf docker-gen-darwin-amd64-$(TAG).tar.gz -C dist/darwin/amd64 docker-gen
	tar -cvzf docker-gen-darwin-i386-$(TAG).tar.gz -C dist/darwin/i386 docker-gen
	tar -cvzf docker-gen-windows-amd64-$(TAG).tar.gz -C dist/windows/amd64 docker-gen
	tar -cvzf docker-gen-windows-i386-$(TAG).tar.gz -C dist/windows/i386 docker-gen

get-deps:
	go get ./...

check-gofmt:
	if [ -n "$(shell gofmt -l .)" ]; then \
		echo 1>&2 'The following files need to be formatted:'; \
		gofmt -l .; \
		exit 1; \
	fi

test:
	go test
