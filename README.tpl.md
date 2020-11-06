[![GitHub Release](https://img.shields.io/github/v/release/mattolenik/cloudflare-ddns-client?label=Release&logo=github&logoColor=white)](https://github.com/mattolenik/cloudflare-ddns-client/releases)
[![Docker Tag](https://img.shields.io/docker/v/mattolenik/cloudflare-ddns-client?color=blue&label=Docker%20Tag&logo=docker&logoColor=white)](https://hub.docker.com/repository/docker/mattolenik/cloudflare-ddns-client)

![Functional Tests](https://img.shields.io/github/workflow/status/mattolenik/cloudflare-ddns-client/Functional%20Tests?label=Functional%20Tests)
[![Code Coverage](https://img.shields.io/coveralls/github/mattolenik/cloudflare-ddns-client?label=Code%20Coverage)](https://coveralls.io/github/mattolenik/cloudflare-ddns-client?branch=main)

![Platforms](https://img.shields.io/badge/Platforms-Linux%2C%20macOS%2C%20Windows%2C%20BSD-blue)

![License](https://img.shields.io/github/license/mattolenik/cloudflare-ddns-client?color=blue&label=License)

# Better Dynamic DNS for your Home Lab
This is a cross-platform DDNS client for CloudFlare, written in Go. It makes multiple attemps to retrieve your public IP, first using DNS and then several public APIs. It is well tested and very robust, and will make several retries before failing. It also has good logging so you can see it working in your system logs.

All it requires from you are your domain name, e.g. `mydomain.com`, your DNS record name, e.g. `mydomain.com` or `sub.mydomain.com`, and a CloudFlare API token with the permissions of `Zone:Zone:Read` and `Zone:DNS:Edit`.

It can also print logs in JSON for consumption by logging tools that process JSON. Just use the `--json` flag.

`cloudflare-ddns` will attempt to resolve your public IP in the following order:
 1. Using OpenDNS
 2. Using Google DNS
 3. Using `http://whatismyip.akamai.com`
 4. Using `https://ipecho.net/plain`
 5. Using `https://wtfismyip.com/text`

## Running It
By passing in all required arguments:
```console
$ cloudflare-ddns --domain mydomain.com --record sub.mydomain.com --token <cloudflare-api-token>
11:16PM INF Found external IP '97.113.235.123'
11:16PM INF DNS record 'sub.mydomain.com' is already set to IP '97.113.235.123'
```

If you have a config file set up (see below), no arguments are needed:
```console
$ cloudflare-ddns
11:16PM INF Found external IP '97.113.235.123'
11:16PM INF DNS record 'sub.mydomain.com' is already set to IP '97.113.235.123'
```

With environment variables:
```console
DOMAIN=mydomain.com RECORD=sub.mydomain.com TOKEN=<your-cloudflare-api-token> cloudflare-ddns
11:16PM INF Found external IP '97.113.235.123'
11:16PM INF DNS record 'sub.mydomain.com' is already set to IP '97.113.235.123'
```

### Running with Docker

With a configuration file:
```sh
docker run --rm -v /path/to/cloudflare-ddns.conf:/etc/cloudflare-ddns.conf mattolenik/cloudflare-ddns-client
```

With environment variables:
```sh
docker run --rm -e DOMAIN=mydomain.com -e RECORD=sub.mydomain.com -e TOKEN=<your-cloudflare-api-token> mattolenik/cloudflare-ddns-client
```

## Installation
### As a Single Binary
`cloudflare-ddns` is distributed as a single binary with no dependencies. Simply download the correct binary for your OS under [releases](https://github.com/mattolenik/cloudflare-ddns-client/releases), rename it to `cloudflare-ddns` and place it in a convenient location such as `/usr/local/bin`.

If on Linux:
```sh
curl -sSLo /usr/local/bin/cloudflare-ddns $(curl -s https://api.github.com/repos/mattolenik/cloudflare-ddns-client/releases/latest | awk -F'"' '/browser_download_url.*linux-amd64/ {print $4}') && chmod +x /usr/local/bin/cloudflare-ddns
```

If on macOS:
```sh
curl -sSLo /usr/local/bin/cloudflare-ddns $(curl -s https://api.github.com/repos/mattolenik/cloudflare-ddns-client/releases/latest | awk -F'"' '/browser_download_url.*darwin-amd64/ {print $4}') && chmod +x /usr/local/bin/cloudflare-ddns
```

### Docker
```
docker pull mattolenik/cloudflare-ddns-client
```

### Ubuntu PPA
Coming soon.

### MacOS HomeBrew
Coming soon.

## Configuration

`cloudflare-ddns` can take configuration either through config file, command-line arguments, or environment variable. Use whichever method you feel is easiest for your use case. Be careful when passing in your CloudFlare API token as a CLI argument, it may be visible in logs if you are running the program from a cron job, systemd, etc.

### Configuration File
Configuration can be provided in any of the formats supported by the [viper configuration library](https://github.com/spf13/viper), including JSON, YAML, and TOML.

Example TOML configuration file:
```toml
{{ run "cat" "cloudflare-ddns.toml.example" }}
```

## Running Periodically with Cron
TBD

## Running Periodically with systemd
TBD

## Command-Line Usage
```
{{ run "go" "run" "main.go" "--help" }}
```
