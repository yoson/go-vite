package verifier

import "github.com/vitelabs/go-vite/metrics"

var notExistType = metrics.GetOrRegisterMeter("/verifier/notExistType", metrics.BranchRegistry)
