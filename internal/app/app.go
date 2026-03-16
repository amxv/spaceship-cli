package app

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ashray/spaceship-cli/internal/client"
	"github.com/ashray/spaceship-cli/internal/credentials"
	"github.com/ashray/spaceship-cli/internal/output"
)

func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		printRootHelp(stdout)
		return nil
	}

	switch args[0] {
	case "auth":
		return runAuth(args[1:], stdout)
	case "domains":
		return runDomains(args[1:], stdout)
	case "dns":
		return runDNS(args[1:], stdout)
	case "help", "-h", "--help":
		printRootHelp(stdout)
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printRootHelp(w io.Writer) {
	fmt.Fprintln(w, "spaceship - simple Spaceship DNS CLI")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  spaceship auth <login|status|logout>")
	fmt.Fprintln(w, "  spaceship domains <list|info>")
	fmt.Fprintln(w, "  spaceship dns <list|set|delete|put>")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  spaceship auth login")
	fmt.Fprintln(w, "  spaceship domains list")
	fmt.Fprintln(w, "  spaceship domains info example.com")
	fmt.Fprintln(w, "  spaceship dns list example.com")
	fmt.Fprintln(w, "  spaceship dns set example.com --type A --name @ --value 1.2.3.4 --ttl 300")
	fmt.Fprintln(w, "  spaceship dns delete example.com --type A --name @ --value 1.2.3.4")
	fmt.Fprintln(w, "  spaceship dns put example.com --file records.json --force=true")
}

func runAuth(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		fmt.Fprintln(stdout, "Usage: spaceship auth <login|status|logout>")
		return nil
	}

	switch args[0] {
	case "login":
		fs := flag.NewFlagSet("auth login", flag.ContinueOnError)
		fs.SetOutput(io.Discard)

		apiKey := fs.String("api-key", "", "Spaceship API key")
		apiSecret := fs.String("api-secret", "", "Spaceship API secret")

		if err := fs.Parse(args[1:]); err != nil {
			return err
		}

		reader := bufio.NewReader(os.Stdin)
		if strings.TrimSpace(*apiKey) == "" {
			v, err := promptLine(stdout, reader, "API key: ")
			if err != nil {
				return err
			}
			*apiKey = v
		}
		if strings.TrimSpace(*apiSecret) == "" {
			v, err := promptLine(stdout, reader, "API secret: ")
			if err != nil {
				return err
			}
			*apiSecret = v
		}

		if strings.TrimSpace(*apiKey) == "" || strings.TrimSpace(*apiSecret) == "" {
			return errors.New("api-key and api-secret are required")
		}

		if err := credentials.Save(strings.TrimSpace(*apiKey), strings.TrimSpace(*apiSecret)); err != nil {
			return err
		}

		fmt.Fprintln(stdout, "Credentials saved to macOS Keychain service \"spaceship-cli\".")
		return nil

	case "status":
		if os.Getenv("SPACESHIP_API_KEY") != "" && os.Getenv("SPACESHIP_API_SECRET") != "" {
			fmt.Fprintln(stdout, "Credentials source: environment variables (SPACESHIP_API_KEY / SPACESHIP_API_SECRET)")
			return nil
		}

		_, _, err := credentials.Load()
		if err != nil {
			if errors.Is(err, credentials.ErrNotFound) {
				fmt.Fprintln(stdout, "No credentials found in env or keychain.")
				return nil
			}
			return err
		}

		fmt.Fprintln(stdout, "Credentials source: macOS Keychain (service spaceship-cli)")
		return nil

	case "logout":
		if err := credentials.Delete(); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "Credentials removed from keychain.")
		return nil

	default:
		return fmt.Errorf("unknown auth command %q", args[0])
	}
}

func runDomains(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		fmt.Fprintln(stdout, "Usage: spaceship domains <list|info>")
		return nil
	}

	c, err := newClientFromStoredCreds()
	if err != nil {
		return err
	}

	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("domains list", flag.ContinueOnError)
		fs.SetOutput(io.Discard)

		take := fs.Int("take", 50, "Number of domains to return (1-100)")
		skip := fs.Int("skip", 0, "Number of domains to skip")
		order := fs.String("order", "", "Sort field: name|-name|unicodeName|-unicodeName|registrationDate|-registrationDate|expirationDate|-expirationDate")
		asJSON := fs.Bool("json", false, "Print raw JSON")

		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *take < 1 || *take > 100 {
			return errors.New("--take must be between 1 and 100")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := c.GetDomainList(ctx, *take, *skip, *order)
		if err != nil {
			return err
		}

		if *asJSON {
			return output.PrintJSON(stdout, resp)
		}
		output.PrintDomainListTable(stdout, resp)
		return nil

	case "info":
		fs := flag.NewFlagSet("domains info", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		asJSON := fs.Bool("json", false, "Print raw JSON")
		domain, err := parseDomainAndFlags(fs, args[1:])
		if err != nil {
			return errors.New("usage: spaceship domains info <domain> [--json]")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := c.GetDomainInfo(ctx, domain)
		if err != nil {
			return err
		}

		if *asJSON {
			return output.PrintJSON(stdout, resp)
		}
		return output.PrintJSON(stdout, resp)

	default:
		return fmt.Errorf("unknown domains command %q", args[0])
	}
}

func runDNS(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		fmt.Fprintln(stdout, "Usage: spaceship dns <list|set|delete|put>")
		return nil
	}

	c, err := newClientFromStoredCreds()
	if err != nil {
		return err
	}

	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("dns list", flag.ContinueOnError)
		fs.SetOutput(io.Discard)

		take := fs.Int("take", 100, "Number of records to return (1-500)")
		skip := fs.Int("skip", 0, "Number of records to skip")
		order := fs.String("order", "", "Sort field: type|-type|name|-name")
		asJSON := fs.Bool("json", false, "Print raw JSON")

		domain, err := parseDomainAndFlags(fs, args[1:])
		if err != nil {
			return errors.New("usage: spaceship dns list <domain> [--take 100 --skip 0 --order name --json]")
		}
		if *take < 1 || *take > 500 {
			return errors.New("--take must be between 1 and 500")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := c.GetResourceRecordsList(ctx, domain, *take, *skip, *order)
		if err != nil {
			return err
		}

		if *asJSON {
			return output.PrintJSON(stdout, resp)
		}
		output.PrintDNSListTable(stdout, resp)
		return nil

	case "set":
		fs := flag.NewFlagSet("dns set", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		cfg := bindRecordFlags(fs)
		cfg.ttl = 3600
		cfg.force = false

		domain, err := parseDomainAndFlags(fs, args[1:])
		if err != nil {
			return errors.New("usage: spaceship dns set <domain> --type A --name @ --value 1.2.3.4 [--ttl 300 --force]")
		}
		item, err := recordFromFlags(cfg, true)
		if err != nil {
			return err
		}

		payload := map[string]any{
			"force": cfg.force,
			"items": []map[string]any{item},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := c.SaveResourceRecords(ctx, domain, payload); err != nil {
			return err
		}

		fmt.Fprintln(stdout, "Record saved.")
		return nil

	case "delete":
		fs := flag.NewFlagSet("dns delete", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		cfg := bindRecordFlags(fs)
		cfg.ttl = -1

		domain, err := parseDomainAndFlags(fs, args[1:])
		if err != nil {
			return errors.New("usage: spaceship dns delete <domain> --type A --name @ --value 1.2.3.4")
		}
		item, err := recordFromFlags(cfg, false)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := c.DeleteResourceRecords(ctx, domain, []map[string]any{item}); err != nil {
			return err
		}

		fmt.Fprintln(stdout, "Record deleted.")
		return nil

	case "put":
		fs := flag.NewFlagSet("dns put", flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		file := fs.String("file", "", "Path to JSON file containing items array or full payload object")
		var force optionalBool
		fs.Var(&force, "force", "Override payload.force (true/false)")

		domain, err := parseDomainAndFlags(fs, args[1:])
		if err != nil || strings.TrimSpace(*file) == "" {
			return errors.New("usage: spaceship dns put <domain> --file records.json [--force=true]")
		}

		payload, err := parsePutPayload(*file, force)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := c.SaveResourceRecords(ctx, domain, payload); err != nil {
			return err
		}
		fmt.Fprintln(stdout, "Records saved.")
		return nil

	default:
		return fmt.Errorf("unknown dns command %q", args[0])
	}
}

func newClientFromStoredCreds() (*client.Client, error) {
	key, secret, err := credentials.Load()
	if err != nil {
		if errors.Is(err, credentials.ErrNotFound) {
			return nil, errors.New("credentials missing: run `spaceship auth login` or set SPACESHIP_API_KEY and SPACESHIP_API_SECRET")
		}
		return nil, err
	}
	return client.New(key, secret), nil
}

func promptLine(stdout io.Writer, reader *bufio.Reader, label string) (string, error) {
	fmt.Fprint(stdout, label)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func parseDomainAndFlags(fs *flag.FlagSet, args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("missing domain")
	}

	if !strings.HasPrefix(args[0], "-") {
		domain := args[0]
		if err := fs.Parse(args[1:]); err != nil {
			return "", err
		}
		if fs.NArg() != 0 {
			return "", errors.New("unexpected extra arguments")
		}
		return domain, nil
	}

	if err := fs.Parse(args); err != nil {
		return "", err
	}
	if fs.NArg() != 1 {
		return "", errors.New("missing domain")
	}
	return fs.Arg(0), nil
}

type recordFlags struct {
	recordType string
	name       string
	ttl        int
	force      bool
	value      string

	address         string
	cname           string
	exchange        string
	preference      int
	nameserver      string
	pointer         string
	aliasName       string
	flag            int
	tag             string
	service         string
	protocol        string
	priority        int
	weight          int
	port            int
	target          string
	scheme          string
	svcPriority     int
	targetName      string
	svcParams       string
	usage           int
	selector        int
	matching        int
	associationData string
	data            pairsFlag
}

func bindRecordFlags(fs *flag.FlagSet) *recordFlags {
	cfg := &recordFlags{preference: -1, flag: -1, priority: -1, weight: -1, port: -1, svcPriority: -1, usage: -1, selector: -1, matching: -1}
	fs.StringVar(&cfg.recordType, "type", "", "Record type (A, AAAA, CNAME, MX, TXT, NS, PTR, ALIAS, CAA, SRV, HTTPS, SVCB, TLSA)")
	fs.StringVar(&cfg.name, "name", "@", "Record name")
	fs.IntVar(&cfg.ttl, "ttl", 3600, "TTL in seconds (60-3600)")
	fs.BoolVar(&cfg.force, "force", false, "Force zone update")
	fs.StringVar(&cfg.value, "value", "", "Common value helper (A/AAAA/TXT/CNAME/NS/PTR/ALIAS)")

	fs.StringVar(&cfg.address, "address", "", "Address value (A/AAAA)")
	fs.StringVar(&cfg.cname, "cname", "", "Canonical name (CNAME)")
	fs.StringVar(&cfg.exchange, "exchange", "", "Mail exchange (MX)")
	fs.IntVar(&cfg.preference, "preference", -1, "MX preference")
	fs.StringVar(&cfg.nameserver, "nameserver", "", "Nameserver (NS)")
	fs.StringVar(&cfg.pointer, "pointer", "", "PTR pointer")
	fs.StringVar(&cfg.aliasName, "alias-name", "", "Alias name (ALIAS)")
	fs.IntVar(&cfg.flag, "flag", -1, "CAA flag")
	fs.StringVar(&cfg.tag, "tag", "", "CAA tag")

	fs.StringVar(&cfg.service, "service", "", "Service name, e.g. _sip (SRV)")
	fs.StringVar(&cfg.protocol, "protocol", "", "Protocol, e.g. _tcp (SRV/TLSA)")
	fs.IntVar(&cfg.priority, "priority", -1, "SRV priority")
	fs.IntVar(&cfg.weight, "weight", -1, "SRV weight")
	fs.IntVar(&cfg.port, "port", -1, "Port (SRV) or numeric port tag (HTTPS/SVCB/TLSA)")
	fs.StringVar(&cfg.target, "target", "", "SRV target")

	fs.StringVar(&cfg.scheme, "scheme", "", "HTTPS/SVCB scheme, e.g. _https")
	fs.IntVar(&cfg.svcPriority, "svc-priority", -1, "HTTPS/SVCB svcPriority")
	fs.StringVar(&cfg.targetName, "target-name", "", "HTTPS/SVCB targetName")
	fs.StringVar(&cfg.svcParams, "svc-params", "", "HTTPS/SVCB svcParams")

	fs.IntVar(&cfg.usage, "usage", -1, "TLSA usage")
	fs.IntVar(&cfg.selector, "selector", -1, "TLSA selector")
	fs.IntVar(&cfg.matching, "matching", -1, "TLSA matching")
	fs.StringVar(&cfg.associationData, "association-data", "", "TLSA associationData")

	fs.Var(&cfg.data, "data", "Additional field as key=value. Repeatable.")
	return cfg
}

func recordFromFlags(cfg *recordFlags, includeTTL bool) (map[string]any, error) {
	rt := strings.ToUpper(strings.TrimSpace(cfg.recordType))
	if rt == "" {
		return nil, errors.New("--type is required")
	}

	item := map[string]any{
		"type": rt,
		"name": strings.TrimSpace(cfg.name),
	}
	if includeTTL {
		if cfg.ttl < 60 || cfg.ttl > 3600 {
			return nil, errors.New("--ttl must be between 60 and 3600")
		}
		item["ttl"] = cfg.ttl
	}

	if err := applyCommonValue(item, rt, strings.TrimSpace(cfg.value)); err != nil {
		return nil, err
	}
	if err := applyExplicitFlags(item, cfg); err != nil {
		return nil, err
	}
	for _, pair := range cfg.data {
		item[pair.Key] = pair.Value
	}
	normalizeRecordFields(item, rt)

	if err := validateRequiredFields(item, rt, includeTTL); err != nil {
		return nil, err
	}
	return item, nil
}

func applyCommonValue(item map[string]any, rt, value string) error {
	if value == "" {
		return nil
	}

	switch rt {
	case "A", "AAAA":
		item["address"] = value
	case "CNAME":
		item["cname"] = value
	case "TXT":
		item["value"] = value
	case "NS":
		item["nameserver"] = value
	case "PTR":
		item["pointer"] = value
	case "ALIAS":
		item["aliasName"] = value
	case "MX", "CAA", "SRV", "HTTPS", "SVCB", "TLSA":
		return fmt.Errorf("--value shortcut is not supported for %s; use type-specific flags", rt)
	default:
		return fmt.Errorf("unsupported record type %s", rt)
	}
	return nil
}

func applyExplicitFlags(item map[string]any, cfg *recordFlags) error {
	if cfg.address != "" {
		item["address"] = cfg.address
	}
	if cfg.cname != "" {
		item["cname"] = cfg.cname
	}
	if cfg.exchange != "" {
		item["exchange"] = cfg.exchange
	}
	if cfg.preference >= 0 {
		item["preference"] = cfg.preference
	}
	if cfg.nameserver != "" {
		item["nameserver"] = cfg.nameserver
	}
	if cfg.pointer != "" {
		item["pointer"] = cfg.pointer
	}
	if cfg.aliasName != "" {
		item["aliasName"] = cfg.aliasName
	}
	if cfg.flag >= 0 {
		item["flag"] = cfg.flag
	}
	if cfg.tag != "" {
		item["tag"] = cfg.tag
	}
	if cfg.service != "" {
		item["service"] = cfg.service
	}
	if cfg.protocol != "" {
		item["protocol"] = cfg.protocol
	}
	if cfg.priority >= 0 {
		item["priority"] = cfg.priority
	}
	if cfg.weight >= 0 {
		item["weight"] = cfg.weight
	}
	if cfg.port >= 0 {
		if _, ok := item["scheme"]; ok {
			item["port"] = fmt.Sprintf("_%d", cfg.port)
		} else {
			item["port"] = cfg.port
		}
	}
	if cfg.target != "" {
		item["target"] = cfg.target
	}
	if cfg.scheme != "" {
		if strings.HasPrefix(cfg.scheme, "_") {
			item["scheme"] = cfg.scheme
		} else {
			item["scheme"] = "_" + cfg.scheme
		}
		if cfg.port >= 0 {
			item["port"] = fmt.Sprintf("_%d", cfg.port)
		}
	}
	if cfg.svcPriority >= 0 {
		item["svcPriority"] = cfg.svcPriority
	}
	if cfg.targetName != "" {
		item["targetName"] = cfg.targetName
	}
	if cfg.svcParams != "" {
		item["svcParams"] = cfg.svcParams
	}
	if cfg.usage >= 0 {
		item["usage"] = cfg.usage
	}
	if cfg.selector >= 0 {
		item["selector"] = cfg.selector
	}
	if cfg.matching >= 0 {
		item["matching"] = cfg.matching
	}
	if cfg.associationData != "" {
		item["associationData"] = cfg.associationData
	}

	return nil
}

func validateRequiredFields(item map[string]any, rt string, includeTTL bool) error {
	required := map[string][]string{
		"A":     {"address"},
		"AAAA":  {"address"},
		"ALIAS": {"aliasName"},
		"CNAME": {"cname"},
		"TXT":   {"value"},
		"NS":    {"nameserver"},
		"PTR":   {"pointer"},
		"MX":    {"exchange", "preference"},
		"CAA":   {"flag", "tag", "value"},
		"SRV":   {"service", "protocol", "priority", "weight", "port", "target"},
		"HTTPS": {"scheme", "port", "svcPriority", "targetName", "svcParams"},
		"SVCB":  {"scheme", "port", "svcPriority", "targetName", "svcParams"},
		"TLSA":  {"port", "protocol", "usage", "selector", "matching", "associationData"},
	}
	fields, ok := required[rt]
	if !ok {
		known := make([]string, 0, len(required))
		for k := range required {
			known = append(known, k)
		}
		sort.Strings(known)
		return fmt.Errorf("unsupported --type %q (supported: %s)", rt, strings.Join(known, ", "))
	}

	for _, field := range fields {
		if _, ok := item[field]; !ok {
			return fmt.Errorf("missing required field for %s: %s", rt, field)
		}
	}

	if includeTTL {
		if _, ok := item["ttl"]; !ok {
			return errors.New("missing ttl")
		}
	}

	return nil
}

func normalizeRecordFields(item map[string]any, rt string) {
	if rt != "HTTPS" && rt != "SVCB" && rt != "TLSA" {
		return
	}

	if raw, ok := item["port"]; ok {
		switch v := raw.(type) {
		case int:
			item["port"] = fmt.Sprintf("_%d", v)
		case float64:
			item["port"] = fmt.Sprintf("_%d", int(v))
		case string:
			if !strings.HasPrefix(v, "_") {
				item["port"] = "_" + v
			}
		}
	}
}

func parsePutPayload(filePath string, force optionalBool) (map[string]any, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	var asObject map[string]any
	if err := json.Unmarshal(data, &asObject); err == nil {
		if _, ok := asObject["items"]; ok {
			if force.set {
				asObject["force"] = force.value
			}
			return asObject, nil
		}
	}

	var asItems []map[string]any
	if err := json.Unmarshal(data, &asItems); err != nil {
		return nil, errors.New("json file must be either an object containing 'items' or an array of record items")
	}

	payload := map[string]any{"items": asItems}
	if force.set {
		payload["force"] = force.value
	}
	return payload, nil
}

type pair struct {
	Key   string
	Value any
}

type pairsFlag []pair

func (p *pairsFlag) String() string {
	if p == nil || len(*p) == 0 {
		return ""
	}
	parts := make([]string, 0, len(*p))
	for _, item := range *p {
		parts = append(parts, fmt.Sprintf("%s=%v", item.Key, item.Value))
	}
	return strings.Join(parts, ",")
}

func (p *pairsFlag) Set(v string) error {
	key, value, ok := strings.Cut(v, "=")
	if !ok {
		return errors.New("--data must be key=value")
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("--data key is empty")
	}

	trimmed := strings.TrimSpace(value)
	if i, err := strconv.Atoi(trimmed); err == nil {
		*p = append(*p, pair{Key: key, Value: i})
		return nil
	}
	if b, err := strconv.ParseBool(trimmed); err == nil {
		*p = append(*p, pair{Key: key, Value: b})
		return nil
	}
	*p = append(*p, pair{Key: key, Value: trimmed})
	return nil
}

type optionalBool struct {
	set   bool
	value bool
}

func (o *optionalBool) Set(v string) error {
	b, err := strconv.ParseBool(v)
	if err != nil {
		return err
	}
	o.set = true
	o.value = b
	return nil
}

func (o *optionalBool) String() string {
	if !o.set {
		return ""
	}
	return strconv.FormatBool(o.value)
}
