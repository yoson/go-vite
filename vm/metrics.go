package vm

import (
	"github.com/vitelabs/go-vite/metrics"
)

var vmImpossible = metrics.GetOrRegisterMeter("/impossible/vm", metrics.BranchRegistry)
