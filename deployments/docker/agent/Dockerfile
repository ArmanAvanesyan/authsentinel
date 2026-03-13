FROM golang:1.25-alpine AS build

WORKDIR /src

COPY . .

RUN go build -o /out/authsentinel-agent ./cmd/agent

FROM alpine:3.19

COPY --from=build /out/authsentinel-agent /usr/local/bin/authsentinel-agent

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/authsentinel-agent"]
