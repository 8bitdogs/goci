FROM golang:1.13.4-alpine

RUN apk add --no-cache git; \
    go get -u github.com/kardianos/govendor

RUN PATH=$PATH:$GOPATH/bin

ENV WD=${GOPATH}/src/github.com/8bitdogs/goci
COPY ./ ${WD}

WORKDIR ${WD}


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

COPY --from=0 /go/src/github.com/8bitdogs/goci/bin/goci /usr/local/bin/goci
COPY docker-entrypoint.sh /usr/local/bin/ 

ENTRYPOINT [ "docker-entrypoint.sh" ]

CMD ["goci" ]
