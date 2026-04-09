package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/awa/nota-fiscal/pkg/bpe"
	"github.com/awa/nota-fiscal/pkg/cte"
	"github.com/awa/nota-fiscal/pkg/mdfe"
	"github.com/awa/nota-fiscal/pkg/nfe"
	"github.com/awa/nota-fiscal/pkg/nfse"
	"github.com/spf13/cobra"
)

type parser func([]byte) (any, error)

var rootCmd = &cobra.Command{
	Use:   "fiscal",
	Short: "Fiscal helps you with brazilian fiscal documents",
	Long: `
▐▘▘      ▜ 
▜▘▌▛▘▛▘▀▌▐ 
▐ ▌▄▌▙▖█▌▐▖

A Fast and Flexible fiscal document parser and converter.`,
}

func init() {
	rootCmd.AddCommand(
		newParseCommand("nfe", "Parse an NF-e XML document", func(data []byte) (any, error) {
			return nfe.Parse(data)
		}),
		newParseCommand("nfse", "Parse an NFS-e XML document", func(data []byte) (any, error) {
			return nfse.Parse(data)
		}),
		newParseCommand("cte", "Parse a CT-e XML document", func(data []byte) (any, error) {
			return cte.Parse(data)
		}),
		newParseCommand("mdfe", "Parse an MDF-e XML document", func(data []byte) (any, error) {
			return mdfe.Parse(data)
		}),
		newParseCommand("bpe", "Parse a BP-e XML document", func(data []byte) (any, error) {
			return bpe.Parse(data)
		}),
	)
}

func newParseCommand(use, short string, parse parser) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <xml>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read fiscal document: %w", err)
			}

			doc, err := parse(data)
			if err != nil {
				return err
			}

			if err := json.NewEncoder(cmd.OutOrStdout()).Encode(doc); err != nil {
				return fmt.Errorf("encode json: %w", err)
			}
			return nil
		},
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
