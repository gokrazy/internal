package updateflag

import (
	"flag"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

var update string

func RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&update,
		"update",
		os.Getenv("GOKRAZY_UPDATE"),
		`URL of a gokrazy installation (e.g. http://gokrazy:mypassword@myhostname/) to update. The special value "yes" uses the stored password and -hostname value to construct the URL`)
}

func RegisterPflags(fs *pflag.FlagSet) {
	fs.StringVar(&update,
		"update",
		os.Getenv("GOKRAZY_UPDATE"),
		`URL of a gokrazy installation (e.g. http://gokrazy:mypassword@myhostname/) to update. The special value "yes" uses the stored password and -hostname value to construct the URL`)
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

func BaseURL(httpPort, schema, hostname, pw string) (*url.URL, error) {
	if update != "yes" && !strings.HasPrefix(update, ":") {
		// already fully qualified, nothing to add
		return url.Parse(update)
	}
	port := httpPort
	if strings.HasPrefix(update, ":") {
		port = strings.TrimPrefix(update, ":")
	}
	update = schema + "://gokrazy:" + pw + "@" + hostname
	if port != "80" {
		update += ":" + port
	}
	update += "/"
	return url.Parse(update)
}

func NewInstallation() bool {
	return update == ""
}

func SetUpdate(u string) {
	update = u
}
