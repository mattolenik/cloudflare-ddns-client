FROM alpine
RUN apk add --no-cache ca-certificates
COPY dist/cloudflare-ddns-linux-amd64 /bin/cloudflare-ddns
ENTRYPOINT [ "/bin/cloudflare-ddns" ]