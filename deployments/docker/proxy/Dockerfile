FROM golang:1.25-alpine AS build

WORKDIR /src

COPY . .

RUN go build -o /out/authsentinel-proxy ./cmd/proxy

FROM alpine:3.19

COPY --from=build /out/authsentinel-proxy /usr/local/bin/authsentinel-proxy

EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/authsentinel-proxy"]
