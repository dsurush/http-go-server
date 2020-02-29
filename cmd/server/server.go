package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const addr = "0.0.0.0:1111"

func main() {
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)
	defer func() {
		err := logFile.Close()
		if err != nil {
			log.Fatalf("can't close logfile: %v\n",err)
		}
	}()

	err = start(addr)
	if err != nil {
		log.Fatalf("can't start func is error: %v\n", err)
	}
}

func start(addr string) (err error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("can't listen %s: %w", addr, err)
	}
	defer func() {
		err := listener.Close()
		if err != nil {
			log.Fatalf("can't listener close: %v\n", err)
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("can't listner accept: %v\n", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatalf("can't close connection: %v\n", err)
		}
	}()

	reader := bufio.NewReader(conn)
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	log.Println(requestLine)
	parts := strings.Split(strings.TrimSpace(requestLine), " ")
	if err != nil {
		log.Printf("can't slip trimSpace: %v\n", err)
		return
	}
	log.Println(parts)

	method, request, protocol := parts[0], parts[1], parts[2]
	if method != "GET" {
		panic("")
	}
	if method == "GET" {
		if protocol == "HTTP/1.1" {
			if len(request) > (len("?download") + 3) {
				if a := request[len(request)-9:]; a == "?download" {
					handleHttp(conn, request[1:len(request)-9], contentTypeDownload)
				}
			} else {
				if request == "/" {
					handleHttp(conn, "index.html", contentTypeHtml)
					return
				}
				requestHttp := request[1:]
				ext:=make(map[string]string)
				ext[".txt"]=contentTypeText
				ext[".pdf"]=contentTypePdf
				ext[".png"]=contentTypePng
				ext[".jpg"]=contentTypeJpg
				ext[".html"]=contentTypeHtml
				contentType, ok :=ext[filepath.Ext(requestHttp)]
				if !ok {
					handleHttp(conn, "404.html", contentTypeHtml)
				}
				handleHttp(conn, requestHttp, contentType)
			}
		}
	}
}

func handleHttp(conn net.Conn, fileName string, contentType string) {
	var fileDir = serverFileDir + fileName
	writer := bufio.NewWriter(conn)
	bytesFile, err := ioutil.ReadFile(fileDir)
	if err != nil {
		log.Printf("can't read file: %v\n", err)
		return
	}
	_, err = writer.WriteString("HTTP/1.1 200 OK\r\n")
	if err != nil {
		log.Printf("can't write: %v\n", err)
		return
	}
	_, err = writer.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(bytesFile)))
	if err != nil {
		log.Printf("can't write: %v\n", err)
		return
	}
	if contentType == contentTypeDownload {
		_, err = writer.WriteString(fmt.Sprintf(contentType + fileName + "\r\n"))
		if err != nil {
			log.Printf("can't write: %v\n", err)
			return
		}
	} else {
		_, err = writer.WriteString(fmt.Sprintf(contentType))
		if err != nil {
			log.Printf("can't write: %v\n", err)
			return
		}
	}

	_, err = writer.WriteString("Connection: Close\r\n")
	if err != nil {
		log.Printf("can't write: %v\n", err)
		return
	}
	_, err = writer.WriteString("\r\n")
	if err != nil {
		log.Printf("can't write: %v\n", err)
		return
	}
	_, err = writer.Write(bytesFile)
	if err != nil {
		log.Printf("can't write: %v\n", err)
		return
	}
	err = writer.Flush()
	if err != nil {
		log.Printf("can't flush: %v\n", err)
	}
	return
}
