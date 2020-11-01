FROM alpine
COPY dist/cloudflare-ddns-linux-amd64 /bin/cloudflare-ddns
ENTRYPOINT [ "/bin/cloudflare-ddns" ]