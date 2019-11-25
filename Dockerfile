FROM golang:1.13.4-alpine

RUN apk add --no-cache git; \
    go get -u github.com/kardianos/govendor

ENV WD=${GOPATH}/src/goci
COPY ./ ${WD}

WORKDIR ${WD}

RUN PATH=$PATH:$GOPATH/bin

# install dependecies
RUN if [ ! -d "vendor" ]; then \
        govendor init; \
    fi; \
    govendor sync +v; \
    govendor fetch +m

# build app
RUN go build -o bin/goci .

# -------------------------
# build goci image
#--------------------------
FROM docker:19.03-git

RUN apk add --update make
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY --from=0 /go/src/goci/bin/goci /usr/local/bin/goci
COPY docker-entrypoint.sh /usr/local/bin/ 

ENTRYPOINT [ "docker-entrypoint.sh" ]

CMD ["goci" ]
