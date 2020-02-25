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

package main

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

// iniDataContainertype is a container of initial data to run server.
type iniDataContainer struct {
	port     int
	ip       string
	certFile string
	login    string
	password string
}

// myUsage append the standart flag.Usage function with positional arguments.
func myUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] login password\n", os.Args[0])
	flag.PrintDefaults()
}

// init parses cmd line arguments into iniDataContainertype.
func (d *iniDataContainer) init() {
	flag.Usage = myUsage
	flag.IntVar(&d.port, "p", 7777, "port")
	flag.StringVar(&d.ip, "a", "localhost", "server's host address")
	flag.StringVar(&d.certFile, "c", "../cert/server.crt", "tls certificate file")
	flag.Parse()

	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}

	d.login = flag.Arg(0)
	d.password = flag.Arg(1)
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

// connect performs connection to grpc server.
func connect(initData *iniDataContainer) (conn *grpc.ClientConn, err error) {
	// Create the client TLS credentials.
	creds, err := credentials.NewClientTLSFromFile(initData.certFile, "")
	if err != nil {
		return nil, fmt.Errorf("could not load tls cert: %s", err)
	}

	// Setup the login/pass.
	auth := Authentication{
		Login:    initData.login,
		Password: initData.password,
	}

	conn, err = grpc.Dial(fmt.Sprintf("%s:%d", initData.ip, initData.port),
		grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(&auth))

	if err != nil {
		return nil, fmt.Errorf("did not connect: %s", err)
	}
	return conn, err
}

// handleSignals handles signals SIGINT, SIGTERM.
func handleSignals() <-chan interface{} {
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

func gameFlow(connection api.GoGameClient, quit <-chan interface{}) {
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

func main() {
	initData := &iniDataContainer{}
	initData.init()

	fmt.Printf("Hello: %s\n", initData.login)

	conn, err := connect(initData)
	if err != nil {
		log.Fatalf("connection: %s", err)
	}
	defer conn.Close()

	c := api.NewGoGameClient(conn)

	quit := handleSignals()

	gameFlow(c, quit)
}
