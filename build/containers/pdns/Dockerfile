# This container is responsible for both building the application and providing an image that can be pushed to an
# upstream repository.
#
# To build this image run the following commands:
#
#   $ docker build \
#         --tag gcr.io/andrewhowdencom/pdns:latest \
#         --file path/to/this/Dockerfile \
#         $(git rev-parse --show-toplevel)
#
# The container is written expecting the entire project to be pushed to the docker server as the build context.

# Application Compilation
FROM golang:1.13.0-buster

WORKDIR /mnt/src

COPY . /mnt

RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
        -a \
        -ldflags '-extldflags "-static"' \
        -o ../bin/pdns

# Deployable Artifact
FROM scratch

COPY --from=0 /mnt/bin/pdns pdns
COPY --from=0 /etc/ssl/certs /etc/ssl/certs

ENTRYPOINT [ "/pdns" ]