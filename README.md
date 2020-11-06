[![GitHub Release](https://img.shields.io/github/v/release/mattolenik/cloudflare-ddns-client?label=Release&logo=github&logoColor=white)](https://github.com/mattolenik/cloudflare-ddns-client/releases)
[![Docker Tag](https://img.shields.io/docker/v/mattolenik/cloudflare-ddns-client?color=blue&label=Docker%20Tag&logo=docker&logoColor=white)](https://hub.docker.com/repository/docker/mattolenik/cloudflare-ddns-client)

![Functional Tests](https://img.shields.io/github/workflow/status/mattolenik/cloudflare-ddns-client/Functional%20Tests?label=Functional%20Tests)

![License](https://img.shields.io/github/license/mattolenik/cloudflare-ddns-client?color=blue&label=License)

# cloudflare-ddns-client
This is a DDNS client for CloudFlare, written in Go. It makes multiple attemps to retrieve your public IP, first using DNS and then several public APIs. It is well tested and very robust, and will make several retries before failing.

All it requires from you are your domain name, e.g. `mydomain.com`, your DNS record name, e.g. `mydomain.com` or `sub.mydomain.com`, and a CloudFlare API token with the permissions of `Zone:Zone:Read` and `Zone:DNS:Edit`.

`cloudflare-ddns` will attempt to resolve your public IP in the following order:
 1. Using OpenDNS
 2. Using Google DNS
 3. Using `http://whatismyip.akamai.com`
 4. Using `https://ipecho.net/plain`
 5. Using `https://wtfismyip.com/text`

## Example
```console
$ cloudflare-ddns --domain mydomain.com --record sub.mydomain.com --token <cloudflare-api-token>
11:16PM INF Found external IP '97.113.235.123'
11:16PM INF DNS record 'sub.mydomain.com' is already set to IP '97.113.235.123'
```

## Installation
Binaries can be downloaded from the GitHub Releases page. They are statically compiled and should run in any Linux distro.

## Configuration

`cloudflare-ddns` can take configuration either through config file, command-line arguments, or environment variable. Use whichever method you feel is easiest for your use case. Be careful when passing in your CloudFlare API token as a CLI argument, it may be visible in logs if you are running the program from a cron job, systemd, etc.

### Configuration File
Configuration can be provided in any of the formats supported by the [viper configuration library](https://github.com/spf13/viper), including JSON, YAML, and TOML.

Example TOML configuration file:
```toml
# Configuration for cloudflare-ddns, TOML format.

# The domain name that you own
domain = "example.com"

# The DNS record to update, may be a subdomain or the domain name itself
record = "subdomain.example.com"

# Your CloudFlare API token, must have permissions Zone:Zone:Read, Zone:DNS:Edit
token = "your-cloudflare-api-token-here"
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

## Command-Line Usage
```
A dynamic DNS client for CloudFlare. Automatically detects your public IP and
creates/updates a DNS record in CloudFlare.

Configuration flags can be set by defining an environment variable of the same
name but prefixed with `CLOUDFLARE_DDNS`.
For example, `CLOUDFLARE_DDNS_DOMAIN=mydomain.com` can be used to set the domain flag.

Usage:
  cloudflare-ddns [flags]

Flags:
      --config string   Path to config file (default is /etc/cloudflare-ddns.toml)
      --domain string   Domain name in CloudFlare, e.g. example.com
  -h, --help            help for cloudflare-ddns
      --json            Log format, either pretty or json, defaults to pretty
      --record string   DNS record name in CloudFlare, may be subdomain or same as domain
      --token string    CloudFlare API token with permissions Zone:Zone:Read and Zone:DNS:Edit
  -v, --verbose         Verbose logging, prints additional log output
      --version         version for cloudflare-ddns

```
