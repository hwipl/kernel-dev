package main

import (
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func createPacket(from, to net.IP, fromPort, toPort layers.UDPPort) []byte {
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
	if payload != nil {
		// with payload
		pl := gopacket.Payload(payload)
		err = gopacket.SerializeLayers(buf, opts, &ip, &udp, pl)
	} else {
		// without payload
		err = gopacket.SerializeLayers(buf, opts, &ip, &udp)
	}
	if err != nil {
		log.Fatal(err)
	}

	return buf.Bytes()
}

func main() {
	// create pcap handle
	device := "wg0"
	snaplen := int32(2048)
	promisc := false
	timeout := pcap.BlockForever
	pcapHandle, err := pcap.OpenLive(device, snaplen, promisc, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer pcapHandle.Close()

	// create packet
	from := net.IPv4(192, 168, 1, 1)
	to := net.IPv4(192, 168, 1, 2)
	fromPort := layers.UDPPort(1234)
	toPort := layers.UDPPort(1234)
	packet := createPacket(from, to, fromPort, toPort)

	// send packet/write packet to pcap handle
	err = pcapHandle.WritePacketData(packet)
	if err != nil {
		log.Fatal(err)
	}
}
