package cmd

import (
	"asocial/wire"
	"fmt"

	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "chat server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("chat called")
		server, err := wire.InitializeChatServer("chat")
		if err != nil {
			fmt.Println(err)
			return
		}
		server.Serve()
	},
}

// server -> Router (+ name, infraCloser, obsInjector ) -> HttpServer (+ grcpServer) -> gin server + websocket



func init() {
	rootCmd.AddCommand(chatCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// chatCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// chatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
