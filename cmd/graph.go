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
package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/mlieberman85/scq/pkg/graph"
	"github.com/mlieberman85/scq/pkg/record"
	"github.com/spf13/cobra"
)

// graphCmd represents the graph command
var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Generates a supply chain graph based on attestations and metadata",
	Long:  `TODO.`,
	Run: func(cmd *cobra.Command, args []string) {
		/*c, err := record.GetMongoClient("mongodb://localhost", "supplychain", "attestations")
		if err != nil {
			log.Fatal(err)
		}*/
		c, err := record.GetRekorClient()
		if err != nil {
			log.Fatal(err)
		}

		// TODO: Make this configurable
		cs := []record.RecordClient{c}

		scg := graph.SupplyChainGraph{
			Nodes: make(map[string]*record.Record),
			Edges: make(map[string][]string),
			RecordManager: &record.Manager{
				Opts: record.ManagerOpts{
					IsTest: true,
				},
				Clients: cs,
			},
		}

		hash, err := cmd.Flags().GetString("hash")
		if err != nil {
			log.Fatal(err)
		}

		err = scg.GenerateFromHash(hash)
		if err != nil {
			log.Fatal(err)
		}

		j, err := json.Marshal(scg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(j))

	},
}

func init() {
	rootCmd.AddCommand(graphCmd)
	graphCmd.Flags().String("hash", "g", "Hash of the artifact you want generate a graph for")
}
