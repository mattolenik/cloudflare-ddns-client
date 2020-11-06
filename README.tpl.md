# cloudflare-ddns-client
This is a DDNS client for CloudFlare, written in Go. It makes multiple attemps to retrieve your public IP, first using DNS and then several public APIs. It is well tested and very robust, and will make several retries before failing.

All it requires from you are your domain name, e.g. `mydomain.com`, your DNS record name, e.g. `mydomain.com` or `sub.mydomain.com`, and a CloudFlare API token with the permissions of `Zone:Zone:Read` and `Zone:DNS:Edit`.

`cloudflare-ddns` will attempt to resolve your public IP in the following order:
 1. Using OpenDNS
 2. Using Google DNS
 3. Using `http://whatismyip.akamai.com`
 4. Using `https://ipecho.net/plain`
 5. Using `https://wtfismyip.com/text`

## Usage
```
{{ run "go" "run" "main.go" "--help" }}
```

## Installation
Binaries can be downloaded from the GitHub Releases page. They are statically compiled and should run in any Linux distro.

## Configuration

`cloudflare-ddns` can take configuration either through config file, command-line arguments, or environment variable. Use whichever method you feel is easiest for your use case. Be careful when passing in your CloudFlare API token as a CLI argument, it may be visible in logs if you are running the program from a cron job, systemd, etc.

### Configuration File
Configuration can be provided in any of the formats supported by the [viper configuration library](https://github.com/spf13/viper), including JSON, YAML, and TOML.

Example TOML configuration file:
```toml
{{ run "cat" "cloudflare-ddns.toml.example" }}
```

### Running with Docker
There is also a Docker image available for this client.

With a configuration file:
```sh
docker run --rm -v /absolute/path/to/cloudflare-ddns.conf:/etc/cloudflare-ddns.conf mattolenik/cloudflare-ddns-client
```

With a environment variables:
```sh
docker run --rm -e CLOUDFLARE_DDNS_DOMAIN=mydomain.com -e CLOUDFLARE_DDNS_RECORD=sub.mydomain.com -e CLOUDFLARE_DDNS_TOKEN=<your-cloudflare-api-token> mattolenik/cloudflare-ddns-client
```