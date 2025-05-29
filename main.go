package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

func main() {
	serverAddr := "127.0.0.1:443"

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h2", "http/1.1"},
	}

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	tlsConn := tls.Client(conn, tlsConfig)

	if err := tlsConn.Handshake(); err != nil {
		log.Fatalf("TLS handshake failed: %v", err)
	}

	fmt.Println("TLS connection established successfully")

	handler := NewFrameHandler()

	settingFrame := Frame{
		Type:     SettingFrameType,
		StreamID: 0,
		Flags:    UnsetFlag,
		Data: any(SettingFrame{
			Params: map[SettingParam]uint32{
				SettingsMaxConcurrentStreams: 100,
				SettingsInitialWindowSize:    33554432,
				SettingsEnablePush:           0,
			},
		}),
	}
	headerFrame := Frame{
		Type:     HeaderFrameType,
		StreamID: 1,
		Flags:    EndStreamFlag | EndHeaderFlag,
		Data: any(HeaderFrame{
			HeaderFields: []HeaderField{
				{name: ":method:", value: "GET"},
				{name: ":path:", value: "/"},
				{name: ":scheme:", value: "https"},
				{name: ":authority:", value: "localhost"},
				{name: "user-agent", value: "go/h2"},
				{name: "accept", value: "*/*"},
			},
		}),
	}

	ackSettingFrame := Frame{
		Type:     SettingFrameType,
		StreamID: 0,
		Flags:    AckFlag,
		Data: any(SettingFrame{
			Params: map[SettingParam]uint32{},
		}),
	}

	// Send Magic
	_, err = tlsConn.Write([]byte("\x50\x52\x49\x20\x2a\x20\x48\x54\x54\x50\x2f\x32\x2e\x30\x0d\x0a\x0d\x0a\x53\x4d\x0d\x0a\x0d\x0a"))
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	_, err = handler.Encode(tlsConn, settingFrame)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	_, err = handler.Encode(tlsConn, headerFrame)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	_, err = handler.Encode(tlsConn, ackSettingFrame)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	frame := Frame{}
	if err := handler.Decode(tlsConn, &frame); err != nil {
		log.Fatal(err)
	}
	fmt.Println(frame)
	frame = Frame{}
	if err := handler.Decode(tlsConn, &frame); err != nil {
		log.Fatal(err)
	}
	fmt.Println(frame)
	frame = Frame{}
	if err := handler.Decode(tlsConn, &frame); err != nil {
		log.Fatal(err)
	}
	fmt.Println(frame)
	frame = Frame{}
	if err := handler.Decode(tlsConn, &frame); err != nil {
		log.Fatal(err)
	}
	fmt.Println(frame)
	frame = Frame{}
	if err := handler.Decode(tlsConn, &frame); err != nil {
		log.Fatal(err)
	}
	fmt.Println(frame)
}
