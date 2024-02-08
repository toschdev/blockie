package cmd

import "github.com/ignite/cli/v28/ignite/services/plugin"

// GetCommands returns the list of blockie app commands.
func GetCommands() []*plugin.Command {
	return []*plugin.Command{
		{
			Use:   "blockie start",
			Short: "blockie is an Ignite App for developers to understand blocks deeper!",
			Commands: []*plugin.Command{
				{
					Use:   "start",
					Short: "Starts getting latest data from the blockchain",
				},
			},
		},
	}
}
