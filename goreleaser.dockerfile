FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY git-sync-controller /manager
USER 65532:65532

ENTRYPOINT ["/manager"]
