# 编译时间
COMPILE_TIME = $(shell date +"%Y%m%d%H%M%S")

# GIT版本号
GIT_REVISION = $(shell git show -s --pretty=format:%h)

# 编译路径
BUILD_PATH = $(shell pwd)/bin/$(COMPILE_TIME)_$(GIT_REVISION)

build:dep
	go build main.go

release:dep
	mkdir -p $(BUILD_PATH)
	gox -osarch="darwin/amd64 linux/386 linux/amd64 linux/arm linux/arm64" -output="$(BUILD_PATH)/course_{{.OS}}_{{.Arch}}"

setup:clean
	go get github.com/golang/dep/cmd/dep
	go get github.com/mitchellh/gox

dep:setup
	dep ensure

clean:
	rm -rf sues-go
	rm -rf main
	rm -rf bin/*