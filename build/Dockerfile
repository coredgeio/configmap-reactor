FROM golang:1.20 as builder

WORKDIR /workspace

COPY ./ .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GO111MODULE=on go build -a -o configmap-reactor cmd/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/configmap-reactor .
USER 65532:65532

ENTRYPOINT ["/configmap-reactor"]
