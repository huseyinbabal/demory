FROM golang:1.18-alpine as builder
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o demory ./cmd/main.go

FROM scratch
COPY --from=builder /go/src/app/demory /demory
ENTRYPOINT ["/demory"]
