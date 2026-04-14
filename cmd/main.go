package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/awafinance/fiscal"
	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "fiscal <xml>",
		Short: "Parse Brazilian fiscal documents (NF-e, NFS-e, CT-e, MDF-e, BP-e)",
		Long: `
▐▘▘      ▜
▜▘▌▛▘▛▘▀▌▐
▐ ▌▄▌▙▖█▌▐▖

A Fast and Flexible fiscal document parser and converter.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read fiscal document: %w", err)
			}

			doc, err := fiscal.Parse(data)
			if err != nil {
				return err
			}

			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				if err := enc.Encode(doc); err != nil {
					return fmt.Errorf("encode json: %w", err)
				}
				return nil
			}

			return printSummary(cmd.OutOrStdout(), doc)
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output the full typed document as JSON")
	return cmd
}

func printSummary(w io.Writer, doc *fiscal.Document) error {
	info := doc.Info()
	if info == nil {
		return errors.New("document has no info")
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	row := func(label, value string) {
		if value == "" {
			return
		}
		fmt.Fprintf(tw, "%s\t%s\n", label, value)
	}

	family := strings.ToUpper(string(doc.Family))
	if doc.RootName != "" {
		family = fmt.Sprintf("%s (%s)", family, doc.RootName)
	}
	row("Family:", family)
	row("Access key:", info.GetAccessKey())
	row("Version:", joinNonEmpty("  Env: ", info.GetVersion(), envLabel(info.GetEnvironment())))
	row("Number:", joinNonEmpty("  Series: ", info.GetNumber(), info.GetSeries()))
	row("Model:", info.GetModel())
	row("Issued:", info.GetIssueDate())
	row("Issuer:", partyLine(info.GetIssuer(), info.GetIssuerDocument()))
	row("Recipient:", partyLine(info.GetRecipient(), info.GetRecipientDocument()))
	row("Amount:", info.GetAmount())
	row("Status:", statusLine(info))
	row("Protocol:", info.GetProtocolNumber())

	if err := tw.Flush(); err != nil {
		return err
	}

	printAmounts(w, info)
	printParties(w, info)
	printRoute(w, info)
	printRelated(w, info)
	return nil
}

func envLabel(env string) string {
	switch env {
	case "1":
		return "production"
	case "2":
		return "homologation"
	default:
		return env
	}
}

func joinNonEmpty(sep string, first, second string) string {
	if first == "" {
		return second
	}
	if second == "" {
		return first
	}
	return first + sep + second
}

func partyLine(name, document string) string {
	switch {
	case name == "" && document == "":
		return ""
	case document == "":
		return name
	case name == "":
		return document
	default:
		return fmt.Sprintf("%s (%s)", name, document)
	}
}

func statusLine(info fiscal.DocumentInfo) string {
	code, reason := info.GetStatusCode(), info.GetStatusReason()
	line := strings.TrimSpace(code + " " + reason)
	if info.IsAuthorized() {
		if line == "" {
			return "authorized"
		}
		return line + "  [authorized]"
	}
	return line
}

func printAmounts(w io.Writer, info fiscal.DocumentInfo) {
	ai, ok := info.(fiscal.AmountsInfo)
	if !ok {
		return
	}
	amounts := ai.GetAmounts()
	if len(amounts) == 0 {
		return
	}
	fmt.Fprintln(w, "\nAmounts:")
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, a := range amounts {
		fmt.Fprintf(tw, "  %s\t%s\n", a.Type, a.Value)
	}
	tw.Flush()
}

func printParties(w io.Writer, info fiscal.DocumentInfo) {
	pi, ok := info.(fiscal.PartiesInfo)
	if !ok {
		return
	}
	parties := pi.GetParties()
	if len(parties) == 0 {
		return
	}
	fmt.Fprintln(w, "\nParties:")
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, p := range parties {
		fmt.Fprintf(tw, "  %s\t%s\n", p.Role, partyLine(p.Name, p.Document))
	}
	tw.Flush()
}

func printRoute(w io.Writer, info fiscal.DocumentInfo) {
	ri, ok := info.(fiscal.RouteInfo)
	if !ok {
		return
	}
	modal := ri.GetModal()
	origin, destination := ri.GetOrigin(), ri.GetDestination()
	if modal == "" && isZeroLocation(origin) && isZeroLocation(destination) {
		return
	}
	fmt.Fprintln(w, "\nRoute:")
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if modal != "" {
		fmt.Fprintf(tw, "  Modal:\t%s\n", modal)
	}
	if !isZeroLocation(origin) {
		fmt.Fprintf(tw, "  Origin:\t%s\n", locationLine(origin))
	}
	if !isZeroLocation(destination) {
		fmt.Fprintf(tw, "  Destination:\t%s\n", locationLine(destination))
	}
	tw.Flush()
}

func printRelated(w io.Writer, info fiscal.DocumentInfo) {
	ri, ok := info.(fiscal.RelatedDocumentsInfo)
	if !ok {
		return
	}
	docs := ri.GetRelatedDocuments()
	if len(docs) == 0 {
		return
	}
	fmt.Fprintln(w, "\nRelated documents:")
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, d := range docs {
		ref := d.AccessKey
		if ref == "" {
			ref = strings.TrimSpace(d.Number + " série " + d.Series)
		}
		fmt.Fprintf(tw, "  %s\t%s\n", d.Type, ref)
	}
	tw.Flush()
}

func isZeroLocation(l fiscal.Location) bool {
	return l == fiscal.Location{}
}

func locationLine(l fiscal.Location) string {
	parts := []string{}
	if l.CityName != "" {
		parts = append(parts, l.CityName)
	}
	if l.State != "" {
		parts = append(parts, l.State)
	}
	if l.CountryCode != "" {
		parts = append(parts, l.CountryCode)
	}
	line := strings.Join(parts, "/")
	if l.CityCode != "" {
		if line == "" {
			return l.CityCode
		}
		return fmt.Sprintf("%s (%s)", line, l.CityCode)
	}
	return line
}

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
