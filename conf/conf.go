package conf

// Version is populated by the ldflags argument during build.
var Version string = "TBD" // Set to TDB so it's visible when using go run

// ModuleName is the name of the Go module of this project, passed in by ldflags during build.
var ModuleName string

var (
	// Config is the path to the config file, if present
	Config = "config"
	// Daemon is the flag for enabling daemon mode
	Daemon = "daemon"
	// Domain is the domain to update within CloudFlare
	Domain = "domain"
	// Record is the DNS record to update within CloudFlare, may be same as Domain or a subdomain
	Record = "record"
	// Token is the CloudFlare API token
	Token = "token"
	// JSONOutput indicates that logging should be in JSON format instead of pretty console format
	JSONOutput = "json"
	// Verbose enables additional log output
	Verbose = "verbose"
	// VerboseShort is the short flag for verbose
	VerboseShort = "v"
)
