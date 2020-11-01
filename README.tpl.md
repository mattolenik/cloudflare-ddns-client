# cloudflare-ddns-client
CloudFlare DDNS client, written in Go

## Usage
```
{{ run "go" "run" "main.go" "-help" }}
```

## Running with Docker
```
docker run --rm mattolenik/cloudflare-ddns-client -v path-to-cloudflare-ddns.conf:/etc/cloudflare-ddns.conf
```