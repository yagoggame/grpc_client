/*
Copyright © 2020 Blinnikov AA <goofinator@mail.ru>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/yagoggame/api"
	"github.com/yagoggame/grpc_client/client"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	initData client.IniDataContainer
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "grpc_client",
		Short: "grpc_client is a grpc client of yagogame.",
		Long: `grpc_client is a part of yagogame.
	yagogame is yet another Go game on the Go, made just for fun.
	grpc_client provides access to the go game grpc_server thru grpc with CLI interface`,

		// Uncomment the following line if your bare application
		// has an action associated with it:
		Run: mainCmdFnc,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.grpc_client.yaml)")
	rootCmd.PersistentFlags().StringVarP(&initData.Login, "login", "l", "", "login to use")
	rootCmd.MarkPersistentFlagRequired("login")
	rootCmd.PersistentFlags().StringVarP(&initData.Password, "password", "p", "", "password to use")
	rootCmd.MarkPersistentFlagRequired("password")
	rootCmd.PersistentFlags().StringVarP(&initData.IP, "address", "A", "localhost", "ip address of grpc_server")
	rootCmd.PersistentFlags().IntVarP(&initData.Port, "port", "P", 7777, "port of grpc_server")
	rootCmd.PersistentFlags().StringVarP(&initData.CertFile, "cert", "C", "", "ip address of grpc_server")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".grpc_client" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".grpc_client")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func mainCmdFnc(cmd *cobra.Command, args []string) {
	fmt.Printf("Hello: %s\n", initData.Login)

	conn, err := client.Connect(&initData)
	if err != nil {
		log.Fatalf("connection: %s", err)
	}
	defer conn.Close()

	c := api.NewGoGameClient(conn)

	quit := client.HandleSignals()

	client.GameFlow(c, quit)
}
