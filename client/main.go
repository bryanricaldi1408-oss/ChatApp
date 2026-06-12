package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func handleInputFromServer(conn net.Conn, username string) {
	serverReader := bufio.NewReader(conn)
	for {
		message, err := serverReader.ReadString('\n')
		if err != nil {
			fmt.Println("\nTerputus dari server")
			os.Exit(0)
		}
		fmt.Print("\r\033[K")
		fmt.Print(message)
		fmt.Printf("[%s]> ", username)
	}
}

func main() {
	keyboardReader := bufio.NewReader(os.Stdin)

	fmt.Print("Silahkan masukan username anda: ")
	username, err := keyboardReader.ReadString('\n')
	if err != nil {
		fmt.Println("Gagal membaca username")
		os.Exit(1)
	}

	username = strings.TrimSpace(username)

	conn, err := net.Dial("tcp", ":9090")
	if err != nil {
		fmt.Println("Gagal terhubung ke server:", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("berhasil terhubung ke server!")

	fmt.Fprintf(conn, "%s\n", username)

	go handleInputFromServer(conn, username)

	for {
		fmt.Printf("[%s]> ", username)
		input, err := keyboardReader.ReadString('\n')
		if err != nil {
			fmt.Println("Gagal menerima input dari keyboard")
			continue
		}

		if input == "/quit" {
			break
		}
		fmt.Fprintf(conn, "%s", input)
	}
}
