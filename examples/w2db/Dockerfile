# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25-alpine

FROM golang:${GO_VERSION} AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /out/w2go ./examples/w2db

FROM scratch
WORKDIR /app

COPY --from=build /out/w2go /app/w2go

EXPOSE 3000

ENTRYPOINT ["/app/w2go"]
CMD ["-addr", "0.0.0.0", "-port", "3000"]
