FROM golang:1.16-alpine as builder

RUN apk add --no-cache make

WORKDIR /go/src/github.com/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /process-version ./cmd/process-version

FROM alpine:latest  
COPY --from=builder /process-version /process-version

CMD /process-version
