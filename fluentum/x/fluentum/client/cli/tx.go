package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
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

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdUpdateFluentum() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-fluentum [index] [title] [body]",
		Short: "Update a fluentum",
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

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdDeleteFluentum() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-fluentum [index]",
		Short: "Delete a fluentum",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			index := args[0]

			// Get the client context
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Create the message
			msg := types.NewMsgDeleteFluentum(
				clientCtx.GetFromAddress().String(),
				index,
			)

			// Validate the message
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// Generate and broadcast the transaction
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
