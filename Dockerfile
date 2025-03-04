#####################################################################################################
# Build the manager binary
FROM registry.access.redhat.com/ubi9/go-toolset:1.20 as builder

# Bring in the go dependencies before anything else so we can take
# advantage of caching these layers in future builds.
COPY vendor/ vendor/

# Copy the go modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY api api
COPY controllers controllers
COPY internal internal
COPY ibu-imager ibu-imager
COPY utils utils
COPY main main

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -mod=vendor -a -o build/manager main/main.go


#####################################################################################################
# Build the imager binary
FROM registry.access.redhat.com/ubi9/go-toolset:1.20 as imager

# Bring in the go dependencies before anything else so we can take
# advantage of caching these layers in future builds.
COPY vendor/ vendor/

# Copy the go modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY main main
COPY utils utils
COPY ibu-imager/clusterinfo ibu-imager/clusterinfo
COPY ibu-imager/cmd ibu-imager/cmd
COPY ibu-imager/ops ibu-imager/ops
COPY ibu-imager/seedcreator ibu-imager/seedcreator
COPY ibu-imager/ostreeclient ibu-imager/ostreeclient

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -mod=vendor -a -o build/ibu-imager main/ibu-imager/main.go


#####################################################################################################
# Use this target to develop / test Imager only
FROM registry.access.redhat.com/ubi9/ubi:latest as imager-dev

RUN dnf -y install jq && \
    dnf clean all && \
    rm -rf /var/cache/dnf

COPY --from=imager /opt/app-root/src/build/ibu-imager /usr/local/bin/ibu-imager
COPY ibu-imager/installation_configuration_files/ /usr/local/installation_configuration_files/

ENTRYPOINT ["/usr/local/bin/ibu-imager"]


#####################################################################################################
# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
#FROM gcr.io/distroless/static:nonroot
FROM registry.access.redhat.com/ubi9/ubi:latest

RUN dnf -y install jq && \
    dnf clean all && \
    rm -rf /var/cache/dnf

COPY --from=builder /opt/app-root/src/build/manager /usr/local/bin/manager

COPY --from=imager /opt/app-root/src/build/ibu-imager /usr/local/bin/ibu-imager
COPY ibu-imager/installation_configuration_files/ /usr/local/installation_configuration_files/

ENTRYPOINT ["/usr/local/bin/manager"]
