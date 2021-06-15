// Backfiller assists with backfilling via promtool.
package backfiller

import (
	"fmt"
	"time"
)

func Help(metric string) string {
	return fmt.Sprintf("# HELP %s The metric.\n# TYPE %s gauge\n", metric, metric)
}

func FromUnixMs(ms int64) time.Time {
	return time.Unix(ms/int64(1000), (ms%int64(1000))*int64(1000000))
}
