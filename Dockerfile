# build stage
FROM golang AS builder

WORKDIR /src/aws-signing/
COPY . .
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.Version=`cat VERSION`" -a -installsuffix cgo -o dist/aws-reverse-proxy ./aws-reverse-proxy/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.Version=`cat VERSION`" -a -installsuffix cgo -o dist/aws-curl ./aws-curl/

# final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /src/aws-signing/dist/aws-curl .
COPY --from=builder /src/aws-signing/dist/aws-reverse-proxy .
RUN chmod +x ./aws-curl
RUN chmod +x ./aws-reverse-proxy

EXPOSE 80
ENTRYPOINT ["./aws-reverse-proxy"]
CMD ["--port=80"]
