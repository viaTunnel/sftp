package main

import (
	"fmt"
	"log"
	"os"

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

	remoteFilePath := "Tunnel/WellsFargo/Outbox"

	err := TestDirPresent(client, "/tmp/foo/bar")
	if err != nil {
		fmt.Println("Directory not present: ", "/tmp/foo/bar")
	} else {
		fmt.Println("Directory present: ", "/tmp/foo/bar")
	}

	errr := TestDirPresent(client, remoteFilePath)
	if errr != nil {
		fmt.Println("Directory not present: ", remoteFilePath)
	} else {
		fmt.Println("Directory present: ", remoteFilePath)
	}

	/*
		UseConcurrentWrites allows the Client to perform concurrent Writes.
		Using concurrency while doing writes, requires special consideration. A write to a later offset in a file after an error,
		could end up with a file length longer than what was successfully written.
		When using this option, if you receive an error during `io.Copy` or `io.WriteTo`, you may need to `Truncate` the target
		Writer to avoid “holes” in the data written.
	*/
	// sftp.UseConcurrentWrites(true)

}

func TestDirPresent(client *sftp.Client, dir string) (err error) {
	var fi os.FileInfo
	fi, err = client.Stat(dir)
	if err == nil {
		if fi.IsDir() {
			return nil
		}
	}
	return err
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
