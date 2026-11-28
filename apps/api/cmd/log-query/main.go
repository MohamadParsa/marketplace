// log-query is a local administrator's read-only, allowlisted log viewer.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/adapters/locallog"
	"github.com/cherriva/marketplace/apps/api/internal/application/operationlog"
)

func main() {
	dir := flag.String("dir", "", "private local log directory")
	typ := flag.String("type", "", "technical/integration/business/security/behavior")
	op := flag.String("operation", "", "operation identifier")
	outcome := flag.String("outcome", "", "event result")
	from := flag.String("from", "", "RFC3339 start time")
	limit := flag.Int("limit", 100, "maximum events (1..200)")
	cursor := flag.String("cursor", "", "next page cursor")
	flag.Parse()
	reader, err := locallog.OpenReader(locallog.DefaultOptions(*dir))
	if err != nil {
		fmt.Fprintln(os.Stderr, "log storage unavailable")
		os.Exit(1)
	}
	q := operationlog.Query{Type: *typ, OperationID: *op, Outcome: *outcome, Limit: *limit, Cursor: *cursor}
	if *from != "" {
		q.From, err = time.Parse(time.RFC3339, *from)
		if err != nil {
			fmt.Fprintln(os.Stderr, "invalid start time")
			os.Exit(1)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	page, err := operationlog.Read(ctx, reader, q)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid query or log storage unavailable")
		os.Exit(1)
	}
	if err = json.NewEncoder(os.Stdout).Encode(page); err != nil {
		os.Exit(1)
	}
}
