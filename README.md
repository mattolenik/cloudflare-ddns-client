# cloudflare-ddns-client
CloudFlare DDNS client, written in Go

## Usage
```
Usage of cloudflare-ddns:
  -config string
    	Path to configuration file (default &#34;/etc/cloudflare-ddns.conf&#34;)
  -log-format string
    	Log output format, either json or pretty (default &#34;pretty&#34;)
  -verbose
    	Print debug logs
  -version
    	Print the program version

```

## Running with Docker
```
docker run --rm mattolenik/cloudflare-ddns-client -v path-to-cloudflare-ddns.conf:/etc/cloudflare-ddns.conf
```