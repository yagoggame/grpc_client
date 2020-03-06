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
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yagoggame/api"
	"github.com/yagoggame/grpc_client/client"
)

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "register this user on the service",
	Long:  `register this user on the service's data base`,
	Run:   registerCmdFnc,
}

func init() {
	rootCmd.AddCommand(registerCmd)
}

func registerCmdFnc(cmd *cobra.Command, args []string) {
	fmt.Printf("do you realy want to register %q user on service?\ntype \"yes\" if you do.\n", initData.Login)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		log.Fatal("Canceled")
	}
	if txt := scanner.Text(); strings.Compare(txt, "yes") != 0 {
		log.Fatal("Canceled")
	}

	conn, err := client.Connect(&initData)
	if err != nil {
		log.Fatalf("connection: %s", err)
	}
	defer conn.Close()

	c := api.NewGoGameClient(conn)

	_, err = c.RegisterUser(context.Background(), &api.EmptyMessage{})
	if err != nil {
		log.Fatalf("RegisterUser: %s", err)
	}
	log.Print("Done")
}
