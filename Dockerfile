FROM golang:1.18-alpine as builder

RUN apk add --no-cache git make
WORKDIR /app
COPY . .
RUN env CGO_ENABLED=0 go build -a -o shortmarks -ldflags "-s -w"

FROM scratch
ENV PORT 8080
EXPOSE 8080
WORKDIR /
COPY --from=builder /app/shortmarks /
CMD ["/shortmarks"]
