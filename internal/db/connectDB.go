package db

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type DatabaseCreds struct {
	SSHHost    string // SSH Server Hostname/IP
	SSHPort    int    // SSH Port
	SSHUser    string // SSH Username
	SSHKeyFile string // SSH Key file location
	DBUser     string // DB username
	DBPass     string // DB Password
	DBHost     string // DB Hostname/IP
	DBName     string // Database name
}

type ViaSSHDialer struct {
	client *ssh.Client
}

func (dialer *ViaSSHDialer) Dial(addr string) (net.Conn, error) {
	return dialer.client.Dial("tcp", addr)
}

func ConnectToDB(dbCreds DatabaseCreds) (*sql.DB, *ssh.Client, error) {

	// Make SSH client: establish a connection to the local ssh-agent
	var agentClient agent.Agent
	if conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		defer conn.Close()
		agentClient = agent.NewClient(conn)
	}

	pemBytes, err := os.ReadFile(dbCreds.SSHKeyFile)
	if err != nil {
		return nil, nil, err
	}
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		return nil, nil, err
	}

	// The client configuration with configuration option to use the ssh-agent
	sshConfig := &ssh.ClientConfig{
		User:            dbCreds.SSHUser,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// When the agentClient connection succeeded, add them as AuthMethod
	if agentClient != nil {
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeysCallback(agentClient.Signers))
	}

	// Connect to the SSH Server
	sshConn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", dbCreds.SSHHost, dbCreds.SSHPort), sshConfig)
	if err != nil {
		return nil, nil, err
	}

	// Now we register the ViaSSHDialer with the ssh connection as a parameter
	mysql.RegisterDialContext("mysql+tcp", func(_ context.Context, addr string) (net.Conn, error) {
		dialer := &ViaSSHDialer{sshConn}
		return dialer.Dial(addr)
	})

	// And now we can use our new driver with the regular mysql connection string tunneled through the SSH connection
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@mysql+tcp(%s)/%s", dbCreds.DBUser, dbCreds.DBPass, dbCreds.DBHost, dbCreds.DBName))
	if err != nil {
		return nil, sshConn, err
	}

	return db, sshConn, err
}
