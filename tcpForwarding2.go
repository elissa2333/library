package main

//go:generate go run main.go -local 1080 -remote 8.8.8.8:2333

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

var (
	localAddress  string
	remoteAddress string
	isCheck       bool

	timeout = 2 * time.Second
)

func init() {
	flag.StringVar(&localAddress, "local", "", "本地端口或地址")
	flag.StringVar(&remoteAddress, "remote", "", "远程地址")
	flag.BoolVar(&isCheck, "check", false, "启动时是否进行远程连接建立测试")
}

func main() {
	flag.Parse()

	if localAddress == "" && remoteAddress == "" {
		fmt.Println("请输入本地端口,和远程端口")
		flag.PrintDefaults()
		return
	}
	if localAddress == "" {
		fmt.Println("请输入本地端口")
		return
	}

	localhost, localPort, err := net.SplitHostPort(localAddress)
	if err != nil {
		_, errConv := strconv.Atoi(localAddress)
		if errConv != nil {
			fmt.Println("远程地址错误请重新输入，格式类似于 -> 8000")
			return
		}
		localPort = localAddress
	}
	if localhost == "" {
		localhost = "127.0.0.1"
	}

	localAddress = net.JoinHostPort(localhost, localPort)

	if remoteAddress == "" {
		fmt.Println("请输入远程ip地址和端口")
		return
	}
	remoteHOST, remotePort, err := net.SplitHostPort(remoteAddress)
	if remoteHOST == "" || remotePort == "" || err != nil {
		fmt.Println("远程地址错误请重新输入，格式类似于 -> 8.8.8.8:53")
		return
	}

	lis, err := net.Listen("tcp", localAddress)
	if err != nil {
		fmt.Printf("本地监听失败: %s\n", err)
		return
	}
	defer lis.Close()

	if isCheck {
		remoteConn, err := net.DialTimeout("tcp", remoteAddress, timeout)
		if err != nil {
			fmt.Printf("远程链接建立失败: %s\n", err)
			return
		}
		_ = remoteConn.Close()
	}

	fmt.Printf("转发服务启动\n\n本机地址 %s 为远程 %s 的映射\n", localAddress, remoteAddress)

	for {
		localConn, err := lis.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				fmt.Println("本机监听服务关闭")
				return
			}

			log.Println(err)
			continue
		}
		go func(conn net.Conn, remoteAddress string) {
			if err := handle(conn, remoteAddress); err != nil {
				log.Println(err)
			}

			_ = localConn.Close()
		}(localConn, remoteAddress)
	}
}

func handle(localConn net.Conn, remoteAddress string) error {
	remoteConn, err := net.DialTimeout("tcp", remoteAddress, timeout)
	if err != nil {
		return err
	}
	defer remoteConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func(localOut net.Conn, remoteIn net.Conn) {
		defer wg.Done()
		_, _ = io.Copy(remoteIn, localOut)
		_ = remoteIn.Close()
	}(localConn, remoteConn)
	go func(localIn net.Conn, remoteOut net.Conn) {
		defer wg.Done()
		_, _ = io.Copy(localIn, remoteOut)
		_ = localIn.Close()
	}(localConn, remoteConn)
	wg.Wait()

	return nil
}
