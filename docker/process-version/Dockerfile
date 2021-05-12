FROM golang:1.16-alpine as builder

RUN apk add --no-cache make nodejs npm git

WORKDIR /go/src/github.com/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /process-version ./cmd/process-version

RUN npm install
RUN cp -r node_modules /node_modules

RUN git clone https://github.com/cdnjs/glob.git /glob
RUN npm install /glob

FROM alpine:latest  

RUN apk add --no-cache nodejs jpegoptim zopfli brotli

COPY --from=builder /process-version /process-version
COPY --from=builder /node_modules /node_modules
COPY --from=builder /glob /glob

CMD /process-version
