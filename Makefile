export GO111MODULE=on
export GOPROXY="https://mirrors.aliyun.com/goproxy,https://goproxy.io,direct"
LDFLAGS := -s -w

DATE=$(shell date +"%Y-%m-%d")
BUILDINFO := -X main.Version=latest -X main.Date=$(DATE)

all:
	env CGO_ENABLED=0 go build -trimpath -ldflags '-w -s $(BUILDINFO)' -o majora

clean:
	rm -fr majora-go
	rm -fr majora
	rm -fr output