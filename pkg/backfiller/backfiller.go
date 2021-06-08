// Backfiller assists with backfilling via promtool.
package backfiller

import "fmt"

func Help(metric string) string {
	return fmt.Sprintf("# HELP %s The metric.\n# TYPE %s gauge\n", metric, metric)
}
