package main

import (
	"encoding/binary"
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/sys/unix"
)

// RawSocket stores a raw socket
type RawSocket struct {
	fd      int
	devName string
	dev     *net.Interface
	addr    *unix.SockaddrLinklayer
}

// Close closes the raw socket
func (r *RawSocket) Close() {
	unix.Close(r.fd)
}

// Send sends data out of the raw socket
func (r *RawSocket) Send(data []byte) {
	err := unix.Sendto(r.fd, data, 0, r.addr)
	if err != nil {
		log.Fatal(err)
	}
}

// htons converts a uint16 to network byte order (expects little endian system)
func htons(x uint16) uint16 {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, x)
	return binary.BigEndian.Uint16(buf)
}

// NewRawSocket creates a new raw socket for device
func NewRawSocket(device string) *RawSocket {
	// create raw socket
	fd, err := unix.Socket(unix.AF_PACKET, unix.SOCK_RAW,
		0)
	//      int(htons(unix.ETH_P_ALL)))
	if err != nil {
		log.Fatal(err)
	}

	// get loopback interface
	dev, err := net.InterfaceByName(device)
	if err != nil {
		log.Fatal(err)
	}

	// create sockaddr
	addr := &unix.SockaddrLinklayer{
		// Protocol: htons(unix.ETH_P_IP),
		Ifindex: dev.Index,
		Halen:   6,
	}

	// create raw socket and return it
	return &RawSocket{
		fd:      fd,
		devName: device,
		dev:     dev,
		addr:    addr,
	}
}

// createIPPacket creates an IP packet
func createIPPacket(from, to net.IP, fromPort, toPort layers.UDPPort) []byte {
	// prepare creation of packet
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	// create ip header
	ip := layers.IPv4{
		Version:  4,
		Flags:    layers.IPv4DontFragment,
		TTL:      64,
		Protocol: layers.IPProtocolUDP,
		SrcIP:    from,
		DstIP:    to,
	}

	// create udp header
	udp := layers.UDP{
		SrcPort: fromPort,
		DstPort: toPort,
	}
	udp.SetNetworkLayerForChecksum(&ip)

	// create payload
	payload := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

	// serialize packet to buffer
	var err error
	buf := gopacket.NewSerializeBuffer()
	pl := gopacket.Payload(payload)
	err = gopacket.SerializeLayers(buf, opts, &ip, &udp, pl)
	if err != nil {
		log.Fatal(err)
	}

	return buf.Bytes()
}

// createEthernetPacket creates an ethernet packet with an IP packet
func createEthernetPacket(fromMAC, toMAC string, ipPacket []byte) []byte {
	// prepare creation of packet
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	// create ethernet header
	srcMAC, err := net.ParseMAC(fromMAC)
	if err != nil {
		log.Fatal(err)
	}
	dstMAC, err := net.ParseMAC(toMAC)
	if err != nil {
		log.Fatal(err)
	}
	eth := layers.Ethernet{
		SrcMAC:       srcMAC,
		DstMAC:       dstMAC,
		EthernetType: layers.EthernetTypeIPv4,
	}

	// serialize packet to buffer
	buf := gopacket.NewSerializeBuffer()
	pl := gopacket.Payload(ipPacket)
	err = gopacket.SerializeLayers(buf, opts, &eth, pl)
	if err != nil {
		log.Fatal(err)
	}

	return buf.Bytes()
}

func main() {
	// create wireguard socket
	device := "wg0"
	sock := NewRawSocket(device)

	// create packet
	from := net.IPv4(192, 168, 1, 1)
	to := net.IPv4(192, 168, 1, 2)
	fromPort := layers.UDPPort(1234)
	toPort := layers.UDPPort(1234)
	packet := createIPPacket(from, to, fromPort, toPort)

	// send packet and close socket
	sock.Send(packet)
	sock.Close()

	// create ethernet socket
	device = "eth0"
	sock = NewRawSocket(device)

	// create packet
	from = net.IPv4(192, 168, 1, 1)
	to = net.IPv4(192, 168, 1, 2)
	fromPort = layers.UDPPort(1234)
	toPort = layers.UDPPort(1234)
	fromMAC := "00:00:5e:00:53:01"
	toMAC := "ff:ff:ff:ff:ff:ff"
	packet = createIPPacket(from, to, fromPort, toPort)
	packet = createEthernetPacket(fromMAC, toMAC, packet)

	// send packet and close socket
	sock.Send(packet)
	sock.Close()
}
