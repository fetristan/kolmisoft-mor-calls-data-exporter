package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

// ViaSSHDialer represents a custom SSH Dialer for network connections.
type ViaSSHDialer struct {
	sshClient *ssh.Client
}

// Dial is used to establish an SSH connection to the provided address.
func (v *ViaSSHDialer) Dial(addr string) (net.Conn, error) {
	return v.sshClient.Dial("tcp", addr)
}

// Db represents a database connection.
type Db struct {
	db *sql.DB
}

// newDb creates a new database connection using the given connection string.
func newDb(dbConnectString string) *Db {
	db, err := sql.Open("mysql", dbConnectString)
	if err != nil {
		log.Fatalf("Unable to connect DB: %v", err)
	}

	return &Db{db: db}
}

// prepare compiles a SQL query and returns a prepared statement.
func (db *Db) prepare(query string) *sql.Stmt {
	stmt, err := db.db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}

	return stmt
}

// MorRequest is responsible for making a request to the Mor database via SSH tunnel.
func MorRequest(request string, getConversion func(stmt *sql.Stmt) ([]any, error)) (res []any, err error) {
	// Retrieve database connection details from configuration.
	DbIpMor := viper.GetString("DB_IP_MOR")
	DbPortMor := viper.GetString("DB_PORT_MOR")
	dbNameMor := viper.GetString("DB_NAME_MOR")
	dbUserMor := viper.GetString("DB_USER_MOR")
	dbPassMor := viper.GetString("DB_PASS_MOR")
	dbSshIpMor := viper.GetString("DB_SSH_IP_MOR")
	dbSshPortMor := viper.GetString("DB_SSH_PORT_MOR")
	dbSshUserMor := viper.GetString("DB_SSH_USER_MOR")
	dbSshKeyMor := viper.GetString("DB_SSH_KEY_MOR")
	dbSshKeyPassMor := viper.GetString("DB_SSH_KEY_PASS_MOR")
	dbSshPortMorInt, _ := strconv.Atoi(dbSshPortMor)

	// Read the SSH private key file and create an SSH signer.
	key, err := os.ReadFile(dbSshKeyMor)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKeyWithPassphrase(key, []byte(dbSshKeyPassMor))
	if err != nil {
		return nil, err
	}

	// Configure the SSH client.
	sshConfig := &ssh.ClientConfig{
		User: dbSshUserMor,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Establish an SSH connection.
	sshcon, errSSH := ssh.Dial("tcp", fmt.Sprintf("%s:%d", dbSshIpMor, dbSshPortMorInt), sshConfig)
	if errSSH != nil {
		return nil, err
	}
	defer sshcon.Close()

	// Register a custom MySQL dialer that routes connections through the SSH tunnel.
	mysql.RegisterDialContext("mysql+tcp", func(_ context.Context, addr string) (net.Conn, error) {
		dialer := &ViaSSHDialer{sshcon}
		return dialer.Dial(addr)
	})

	// Create a Data Source Name (DSN) for the MySQL connection.
	dsn := fmt.Sprintf("%s:%s@mysql+tcp(%s)/%s", dbUserMor, dbPassMor, DbIpMor+":"+DbPortMor, dbNameMor)

	// Create a new database connection.
	db := newDb(dsn)
	defer db.db.Close()

	// Prepare the SQL query and execute it.
	query := db.prepare(request)
	defer query.Close()

	// Retrieve and return the results using the provided function.
	return getConversion(query)
}
