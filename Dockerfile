# Build
FROM golang:alpine AS build

ARG TAG
ARG BUILD

ENV LIBRARY je
ENV SERVER je
ENV CLIENT job
ENV REPO prologic/$LIBRARY

RUN apk add --update git make build-base && \
    rm -rf /var/cache/apk/*

WORKDIR /go/src/github.com/$REPO
COPY . /go/src/github.com/$REPO
RUN make deps
RUN make TAG=$TAG BUILD=$BUILD build

RUN cd samples && \
    gcc -o hello hello.c

# Runtime
FROM alpine

ENV LIBRARY je
ENV SERVER je
ENV CLIENT job
ENV REPO prologic/$LIBRARY

LABEL msgbud.app main

COPY --from=build /go/src/github.com/${REPO}/samples /samples
COPY --from=build /go/src/github.com/${REPO}/cmd/${SERVER}/${SERVER} /${SERVER}
COPY --from=build /go/src/github.com/${REPO}/cmd/${CLIENT}/${CLIENT} /${CLIENT}

EXPOSE 8000/tcp

VOLUME /data

ENTRYPOINT ["/je"]
CMD ["--help"]
