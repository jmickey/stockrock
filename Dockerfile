FROM golang:1.20.1 as builder

WORKDIR /app
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o bin .

FROM alpine:3
RUN apk add --no-cache tzdata

COPY --from=builder /app/bin /app/bin

EXPOSE 8080

ENTRYPOINT ["/app/bin"]