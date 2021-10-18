FROM golang:alpine as builder
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o demory .

FROM scratch
COPY --from=builder /go/src/app/demory /demory
ENTRYPOINT ["/demory"]