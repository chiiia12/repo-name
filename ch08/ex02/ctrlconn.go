package main

import (
	"net"
	"log"
	"strings"
	"path/filepath"
	"fmt"
	"strconv"
	"io/ioutil"
	"bytes"
	"bufio"
	"os"
)

type CtrlConnectionManager struct {
	conn     net.Conn
	in       chan string
	out      chan string
	ack      chan struct{}
	done     chan struct{}
	dir      string
	dataConn DataConnectionManager
}

func (cm *CtrlConnectionManager) Init() {
	go func() {
		defer close(cm.done)
		for {
			select {
			case mes := <-cm.out:
				fmt.Fprintf(cm.conn, mes)
				cm.ack <- struct{}{}
			case <-cm.done:
				return
			}
		}
	}()
	go func() {
		defer cm.conn.Close()
		input := bufio.NewScanner(cm.conn)
		for input.Scan() {
			cm.in <- input.Text()
		}
	}()

}

func (cm *CtrlConnectionManager) HandleCommand() {
	log.Println("HandleCommand")
	for {
		msg := <-cm.in
		log.Println(msg)
		command := strings.Split(msg, " ")
		switch(command[0]) {
		case "SYST":
			cm.Send(SystemType, "UNIX Type: L8")
		case "FEAT":
			cm.Send(SystemStatus, "End FEAT.")
		case "PWD":
			cm.Send(PathNameCreated, "\""+cm.dir+"\" is the current directory.")
		case "PORT":
			cm.port(command)
			cm.Send(CommandSucceeded, "port command succeed")
		case "LIST":
			cm.Send(TransferStarting, "transfer start")
			cm.listFiles()
			cm.Send(ClosingConnection, "Transfer complete")
		case "CWD":
			files, _ := ioutil.ReadDir(cm.dir)
			for _, v := range files {
				if v.Name() == command[1] && v.IsDir() {
					cm.dir = filepath.Join(cm.dir, command[1])
					log.Println(cm.dir)
					cm.Send(CommandSucceeded, fmt.Sprintf("%s is the current directory.", cm.dir))
				}
			}
			cm.Send(ActionNotTaken, "No such file or directory.")
		case "SIZE":
			filepath := filepath.Join(cm.dir, command[1])
			file, err := os.OpenFile(filepath, os.O_RDWR, 0666)
			if err != nil {
				log.Println(err)
			}
			fileInfo, _ := file.Stat()
			cm.Send(FileStatus, strconv.Itoa(int(fileInfo.Size())))
			file.Close()
		case "RETR":
			cm.Send(TransferStarting, "transfer start")
			cm.getFileInfo(command[1])
			cm.Send(ClosingConnection, "Transfer complete")
		case "QUIT":
			cm.Send(ServiceClosing, "GoodBye")
			cm.conn.Close()
		default:
			cm.Send(NotImplemented, "command not implemented")
		}
	}
}
func (cm *CtrlConnectionManager) port(command []string) {
	arg := strings.Split(command[1], ",")
	address := strings.Join(arg[0:4], ".")
	a, _ := strconv.Atoi(arg[4])
	b, _ := strconv.Atoi(arg[5])
	port := a*256 + b
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		log.Println(err)
	}
	cm.dataConn.conn = conn
	cm.dataConn.Init()
	log.Println("port")

}

func (cm *CtrlConnectionManager) listFiles() {
	log.Println(cm.dir)
	files, err := ioutil.ReadDir(cm.dir)
	if err != nil {
		log.Println(err)
	}
	for _, v := range files {
		log.Printf("%s\t", v.Name())
		p := fmt.Sprintf("%s\t%s\t%d\t%s\n", v.Mode(), v.ModTime(), v.Size(), v.Name())
		cm.dataConn.out <- bytes.NewBufferString(p)
		<-cm.dataConn.ack
	}
	close(cm.dataConn.done)
	cm.dataConn.conn.Close()
}

func (cm *CtrlConnectionManager) Login() {
	cm.Send(ReadyForUser, "Service ready for new user")
	inputUserName := <-cm.in
	log.Println(inputUserName)
	if inputUserName != UserName {
		cm.Send(SyntaxError, "syntax error")
	}
	cm.Send(NeedPassword, "User name okay, need password.")
	inputPassword := <-cm.in
	log.Println(inputPassword)
	if inputPassword != Password {
		cm.Send(SyntaxError, "syntax error")
	}
	cm.Send(UserLoggedIn, "User logged in, proceed.")
	log.Println("login succeeded")
}

func (cm *CtrlConnectionManager) Send(statusCode int, msg string) {
	cm.out <- fmt.Sprintf("%d %s\n", statusCode, msg)
	<-cm.ack
}
func (cm *CtrlConnectionManager) getFileInfo(filename string) {
	filepath := filepath.Join(cm.dir, filename)
	file, err := os.OpenFile(filepath, os.O_RDWR, 0777)
	if err != nil {
		log.Println(err)
	}
	fileInfo, _ := file.Stat()

	cm.dataConn.out <- bytes.NewBufferString(strconv.Itoa(int(fileInfo.Size())))
	<-cm.dataConn.ack

	file.Close()
	close(cm.dataConn.done)
	cm.dataConn.conn.Close()
}
