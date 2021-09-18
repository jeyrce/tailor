FROM registry.woqutech.com/library/golang:1.16.4 as builder
WORKDIR /GoProject/src/github.com/woqutech/tailor
ENV GOPATH=/GoProject
ENV GOPROXY=https://goproxy.io
ARG archive
COPY ./ /GoProject/src/github.com/woqutech/tailor/
RUN make build archive=${archive}

FROM registry.woqutech.com/google_containers/alpine:3.10 as prod
MAINTAINER jianxin.lu<jianxin.lu@woqutech.com>
LABEL description="tailor: 管理promtail的配置文件"
ARG ImageVersion
ARG archive
COPY --from=builder /GoProject/src/github.com/woqutech/tailor/_output/tailor-${ImageVersion}-${archive} /app/tailor
EXPOSE 15100:15100
ENTRYPOINT ["/app/tailor"]
