package output

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
)

func PrintTable(out io.Writer, headers []string, rows [][]string) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, strings.Join(headers, "\t"))
	_, _ = fmt.Fprintln(w, strings.Repeat("-\t", len(headers)))
	for _, row := range rows {
		_, _ = fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	_ = w.Flush()
}

func SortKeys(keys []string) {
	sort.Strings(keys)
}
