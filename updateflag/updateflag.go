package updateflag

import (
	"flag"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

var update string

func registerFlags(fs interface {
	StringVar(*string, string, string, string)
}, updateFlagName string) {
	fs.StringVar(&update,
		updateFlagName,
		os.Getenv("GOKRAZY_UPDATE"),
		`URL of a gokrazy installation (e.g. http://gokrazy:mypassword@myhostname/) to work with. The special value "yes" uses the stored password and -hostname value to construct the URL`)
}

func RegisterFlags(fs *flag.FlagSet, updateFlagName string) {
	registerFlags(fs, updateFlagName)
}

func RegisterPflags(fs *pflag.FlagSet, updateFlagName string) {
	registerFlags(fs, updateFlagName)
}

func GetUpdateTarget(hostname string) (defaultPassword, updateHostname string) {
	if update == "" {
		// -update not set
		return "", hostname
	}
	if update == "yes" {
		// -update=yes
		return "", hostname
	}
	if strings.HasPrefix(update, ":") {
		// port number syntax, e.g. -update=:2080
		return "", hostname
	}
	// -update=<url> syntax
	u, err := url.Parse(update)
	if err != nil {
		return "", hostname
	}
	defaultPassword, _ = u.User.Password()
	return defaultPassword, u.Host
}

func BaseURL(httpPort, httpsPort, schema, hostname, pw string) (*url.URL, error) {
	if update != "yes" && !strings.HasPrefix(update, ":") {
		// already fully qualified, nothing to add
		return url.Parse(update)
	}
	port := httpPort
	defaultPort := "80"
	if schema == "https" {
		port = httpsPort
		defaultPort = "443"
	}
	if strings.HasPrefix(update, ":") {
		port = strings.TrimPrefix(update, ":")
	}
	update = schema + "://gokrazy:" + pw + "@" + hostname
	if port != defaultPort {
		update += ":" + port
	}
	update += "/"
	return url.Parse(update)
}

func NewInstallation() bool {
	return update == ""
}

func SetUpdate(u string) { update = u }

func GetUpdate() string { return update }
