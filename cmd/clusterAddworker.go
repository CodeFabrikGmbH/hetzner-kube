// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
	"errors"
	"fmt"
	"log"
	"strings"
	"strconv"
	"github.com/xetys/hetzner-kube/pkg"
)

// clusterAddworkerCmd represents the clusterAddworker command
var clusterAddworkerCmd = &cobra.Command{
	Use:   "addworker",
	Short: "Add worker nodes",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return nil
		}

		if name == "" {
			return errors.New("flag --name is required")
		}

		idx, _ := AppConf.Config.FindClusterByName(name)

		if idx == -1 {
			return errors.New(fmt.Sprintf("cluster '%s' not found", name))
		}

		var workerServerType string
		if workerServerType, _ = cmd.Flags().GetString("worker-server-type"); workerServerType == "" {
			return errors.New("flag --worker_server_type is required")
		}

		if err != nil {
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		nodeCount, _ := cmd.Flags().GetInt("nodes")
		name, _ := cmd.Flags().GetString("name")
		_, cluster := AppConf.Config.FindClusterByName(name)
		workerServerType, _ := cmd.Flags().GetString("worker-server-type")
		var sshKeyName string

		for _, node := range cluster.Nodes {
			if node.IsMaster {
				sshKeyName = node.SSHKeyName
			}
		}

		if sshKeyName == "" {
			log.Fatal("master not found")
		}

		maxNo := 0
		for _, node := range cluster.Nodes {
			if !node.IsMaster {
				nameParts := strings.Split(node.Name, "-")
				no, _ := strconv.Atoi(nameParts[len(nameParts)-1])

				if no > maxNo {
					maxNo = no
				}
			}
		}

		cluster.coordinator = pkg.NewProgressCoordinator()

		nodes, err := cluster.CreateWorkerNodes(sshKeyName, workerServerType, nodeCount, maxNo)

		if err != nil {
			log.Fatal(err)
		}

		saveCluster(cluster)

		cluster.ProvisionNodes(nodes)
		saveCluster(cluster)

		cluster.InstallWorkers(nodes)
		saveCluster(cluster)
	},
}

func init() {
	clusterCmd.AddCommand(clusterAddworkerCmd)
	clusterAddworkerCmd.Flags().StringP("name", "", "", "Name of the cluster to delete")
	clusterAddworkerCmd.Flags().String("worker-server-type", "cx11", "Server type used of workers")
	clusterAddworkerCmd.Flags().IntP("nodes", "n", 2, "Number of nodes for the cluster")
}
