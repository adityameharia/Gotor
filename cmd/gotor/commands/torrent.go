/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package commands

import (
	file "github.com/adityameharia/gotor/file"
	"log"

	"github.com/spf13/cobra"
)

// torrentCmd represents the torrent command
var torrentCmd = &cobra.Command{
	Use:   "torrent",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		download(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(torrentCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// torrentCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// torrentCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
func download(path string, dest string) {
	f, err := file.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	err = f.DownloadFile(dest)
	if err != nil {
		log.Fatal(err)
	}
}
