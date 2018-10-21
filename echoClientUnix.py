#! /usr/bin/env python3
# -*- coding: utf-8 -*-

import socket


sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) #发起链接
sock.connect("test.sock") #指定进程文件

while True:
    data = input("请输入您需要发送的数据:") #读取用户输入
    sock.send(data.encode("utf-8")) #发送数据。因为只支持 byte 所以需要先编码成 utf-8 
    receive_data = sock.recv(len(data)) #接收数据，因为 python3 的接收需要指定长度，所以这里发多少就接多少
    print(receive_data) #打印接收的数据

sock.close()
