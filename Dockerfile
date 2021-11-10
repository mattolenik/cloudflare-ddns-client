FROM alpine
ARG TARGETOS
ARG TARGETARCH
RUN apk add --no-cache ca-certificates
COPY dist/cloudflare-ddns-${TARGETOS}-${TARGETARCH} /bin/cloudflare-ddns
ENTRYPOINT [ "/bin/cloudflare-ddns" ]
CMD ["--daemon"]