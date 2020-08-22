---
description: This section focuses on gathering network packet information with netcap
---

# Packet Collection

Packets are fetched from an input source \(offline dump file or live from an interface\) and distributed via round-robin to a pool of workers. Each worker dissects all layers of a packet and writes the generated _protobuf_ audit records to the corresponding file. By default, the data is compressed with _gzip_ to save storage space and buffered to avoid an overhead due to excessive _syscalls_ for writing data to disk.

![Packet collection process](.gitbook/assets/netcap%20%283%29.svg)

## Encoders

Encoders take care of converting decoded packet data into protocol buffers for the audit records. Two types of decoders exist: the [Layer Encoder](https://github.com/dreadl0ck/netcap/blob/master/decoder/layerEncoder.go), which operates on _gopacket_ layer types, and the [Custom Decoder](https://github.com/dreadl0ck/netcap/blob/master/decoder/customEncoder.go), for which any desired logic can be implemented, including decoding application layer protocols that are not yet supported by gopacket or protocols that require stream reassembly.

## Unknown Protocols

Protocols that cannot be decoded will be dumped in the unknown.pcap file for later analysis, as this contains potentially interesting traffic that is not represented in the generated output. Separating everything that could not be understood makes it easy to reveal hidden communication channels, which are based on custom protocols.

## Error Log

Errors that happen in the gopacket lib due to malformed packets or implementation errors are written to disk in the errors.log file, and can be checked by the analyst later. Each packet that had a decoding error on at least one layer will be added to the errors.pcap. An entry to the error log has the following format:

```text
<UTC Timestamp>
Error: <Description>
Packet:
<full packet hex dump with layer information>
```

At the end of the error log, a summary of all errors and the number of their occurrences will be appended.

```text
...
<error name>: <number of occurrences>
...
```

## Inclusion and Exclusion of Encoders

The _-decoders_ flag can be used to list all available decoders. In case not all of them are desired, selective inclusion and exclusion is possible, by using the _-include_ and _-exclude_ flags.

List all decoders:

```text
$ net capture -decoders
custom: 11
+ TLSClientHello
+ TLSServerHello
+ LinkFlow
+ NetworkFlow
+ TransportFlow
+ HTTP
+ Flow
+ Connection
+ DeviceProfile
+ File
+ POP3
layer: 55
+ TCP
+ UDP
+ IPv4
+ IPv6
+ DHCPv4
+ DHCPv6
+ ICMPv4
+ ICMPv6
+ ICMPv6Echo
+ ICMPv6NeighborSolicitation
+ ICMPv6RouterSolicitation
+ DNS
+ ARP
+ Ethernet
+ Dot1Q
+ Dot11
+ NTP
+ SIP
+ IGMP
+ LLC
+ IPv6HopByHop
+ SCTP
+ SNAP
+ LinkLayerDiscovery
+ ICMPv6NeighborAdvertisement
+ ICMPv6RouterAdvertisement
+ EthernetCTP
+ EthernetCTPReply
+ LinkLayerDiscoveryInfo
+ IPSecAH
+ IPSecESP
+ Geneve
+ IPv6Fragment
+ VXLAN
+ USB
+ LCM
+ MPLS
+ Modbus
+ OSPF
+ OSPF
+ BFD
+ GRE
+ FDDI
+ EAP
+ VRRP
+ EAPOL
+ EAPOLKey
+ CiscoDiscovery
+ CiscoDiscoveryInfo
+ USBRequestBlockSetup
+ NortelDiscovery
+ CIP
+ Ethernet/IP
+ SMTP
+ Diameter
...
```

Include specific decoders \(only those named will be used\):

```text
$ net capture -read traffic.pcap -include Ethernet,Dot1Q,IPv4,IPv6,TCP,UDP,DNS
```

Exclude decoders \(this will prevent decoding of layers encapsulated by the excluded ones\):

```text
$ net capture -read traffic.pcap -exclude TCP,UDP
```

## Applying Berkeley Packet Filters

_Netcap_ will decode all traffic it is exposed to, therefore it might be desired to set a berkeley packet filter, to reduce the workload imposed on _Netcap_. This is possible for both live and offline operation. In case a [BPF](https://www.kernel.org/doc/Documentation/networking/filter.txt) should be set for offline use, the [gopacket/pcap](https://godoc.org/github.com/google/gopacket/pcap) package with bindings to the _libpcap_ will be used, since setting BPF filters is not yet supported by the native [pcapgo](https://godoc.org/github.com/google/gopacket/pcapgo) package.

When capturing live from an interface:

```text
$ net capture -iface en0 -bpf "host 192.168.1.1"
```

When reading offline dump files:

```text
$ net capture -read traffic.pcap -bpf "host 192.168.1.1"
```

