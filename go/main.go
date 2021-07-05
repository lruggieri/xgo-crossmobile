package main


/*
#include <string.h> // for strcpy
#include <stdlib.h> // for 'free' function
#include <stdbool.h>

typedef void (*loggerFunc) (char* message);
void bridge_logger(loggerFunc, char* message);
*/
import "C"
import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
	"unsafe"
)

var externalLogger func(format string)

func main() {
	mode := flag.String("mode", "", "'server' / 'client' / 'single_com'")
	port := flag.Int("port", 0, "port to be used")
	flag.Parse()

	*mode = strings.TrimSpace(*mode)

	switch *mode {
	case "":
		panic("choose a mode")
	case "server":
		StartServer(C.int(*port))
	case "client":
		StartClient(C.int(*port))
	case "single_com":
		bufferLen := 10000
		goBuffer := bytes.NewBuffer(make([]byte, bufferLen))
		cBuffer := C.CString(goBuffer.String())
		result := ServerClientSingleCommunication(C.int(1234), cBuffer, C.int(bufferLen))
		fmt.Println(C.GoString(cBuffer))
		if !bool(result) {
			os.Exit(1)
		}
	default:
		panic("wrong mode")
	}
}

//export StartClient
func StartClient(iPort C.int) {
	port := int(iPort)
	if port <= 0 {
		port = 1234
	}
	p :=  make([]byte, 2048)
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		panic(err)
	}
	logMessage(fmt.Sprintf("starting client on %s\n", addr))

	sent := 0
	for {
		sent++
		_, err = fmt.Fprintf(conn, "This is client's message %d", sent)
		if err != nil {
			logMessage(fmt.Sprintf("canno write to connection"))
		} else {
			_, err = bufio.NewReader(conn).Read(p)
			if err == nil {
				logMessage(fmt.Sprintf("Client: received %s\n", p))
			} else {
				logMessage(fmt.Sprintf("Some error %v\n", err))
			}
		}

		time.Sleep(time.Second)
	}
	conn.Close()
}

func sendResponse(conn *net.UDPConn, addr *net.UDPAddr, receivedMessage string) {
	_,err := conn.WriteToUDP([]byte(fmt.Sprintf("From server: Hello I got your message:\n\t'%s'", receivedMessage)), addr)
	if err != nil {
		logMessage(fmt.Sprintf("Couldn't send response %v", err))
	}
}

//export StartServer
func StartServer(iPort C.int) {
	port := int(iPort)
	if port <= 0 {
		port = 1234
	}
	p := make([]byte, 2048)
	addr := net.UDPAddr{
		Port: port,
		IP: net.ParseIP("127.0.0.1"),
	}
	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		logMessage(fmt.Sprintf("Some error %v\n", err))
		return
	}
	logMessage(fmt.Sprintf("starting server on %s\n", addr.String()))
	for {
		_, remoteAddr,err := ser.ReadFromUDP(p)
		logMessage(fmt.Sprintf("Read a message from '%v': '%s'\n", remoteAddr, p))
		if err !=  nil {
			logMessage(fmt.Sprintf("Some error  %v", err))
			continue
		}
		go sendResponse(ser, remoteAddr, string(p))
	}
}

//export ServerClientSingleCommunication
func ServerClientSingleCommunication(iPort C.int, iBuffer *C.char, iBufferLen C.int) C.bool {

	communicationIP := "127.0.0.1"
	communicationPort := int(iPort)

	logMessage("[ServerClientSingleCommunication] starting")

	writeToBuffer := func(message string) {
		if len(message) > int(iBufferLen) {
			logMessage(fmt.Sprintf("Cannot copy string of length: [%d] to a buffer of size: %d. Tried to copy '%s'\n",
				len(message), int(iBufferLen), message))
		} else {
			strCpy(message, iBuffer)
		}
	}

	addr := net.UDPAddr{
		Port: communicationPort,
		IP: net.ParseIP(communicationIP),
	}
	server, err := net.ListenUDP("udp", &addr)
	if err != nil {
		writeToBuffer(fmt.Sprintf("Error on ListenUDP %v\n", err))
		return C.bool(false)
	}
	defer server.Close()
	errChan := make(chan error)
	go func() {
		serverBuffer := make([]byte, 2048)
		_, remoteAddr,err := server.ReadFromUDP(serverBuffer)
		if err !=  nil {
			errChan <- fmt.Errorf("Some error %v\n", err)
			return
		}
		sendResponse(server, remoteAddr, string(serverBuffer))
	}()

	clientAddress := fmt.Sprintf("%s:%d", communicationIP, communicationPort)
	clientConnection, err := net.Dial("udp", clientAddress)
	if err != nil {
		writeToBuffer(fmt.Sprintf("Error on Dial %v\n", err))
		return C.bool(false)
	}
	defer clientConnection.Close()

	responseReceivedChan := make(chan string)
	var clientMessageSentTime time.Time
	go func() {
		clientMessageSentTime = time.Now()
		_, err = fmt.Fprintf(clientConnection, "This is client's message!")
		if err != nil {
			errChan <- fmt.Errorf("Error sending to server %v\n", err)
		}
		clientBuffer := make([]byte, 2048)
		_, err = bufio.NewReader(clientConnection).Read(clientBuffer)
		if err != nil {
			errChan <- fmt.Errorf("Error reading server response %v\n", err)
		}
		responseReceivedChan <- string(clientBuffer)
	}()

	select {
	case <-time.NewTicker(5 * time.Second).C:
		writeToBuffer(fmt.Sprintf("Error: timeout"))
		return C.bool(false)
	case err = <- errChan:
		writeToBuffer(fmt.Sprintf("error received: %s", err.Error()))
		return C.bool(false)
	case resp := <- responseReceivedChan:
		writeToBuffer(fmt.Sprintf("response from server (took %s): \n\t'%s'",
			time.Since(clientMessageSentTime).String(),
			resp))
		return C.bool(true)
	}
}

//export NewStringBuffer
func NewStringBuffer(iBufferLen C.int) *C.char {
	bufferMaxLen := int(iBufferLen)
	goBuffer := bytes.NewBuffer(make([]byte, bufferMaxLen))
	return C.CString(goBuffer.String())
}

//export FreeBuffer
func FreeBuffer(iBuffer *C.char){
	C.free(unsafe.Pointer(iBuffer))
}

//export RegisterLogger
func RegisterLogger(iLoggerFp C.loggerFunc) {
	logWrapper := func(iLoggerFp C.loggerFunc) func(format string){
		return func(format string) {
			// TODO possible leak for C.CString(format)
			C.bridge_logger(iLoggerFp, C.CString(format))
		}
	}

	externalLogger = logWrapper(iLoggerFp)
}

func logMessage(format string) {
	if externalLogger != nil {
		externalLogger(format)
	} else {
		fmt.Println(format)
	}
}

func strCpy(goString string, buffer *C.char) {
	cs := C.CString(goString)
	C.strcpy(buffer, cs)
	C.free(unsafe.Pointer(cs))
}

