package cli

import (
	"fmt"

	"cosmossdk.io/client"
	"github.com/spf13/cobra"

	"github.com/fluentum-chain/fluentum/fluentum/x/fluentum/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group fluentum queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdListFluentum())
	cmd.AddCommand(CmdShowFluentum())

	return cmd
}

func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "shows the parameters of the module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	// Add query flags - using the new approach for Cosmos SDK v0.50.6
	cmd.Flags().String("node", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().String("output", "text", "Output format (text|json)")

	return cmd
}

func CmdListFluentum() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-fluentum",
		Short: "list all fluentum",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllFluentumRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.FluentumAll(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	// Add pagination flags
	cmd.Flags().String("node", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().String("output", "text", "Output format (text|json)")
	cmd.Flags().Uint64("limit", 100, "Query number of fluentum per page")
	cmd.Flags().String("page-key", "", "Query pagination page-key")

	return cmd
}

func CmdShowFluentum() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-fluentum [index]",
		Short: "shows a fluentum",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			argIndex := args[0]

			params := &types.QueryGetFluentumRequest{
				Index: argIndex,
			}

			res, err := queryClient.Fluentum(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	// Add query flags
	cmd.Flags().String("node", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().String("output", "text", "Output format (text|json)")

	return cmd
}
