package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fluentum-chain/fluentum/p2p"
)

// ShowNodeIDCmd dumps node's ID to the standard output.
var ShowNodeIDCmd = &cobra.Command{
	Use:     "show-node-id",
	Aliases: []string{"show_node_id"},
	Short:   "Show this node's ID",
	RunE:    showNodeID,
	PreRun:  deprecateSnakeCase,
}

func showNodeID(cmd *cobra.Command, args []string) error {
	nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
	if err != nil {
		return err
	}

	fmt.Println(nodeKey.ID())
	return nil
}
