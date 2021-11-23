
FROM golang:alpine AS staging

WORKDIR /src
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o /app

RUN apk add ca-certificates tzdata

FROM scratch
COPY --from=staging /app /app
COPY --from=staging /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=staging /usr/share/zoneinfo /usr/share/zoneinfo

CMD ["/app"]
