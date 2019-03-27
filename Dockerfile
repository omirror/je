# Build
FROM prologic/go-builder:latest AS build

ARG TAG
ARG BUILD

WORKDIR /src/je
COPY . /src/je
RUN make TAG=$TAG BUILD=$BUILD build
RUN cd samples && \
    gcc -o hello hello.c

# Runtime
FROM alpine

COPY --from=build /src/je/samples /samples
COPY --from=build /src/je/cmd/je /je
COPY --from=build /src/je/cmd/job /job

EXPOSE 8000/tcp

VOLUME /data

ENTRYPOINT ["/je"]
CMD ["-datadir", "/data"]
