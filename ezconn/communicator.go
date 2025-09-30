package ezconn

import (
	"log"
	"strings"
)

type ICommunicator interface {
	This() ICommunicator
	Protocol() string
	Init(addr string, processor IProcessor, args ...any) error
	// ProtocolInit(addr string, args ...any) error
	Close()
	SendToRemote(addr string, msg any) error
	// RegisterHandler(headerid any, msg any, handler func(conn IConn, req any)) error
}

func NewCommunicator(protocol, addr string, maxConnNum, pendingWriteNum, threadnum int, processor IProcessor) ICommunicator {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	if protocol == "" {
		log.Printf("[E]invalid arg\n")
		return nil
	}
	protocol = strings.ToLower(protocol)
	var communicator ICommunicator
	switch protocol {
	case "tcpclient":
		communicator = &TCPClient{}
	case "tcpserver":
		communicator = &TCPServer{}
	case "udp":
		communicator = &UDPCommunicator{}
	default:
		log.Printf("[E]unsurpported protocol(%s)\n", protocol)
		return nil
	}
	err := communicator.Init(addr, processor)
	if err != nil {
		log.Printf("[E]tcpserver init failed:%v\n", err)
		return nil
	}
	return communicator
}
