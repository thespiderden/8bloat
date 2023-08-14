package conf

import "flag"

var (
	file      = *(flag.String("f", "", `config file, use a dash for stdin`))
	writeConf = *(flag.Bool("wc", false, `write a sample configuration file to stdout`))
)
