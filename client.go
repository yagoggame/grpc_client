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
func connect(login, password string) (conn *grpc.ClientConn, err error) {
	// Create the client TLS credentials.
	creds, err := credentials.NewClientTLSFromFile("../cert/server.crt", "")
	if err != nil {
		return nil, fmt.Errorf("could not load tls cert: %s", err)
	}

	// Setup the login/pass.
	auth := Authentication{
		Login:    login,
		Password: password,
	}

	conn, err = grpc.Dial("localhost:7777",
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

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s login password", os.Args[0])
	}
	login := os.Args[1]
	password := os.Args[2]

	fmt.Printf("Hello: %s\n", login)

	conn, err := connect(login, password)
	if err != nil {
		log.Fatalf("connection: %s", err)
	}
	defer conn.Close()

	c := api.NewGoGameClient(conn)

	quit := handleSignals()

	// Enter a lobby of gamers.
	fmt.Printf("Try to enter the Lobby...\n")
	_, err = c.EnterTheLobby(context.Background(), &api.EmptyMessage{})
	if err != nil {
		st := status.Convert(err)
		log.Fatalf("status error when calling EnterTheLobby: %v: %s", st.Code(), st.Message())
	}

	// After all - Leave a lobby of gamers.
	defer func(c api.GoGameClient) {
		fmt.Printf("Leave the Lobby...\n")
		_, err = c.LeaveTheLobby(context.Background(), &api.EmptyMessage{})
		if err != nil {
			st := status.Convert(err)
			log.Fatalf("status error when calling LeaveTheLobby: %v: %s", st.Code(), st.Message())
		}
	}(c)

	err = manageGame(c, quit)
	if err != nil {
		st := status.Convert(err)
		log.Fatalf("status error when calling EnterTheLobby: %v: %s", st.Code(), st.Message())
	}
}
