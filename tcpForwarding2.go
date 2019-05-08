/*
go build tcpForwarding2.go

./tcpForwarding2 -local 1080 -remote 8.8.8.8:2333
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	loc int
	rem string
)

func init() {
	flag.IntVar(&loc, "local", -1, "本地端口号")
	flag.StringVar(&rem, "remote", "", "远程ip地址和端口号")
}

func main() {
	flag.Parse()
	local, remote := checkInput(&loc, &rem)
	forword(local, remote)
}

func checkInput(local *int, remote *string) (locals, remotes string) {
	if *local == -1 && *remote == "" {
		fmt.Println("请输入本地端口,和远程端口")
		flag.PrintDefaults()
		os.Exit(0)
	}
	if *local == -1 {
		fmt.Println("请输入本地端口")
		os.Exit(0)
	} else {
		if *local > 0 && *local < 65535 {
			locals = strconv.Itoa(*local)
		}
	}

	if *remote == "" {
		fmt.Println("请输入远程ip地址和端口")
		os.Exit(0)
	} else {
		items := strings.Split(*remote, ":")
		if items == nil || len(items) != 2 {
			fmt.Println("输入错误请重新输入,格式类似于 -> 8.8.8.8:53")
			os.Exit(0)
		}
		ip := net.ParseIP(items[0])
		if ip == nil {
			fmt.Println("远程ip地址格式错误")
			os.Exit(0)
		}
		remotePort, err := strconv.Atoi(items[1])
		if err != nil {
			fmt.Println("远程地址端口错误")
			os.Exit(0)
		}
		if remotePort > 0 && remotePort < 65535 {
			remotes = ip.String() + ":" + strconv.Itoa(remotePort)
		}
	}
	return locals, remotes

}

func forword(localPort, remoteAddress string) {
	lis, err := net.Listen("tcp", ":"+localPort)
	if err != nil {
		log.Fatal("端口监听失败 -> ", err)
	}
	defer lis.Close()

	remoteConn, err := net.DialTimeout("tcp", remoteAddress, 2*time.Second)
	if err != nil {
		lis.Close()
		log.Fatal("远程链接建立失败", err)
	}
	remoteConn.Close()

	for {
		localConn, err := lis.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handle(localConn, remoteAddress)
	}
}

func handle(localConn net.Conn, remoteAddress string) {
	var wg sync.WaitGroup

	remoteConn, err := net.DialTimeout("tcp", remoteAddress, 2*time.Second)
	if err != nil {
		log.Fatal("远程链接建立失败", err)
	}

	wg.Add(2)
	go func(localOut net.Conn, remoteIn net.Conn) {
		defer wg.Done()
		io.Copy(remoteIn, localOut)
		remoteIn.Close()
	}(localConn, remoteConn)
	go func(localIn net.Conn, remoteOut net.Conn) {
		defer wg.Done()
		io.Copy(localIn, remoteOut)
		localIn.Close()
	}(localConn, remoteConn)
	wg.Wait()
}
