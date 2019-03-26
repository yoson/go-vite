package mobile

import (
	"github.com/vitelabs/go-vite/log15"
)


func DisableAllLog() {
	log15.Root().SetHandler(log15.FilterHandler(func(r *log15.Record) bool {
		return false
	}, log15.StdoutHandler))
}
