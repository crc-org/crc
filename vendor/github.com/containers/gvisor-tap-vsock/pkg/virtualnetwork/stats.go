package virtualnetwork

import (
	"gvisor.dev/gvisor/pkg/tcpip"
)

type Stats struct {
	BytesSent                  uint64   `json:"BytesSent"`
	BytesReceived              uint64   `json:"BytesReceived"`
	UnknownProtocolRcvdPackets uint64   `json:"UnknownProtocolRcvdPackets"`
	MalformedRcvdPackets       uint64   `json:"MalformedRcvdPackets"`
	DroppedPackets             uint64   `json:"DroppedPackets"`
	IP                         IPStats  `json:"IP"`
	TCP                        TCPStats `json:"TCP"`
	UDP                        UDPStats `json:"UDP"`
}

type IPStats struct {
	PacketsReceived                     uint64 `json:"PacketsReceived"`
	InvalidDestinationAddressesReceived uint64 `json:"InvalidDestinationAddressesReceived"`
	InvalidSourceAddressesReceived      uint64 `json:"InvalidSourceAddressesReceived"`
	PacketsDelivered                    uint64 `json:"PacketsDelivered"`
	PacketsSent                         uint64 `json:"PacketsSent"`
	OutgoingPacketErrors                uint64 `json:"OutgoingPacketErrors"`
	MalformedPacketsReceived            uint64 `json:"MalformedPacketsReceived"`
	MalformedFragmentsReceived          uint64 `json:"MalformedFragmentsReceived"`
}

type TCPStats struct {
	ActiveConnectionOpenings           uint64 `json:"ActiveConnectionOpenings"`
	PassiveConnectionOpenings          uint64 `json:"PassiveConnectionOpenings"`
	CurrentEstablished                 uint64 `json:"CurrentEstablished"`
	CurrentConnected                   uint64 `json:"CurrentConnected"`
	EstablishedResets                  uint64 `json:"EstablishedResets"`
	EstablishedClosed                  uint64 `json:"EstablishedClosed"`
	EstablishedTimedout                uint64 `json:"EstablishedTimedout"`
	ListenOverflowSynDrop              uint64 `json:"ListenOverflowSynDrop"`
	ListenOverflowAckDrop              uint64 `json:"ListenOverflowAckDrop"`
	ListenOverflowSynCookieSent        uint64 `json:"ListenOverflowSynCookieSent"`
	ListenOverflowSynCookieRcvd        uint64 `json:"ListenOverflowSynCookieRcvd"`
	ListenOverflowInvalidSynCookieRcvd uint64 `json:"ListenOverflowInvalidSynCookieRcvd"`
	FailedConnectionAttempts           uint64 `json:"FailedConnectionAttempts"`
	ValidSegmentsReceived              uint64 `json:"ValidSegmentsReceived"`
	InvalidSegmentsReceived            uint64 `json:"InvalidSegmentsReceived"`
	SegmentsSent                       uint64 `json:"SegmentsSent"`
	SegmentSendErrors                  uint64 `json:"SegmentSendErrors"`
	ResetsSent                         uint64 `json:"ResetsSent"`
	ResetsReceived                     uint64 `json:"ResetsReceived"`
	Retransmits                        uint64 `json:"Retransmits"`
	FastRecovery                       uint64 `json:"FastRecovery"`
	SACKRecovery                       uint64 `json:"SACKRecovery"`
	SlowStartRetransmits               uint64 `json:"SlowStartRetransmits"`
	FastRetransmit                     uint64 `json:"FastRetransmit"`
	Timeouts                           uint64 `json:"Timeouts"`
}

type UDPStats struct {
	PacketsReceived          uint64 `json:"PacketsReceived"`
	UnknownPortErrors        uint64 `json:"UnknownPortErrors"`
	ReceiveBufferErrors      uint64 `json:"ReceiveBufferErrors"`
	MalformedPacketsReceived uint64 `json:"MalformedPacketsReceived"`
	PacketsSent              uint64 `json:"PacketsSent"`
	PacketSendErrors         uint64 `json:"PacketSendErrors"`
}

func statsAsJSON(sent, received uint64, stats tcpip.Stats) Stats {
	return Stats{
		BytesSent:                  sent,
		BytesReceived:              received,
		UnknownProtocolRcvdPackets: stats.UnknownProtocolRcvdPackets.Value(),
		MalformedRcvdPackets:       stats.MalformedRcvdPackets.Value(),
		DroppedPackets:             stats.DroppedPackets.Value(),
		IP: IPStats{
			PacketsReceived:                     stats.IP.PacketsReceived.Value(),
			InvalidDestinationAddressesReceived: stats.IP.InvalidDestinationAddressesReceived.Value(),
			InvalidSourceAddressesReceived:      stats.IP.InvalidSourceAddressesReceived.Value(),
			PacketsDelivered:                    stats.IP.PacketsDelivered.Value(),
			PacketsSent:                         stats.IP.PacketsSent.Value(),
			OutgoingPacketErrors:                stats.IP.OutgoingPacketErrors.Value(),
			MalformedPacketsReceived:            stats.IP.MalformedPacketsReceived.Value(),
			MalformedFragmentsReceived:          stats.IP.MalformedFragmentsReceived.Value(),
		},
		TCP: TCPStats{
			ActiveConnectionOpenings:           stats.TCP.ActiveConnectionOpenings.Value(),
			PassiveConnectionOpenings:          stats.TCP.PassiveConnectionOpenings.Value(),
			CurrentEstablished:                 stats.TCP.CurrentEstablished.Value(),
			CurrentConnected:                   stats.TCP.CurrentConnected.Value(),
			EstablishedResets:                  stats.TCP.EstablishedResets.Value(),
			EstablishedClosed:                  stats.TCP.EstablishedClosed.Value(),
			EstablishedTimedout:                stats.TCP.EstablishedTimedout.Value(),
			ListenOverflowSynDrop:              stats.TCP.ListenOverflowSynDrop.Value(),
			ListenOverflowAckDrop:              stats.TCP.ListenOverflowAckDrop.Value(),
			ListenOverflowSynCookieSent:        stats.TCP.ListenOverflowSynCookieSent.Value(),
			ListenOverflowSynCookieRcvd:        stats.TCP.ListenOverflowSynCookieRcvd.Value(),
			ListenOverflowInvalidSynCookieRcvd: stats.TCP.ListenOverflowInvalidSynCookieRcvd.Value(),
			FailedConnectionAttempts:           stats.TCP.FailedConnectionAttempts.Value(),
			ValidSegmentsReceived:              stats.TCP.ValidSegmentsReceived.Value(),
			InvalidSegmentsReceived:            stats.TCP.InvalidSegmentsReceived.Value(),
			SegmentsSent:                       stats.TCP.SegmentsSent.Value(),
			SegmentSendErrors:                  stats.TCP.SegmentSendErrors.Value(),
			ResetsSent:                         stats.TCP.ResetsSent.Value(),
			ResetsReceived:                     stats.TCP.ResetsReceived.Value(),
			Retransmits:                        stats.TCP.Retransmits.Value(),
			FastRecovery:                       stats.TCP.FastRecovery.Value(),
			SACKRecovery:                       stats.TCP.SACKRecovery.Value(),
			SlowStartRetransmits:               stats.TCP.SlowStartRetransmits.Value(),
			FastRetransmit:                     stats.TCP.FastRetransmit.Value(),
			Timeouts:                           stats.TCP.Timeouts.Value(),
		},
		UDP: UDPStats{
			PacketsReceived:          stats.UDP.PacketsReceived.Value(),
			UnknownPortErrors:        stats.UDP.UnknownPortErrors.Value(),
			ReceiveBufferErrors:      stats.UDP.ReceiveBufferErrors.Value(),
			MalformedPacketsReceived: stats.UDP.MalformedPacketsReceived.Value(),
			PacketsSent:              stats.UDP.PacketsSent.Value(),
			PacketSendErrors:         stats.UDP.PacketSendErrors.Value(),
		},
	}
}
