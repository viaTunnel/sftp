package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/viaTunnel/sftp"
	"golang.org/x/crypto/ssh"
)

func main() {

	if err := godotenv.Load(); err != nil {
		fmt.Printf("Error loading the .env file: %v", err)
	}

	sftpHost := os.Getenv("HOSTNAME")
	port := os.Getenv("PORT")

	client := AuthWithPublicKey(os.Getenv("USERNAME"), sftpHost, port)
	// client := AuthWithPassword(os.Getenv("USERNAME"), os.Getenv("PASSWORD"), sftpHost, port)

	defer client.Close()

	remoteFilePath := "Tunnel/Wells_fargo/Outbox"
	remoteFileName := "test.txt"

	UploadFileFromString(client, remoteFilePath, remoteFileName, "Testing Testing 1... 2... 3... Check Check")

	/*
		UseConcurrentWrites allows the Client to perform concurrent Writes.
		Using concurrency while doing writes, requires special consideration. A write to a later offset in a file after an error,
		could end up with a file length longer than what was successfully written.
		When using this option, if you receive an error during `io.Copy` or `io.WriteTo`, you may need to `Truncate` the target
		Writer to avoid “holes” in the data written.
	*/
	// sftp.UseConcurrentWrites(true)

}

func UploadFileFromString(client *sftp.Client, remotePath string, remoteFileName string, content string) {
	localReader := strings.NewReader(content)

	remoteFile, err := client.OpenFile(remotePath+"/"+remoteFileName, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		log.Fatal("Failed to create remote file: ", err)
	}
	defer remoteFile.Close()

	remoteFile.ReadFrom(localReader)
	if err != nil {
		log.Fatal("Failed to write to remote file: ", err)
	}
}

func AuthWithPassword(username string, password string, hostname string, port string) *sftp.Client {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		// TODO: Update host key callback before prod release
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", hostname+":"+port, config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatal("Failed to create SFTP client: ", err)
	}

	return client
}

func AuthWithPublicKey(username string, hostname string, port string) *sftp.Client {
	// Read key file
	key, err := os.ReadFile("sftp.pem")
	if err != nil {
		log.Fatal("Failed to read private key: ", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}

	// Open SFTP connection
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		// TODO: Update host key callback before prod release
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	config.SetDefaults()

	conn, err := ssh.Dial("tcp", hostname+":"+port, config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		log.Fatal("Failed to create SFTP client: ", err)
	}

	return client
}
