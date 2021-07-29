BuildDate := $(shell date "+%Y%m%d%H%M%S")
BuildUser := $(shell whoami)
Branch := $(shell git symbolic-ref --short -q HEAD)
CommitID := $(shell git rev-parse --short HEAD)


HarborUsername:=qdata
ImageVersion:=v0.1.0
ImageName:=registry.woqutech.com/${HarborUsername}/tailor:${ImageVersion}


# make
all: build_x86 build_arm upload_x86 upload_arm image

.PHONY: build_x86
build_x86:
	make build archive=amd64

.PHONY: build_arm
build_arm:
	make build archive=arm64

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=${archive} go build -ldflags " \
		-X github.com/prometheus/common/version.Version=${ImageVersion} \
		-X github.com/prometheus/common/version.Revision=${CommitID} \
		-X github.com/prometheus/common/version.Branch=${Branch} \
		-X github.com/prometheus/common/version.BuildUser=${BuildUser} \
		-X github.com/prometheus/common/version.BuildDate=${BuildDate} \
	"  -o _output/tailor-${ImageVersion}-${archive} tailor.go

.PHONY: clean
clean:
	go clean
	rm -rf _output/*


.PHONY: fmt
fmt:
	go fmt
	gofmt -s -w -l .
	goimports -w -l .


.PHONY: image
image:
	sudo docker build --build-arg ImageVersion=$(ImageVersion) --build-arg archive=amd64 -t ${ImageName} .
	sudo docker push ${ImageName}

.PHONY: version
version:
	echo ${ImageVersion}

.PHONY: upload
upload:
	curl -u common:cljslrl0620 -T _output/tailor-${ImageVersion}-${archive}  http://mirrors.woqutech.com/remote.php/dav/files/common/Loki/

.PHONY: upload_x86
upload_x86:
	make upload archive=amd64


.PHONY: upload_arm
upload_arm:
	make upload archive=arm64
