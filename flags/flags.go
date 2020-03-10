package flags

import (
	"flag"
)

var (
	Help       bool
	Version    bool
	ConfigPath string
)

func init() {
	flag.BoolVar(&Help,"h",false,"help information")
	flag.BoolVar(&Version, "v", false, "BookingBot version")
	flag.StringVar(&ConfigPath, "conf", "./config/config.json", "configuration file path")
}
