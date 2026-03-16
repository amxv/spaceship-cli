package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
)

func PrintJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func PrintDomainListTable(w io.Writer, payload map[string]any) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "DOMAIN\tUNICODE\tEXPIRATION\tSTATUS")
	items, _ := payload["items"].([]any)
	for _, item := range items {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\n",
			toString(row["name"]),
			toString(row["unicodeName"]),
			toString(row["expirationDate"]),
			toString(row["lifecycleStatus"]),
		)
	}
	_ = tw.Flush()

	if total, ok := payload["total"]; ok {
		fmt.Fprintf(w, "Total: %v\n", total)
	}
}

func PrintDNSListTable(w io.Writer, payload map[string]any) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "TYPE\tNAME\tTTL\tVALUE")
	items, _ := payload["items"].([]any)
	for _, item := range items {
		record, ok := item.(map[string]any)
		if !ok {
			continue
		}
		fmt.Fprintf(
			tw,
			"%s\t%s\t%s\t%s\n",
			toString(record["type"]),
			toString(record["name"]),
			toString(record["ttl"]),
			recordValue(record),
		)
	}
	_ = tw.Flush()

	if total, ok := payload["total"]; ok {
		fmt.Fprintf(w, "Total: %v\n", total)
	}
}

func toString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	default:
		return fmt.Sprintf("%v", v)
	}
}

func recordValue(record map[string]any) string {
	priority := []string{
		"address", "cname", "value", "exchange", "nameserver", "pointer", "aliasName", "target", "targetName",
	}
	for _, k := range priority {
		if v, ok := record[k]; ok {
			return toString(v)
		}
	}

	ignored := map[string]bool{"type": true, "name": true, "ttl": true, "group": true}
	keys := make([]string, 0, len(record))
	for k := range record {
		if ignored[k] {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return ""
	}

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, record[k]))
	}
	return strings.Join(parts, ",")
}
