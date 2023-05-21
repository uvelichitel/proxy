package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
)

const localAddr string = ":4430"
const remoteAddr string = "suckless.org:443"
const MaxConn int16 = 3

func Pipe(conn1, conn2 net.Conn) {
	_, err := io.Copy(conn1, conn2)
	if err != nil {
		conn1.Close()
		conn2.Close()
	}
}

func proxyConn(conn net.Conn) {

	rAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		fmt.Println(err)
	}

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	rConn, err := tls.Dial("tcp", rAddr.String(), conf)
	if err != nil {
		fmt.Println(err)
	}

	go Pipe(conn, rConn)
	go Pipe(rConn, conn)
	fmt.Printf("handleConnection end: %s\n", conn.RemoteAddr())
}

func main() {
	caCert, err := ioutil.ReadFile("client.pem")
	if err != nil {
		fmt.Println(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cer, err := tls.LoadX509KeyPair("server.pem", "server.key")
	if err != nil {
		fmt.Println(err)
		return
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cer},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}
	config.BuildNameToCertificate()

	ln, err := tls.Listen("tcp", localAddr, config)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ln.Close()

	fmt.Printf("Listening: %v -> %v\n\n", localAddr, remoteAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go proxyConn(conn)
	}

}
