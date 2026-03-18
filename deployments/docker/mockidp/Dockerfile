FROM golang:1.25-alpine AS build

WORKDIR /src
COPY . .
RUN go build -o /out/authsentinel-mockidp ./cmd/mockidp

FROM alpine:3.19
COPY --from=build /out/authsentinel-mockidp /usr/local/bin/authsentinel-mockidp
EXPOSE 8082
ENTRYPOINT ["/usr/local/bin/authsentinel-mockidp"]

