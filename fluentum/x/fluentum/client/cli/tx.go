package cli

import (
	"fmt"

	"cosmossdk.io/client"
	"cosmossdk.io/client/tx"
	"github.com/spf13/cobra"

	"github.com/fluentum-chain/fluentum/fluentum/x/fluentum/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdCreateFluentum())
	cmd.AddCommand(CmdUpdateFluentum())
	cmd.AddCommand(CmdDeleteFluentum())

	return cmd
}

func CmdCreateFluentum() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-fluentum [index] [title] [body]",
		Short: "Create a new fluentum",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// Get indexes
			index := args[0]
			title := args[1]
			body := args[2]

			// Get the client context
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Create the message
			msg := types.NewMsgCreateFluentum(
				clientCtx.GetFromAddress().String(),
				index,
				title,
				body,
			)

			// Validate the message
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// Generate and broadcast the transaction
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	// Add transaction flags - using the new approach for Cosmos SDK v0.50.6
	cmd.Flags().String("chain-id", "", "The network chain ID")
	cmd.Flags().String("fees", "", "Fees to pay along with transaction")
	cmd.Flags().String("gas", "auto", "gas limit to set per-block")
	cmd.Flags().String("gas-adjustment", "1.3", "adjustment factor to be multiplied against the estimate returned by the tx simulation")
	cmd.Flags().String("gas-prices", "", "Gas prices in decimal format to determine the transaction fee (e.g. 0.1uatom)")
	cmd.Flags().String("node", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().String("output", "text", "Output format (text|json)")

	return cmd
}

func CmdUpdateFluentum() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-fluentum [index] [title] [body]",
		Short: "Update fluentum",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// Get indexes
			index := args[0]
			title := args[1]
			body := args[2]

			// Get the client context
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Create the message
			msg := types.NewMsgUpdateFluentum(
				clientCtx.GetFromAddress().String(),
				index,
				title,
				body,
			)

			// Validate the message
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// Generate and broadcast the transaction
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	// Add transaction flags
	cmd.Flags().String("chain-id", "", "The network chain ID")
	cmd.Flags().String("fees", "", "Fees to pay along with transaction")
	cmd.Flags().String("gas", "auto", "gas limit to set per-block")
	cmd.Flags().String("gas-adjustment", "1.3", "adjustment factor to be multiplied against the estimate returned by the tx simulation")
	cmd.Flags().String("gas-prices", "", "Gas prices in decimal format to determine the transaction fee (e.g. 0.1uatom)")
	cmd.Flags().String("node", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().String("output", "text", "Output format (text|json)")

	return cmd
}

func CmdDeleteFluentum() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-fluentum [index]",
		Short: "Delete fluentum",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			index := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteFluentum(
				clientCtx.GetFromAddress().String(),
				index,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	// Add transaction flags
	cmd.Flags().String("chain-id", "", "The network chain ID")
	cmd.Flags().String("fees", "", "Fees to pay along with transaction")
	cmd.Flags().String("gas", "auto", "gas limit to set per-block")
	cmd.Flags().String("gas-adjustment", "1.3", "adjustment factor to be multiplied against the estimate returned by the tx simulation")
	cmd.Flags().String("gas-prices", "", "Gas prices in decimal format to determine the transaction fee (e.g. 0.1uatom)")
	cmd.Flags().String("node", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().String("output", "text", "Output format (text|json)")

	return cmd
}
