FROM golang:1.19 as builder
WORKDIR /workspace
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make

FROM alpine
WORKDIR /
COPY --from=builder /workspace/razchess .
COPY --from=builder /workspace/bot .
COPY --from=builder /workspace/uci .
ENTRYPOINT ["/razchess"]
