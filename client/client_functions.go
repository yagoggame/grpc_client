// Copyright ©2020 BlinnikovAA. All rights reserved.
// This file is part of yagogame.
//
// yagogame is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// yagogame is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with yagogame.  If not, see <https://www.gnu.org/licenses/>.

package client

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yagoggame/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// IniDataContainer is a container of initial data to run server.
type IniDataContainer struct {
	//Port is the port of grpc service
	Port int
	//IP is the ip address of grpc service
	IP string
	//CertFile is the path to certificate file of grpc service
	CertFile string
	//Login is the user's login
	Login string
	//Password is the user's password
	Password string
}

// myUsage append the standart flag.Usage function with positional arguments.
func myUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] login password\n", os.Args[0])
	flag.PrintDefaults()
}

// Authentication holds the login/password.
type Authentication struct {
	Login    string
	Password string
}

// GetRequestMetadata gets the current request metadata.
func (a *Authentication) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"login":    a.Login,
		"password": a.Password,
	}, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security.
func (a *Authentication) RequireTransportSecurity() bool {
	return true
}

// Connect performs connection to grpc server.
func Connect(initData *IniDataContainer) (conn *grpc.ClientConn, err error) {
	// Create the client TLS credentials.
	creds, err := credentials.NewClientTLSFromFile(initData.CertFile, "")
	if err != nil {
		return nil, fmt.Errorf("could not load tls cert: %s", err)
	}

	// Setup the login/pass.
	auth := Authentication{
		Login:    initData.Login,
		Password: initData.Password,
	}

	conn, err = grpc.Dial(fmt.Sprintf("%s:%d", initData.IP, initData.Port),
		grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(&auth))

	if err != nil {
		return nil, fmt.Errorf("did not connect: %s", err)
	}

	return conn, err
}

// HandleSignals handles signals SIGINT, SIGTERM.
func HandleSignals() <-chan interface{} {
	sigs := make(chan os.Signal, 1)
	done := make(chan interface{})

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func(chan<- interface{}) {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		close(done)
	}(done)

	return done
}

// GameFlow performs main interactive procedure to interact with the game
func GameFlow(connection api.GoGameClient, quit <-chan interface{}) {
	fmt.Printf("Try to enter the Lobby...\n")
	_, err := connection.EnterTheLobby(context.Background(), &api.EmptyMessage{})
	if err != nil {
		st := status.Convert(err)
		log.Fatalf("status error when calling EnterTheLobby: %v: %s", st.Code(), st.Message())
	}

	defer func(c api.GoGameClient) {
		fmt.Printf("Leave the Lobby...\n")
		_, err = c.LeaveTheLobby(context.Background(), &api.EmptyMessage{})
		if err != nil {
			st := status.Convert(err)
			log.Fatalf("status error when calling LeaveTheLobby: %v: %s", st.Code(), st.Message())
		}
	}(connection)

	err = manageGame(connection, quit)
	if err != nil {
		st := status.Convert(err)
		log.Fatalf("status error when calling EnterTheLobby: %v: %s", st.Code(), st.Message())
	}
}
