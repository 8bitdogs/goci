FROM golang:1.25.9-alpine AS builder

COPY ./ /app
WORKDIR /app

RUN go mod tidy

# build app
RUN go build -o /app/bin/goci .

# -------------------------
# build goci image
#--------------------------
FROM docker:29.4.1-alpine3.23

COPY --from=builder /app/bin/goci /usr/local/bin/goci
COPY docker-entrypoint.sh /usr/local/bin/ 	

ENTRYPOINT [ "docker-entrypoint.sh" ]

CMD ["goci" ]
