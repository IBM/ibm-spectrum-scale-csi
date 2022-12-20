package utils

import (
	"flag"
)

func InitLogger() {
	_ = flag.Set("alsologtostderr", "false")
	_ = flag.Set("stderrthreshold", "INFO")
	_ = flag.Set("v", "4")
	_ = flag.Set("log_dir", "/var/log/")
	flag.Parse()
}
