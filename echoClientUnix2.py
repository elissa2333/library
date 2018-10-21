#! /usr/bin/env python3
# -*- coding: utf-8 -*-

import socket


sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
sock.connect("test.sock")

while True:
    data = input("请输入您需要发送的数据:")
    data = data + "\n" #\n 是给 go 服务端读的
    sock.send(data.encode("utf-8"))
    receive = sock.recv(1024) #多读一些数据，不影响的
    print("py : ",receive.decode("utf-8"))

sock.close()
