package nameserver

import (
	"errors"
	"fmt"
	"net/netip"
	"syscall"
	"unsafe"
)

func GetDNSServers() (nameservers []netip.AddrPort) {
	const defaultDNSPort = 53
	defaultLocalNameservers := []netip.AddrPort{
		netip.AddrPortFrom(netip.AddrFrom4([4]byte{127, 0, 0, 1}), defaultDNSPort),
		netip.AddrPortFrom(netip.AddrFrom16([16]byte{0, 0, 0, 0, 0, 0, 0, 1}), defaultDNSPort),
	}

	adapterAddresses, err := getAdapterAddresses()
	if err != nil {
		return defaultLocalNameservers
	}

	for _, adapterAddress := range adapterAddresses {
		const statusUp = 0x01
		if adapterAddress.operStatus != statusUp {
			continue
		}

		if adapterAddress.firstGatewayAddress == nil {
			// Only search DNS servers for adapters having a gateway
			continue
		}

		dnsServerAddress := adapterAddress.firstDnsServerAddress
		for dnsServerAddress != nil {
			ip, ok := sockAddressToIP(dnsServerAddress.address.rawSockAddrAny)
			if !ok || ipIsSiteLocalAnycast(ip) {
				// fec0/10 IPv6 addresses are site local anycast DNS
				// addresses Microsoft sets by default if no other
				// IPv6 DNS address is set. Site local anycast is
				// deprecated since 2004, see
				// https://datatracker.ietf.org/doc/html/rfc3879
				dnsServerAddress = dnsServerAddress.next
				continue
			}

			nameserver := netip.AddrPortFrom(ip, defaultDNSPort)
			nameservers = append(nameservers, nameserver)
			dnsServerAddress = dnsServerAddress.next
		}
	}

	if len(nameservers) == 0 {
		return defaultLocalNameservers
	}
	return nameservers
}

var (
	errBufferOverflowUnexpected = errors.New("unexpected buffer overflowed because buffer was large enough")
)

func getAdapterAddresses() (
	adapterAddresses []*ipAdapterAddresses, err error) {
	var buffer []byte
	const initialBufferLength uint32 = 15000
	sizeVar := initialBufferLength

	for {
		buffer = make([]byte, sizeVar)
		err := runProcGetAdaptersAddresses(
			(*ipAdapterAddresses)(unsafe.Pointer(&buffer[0])),
			&sizeVar)
		if err != nil {
			if err.(syscall.Errno) == syscall.ERROR_BUFFER_OVERFLOW {
				if sizeVar <= uint32(len(buffer)) {
					return nil, fmt.Errorf("%w: buffer size variable %d is "+
						"equal or lower to the buffer current length %d",
						errBufferOverflowUnexpected, sizeVar, len(buffer))
				}
				continue
			}
			return nil, fmt.Errorf("getting adapters addresses: %w", err)
		}

		noDataFound := sizeVar == 0
		if noDataFound {
			return nil, nil
		}
		break
	}

	adapterAddress := (*ipAdapterAddresses)(unsafe.Pointer(&buffer[0]))
	for adapterAddress != nil {
		adapterAddresses = append(adapterAddresses, adapterAddress)
		adapterAddress = adapterAddress.next
	}

	return adapterAddresses, nil
}

var (
	procGetAdaptersAddresses = syscall.NewLazyDLL("iphlpapi.dll").
		NewProc("GetAdaptersAddresses")
)

func runProcGetAdaptersAddresses(adapterAddresses *ipAdapterAddresses,
	sizePointer *uint32) (errcode error) {
	const family = syscall.AF_UNSPEC
	const GAA_FLAG_SKIP_UNICAST = 0x0001
	const GAA_FLAG_SKIP_ANYCAST = 0x0002
	const GAA_FLAG_SKIP_MULTICAST = 0x0004
	const GAA_FLAG_SKIP_FRIENDLY_NAME = 0x0020
	const GAA_FLAG_INCLUDE_GATEWAYS = 0x0080
	const flags = GAA_FLAG_SKIP_UNICAST | GAA_FLAG_SKIP_ANYCAST |
		GAA_FLAG_SKIP_MULTICAST | GAA_FLAG_SKIP_FRIENDLY_NAME |
		GAA_FLAG_INCLUDE_GATEWAYS
	const reserved = 0
	// See https://learn.microsoft.com/en-us/windows/win32/api/iphlpapi/nf-iphlpapi-getadaptersaddresses
	r1, _, err := syscall.SyscallN(procGetAdaptersAddresses.Addr(),
		uintptr(family), uintptr(flags), uintptr(reserved),
		uintptr(unsafe.Pointer(adapterAddresses)),
		uintptr(unsafe.Pointer(sizePointer)))
	switch {
	case err != 0:
		return err
	case r1 != 0:
		return syscall.Errno(r1)
	default:
		return nil
	}
}

func sockAddressToIP(rawSockAddress *syscall.RawSockaddrAny) (ip netip.Addr, ok bool) {
	if rawSockAddress == nil {
		return netip.Addr{}, false
	}

	sockAddress, err := rawSockAddress.Sockaddr()
	if err != nil {
		return netip.Addr{}, false
	}

	switch sockAddress := sockAddress.(type) {
	case *syscall.SockaddrInet4:
		return netip.AddrFrom4([4]byte{
				sockAddress.Addr[0], sockAddress.Addr[1], sockAddress.Addr[2], sockAddress.Addr[3]}),
			true
	case *syscall.SockaddrInet6:
		return netip.AddrFrom16([16]byte{
				sockAddress.Addr[0], sockAddress.Addr[1], sockAddress.Addr[2], sockAddress.Addr[3],
				sockAddress.Addr[4], sockAddress.Addr[5], sockAddress.Addr[6], sockAddress.Addr[7],
				sockAddress.Addr[8], sockAddress.Addr[9], sockAddress.Addr[10], sockAddress.Addr[11],
				sockAddress.Addr[12], sockAddress.Addr[13], sockAddress.Addr[14], sockAddress.Addr[15]}),
			true
	default:
		return netip.Addr{}, false
	}
}

func ipIsSiteLocalAnycast(ip netip.Addr) bool {
	if !ip.Is6() {
		return false
	}

	array := ip.As16()
	return array[0] == 0xfe && array[1] == 0xc0
}

// See https://learn.microsoft.com/en-us/windows/win32/api/iptypes/ns-iptypes-ip_adapter_addresses_lh
type ipAdapterAddresses struct {
	// The order of fields DOES matter since they are read
	// raw from a bytes buffer. However, we are only interested
	// in a few select fields, so unneeded fields are either
	// named as "_" or removed if they are after the fields
	// we are interested in.
	_                     uint32
	_                     uint32
	next                  *ipAdapterAddresses
	_                     *byte
	_                     *ipAdapterUnicastAddress
	_                     *ipAdapterAnycastAddress
	_                     *ipAdapterMulticastAddress
	firstDnsServerAddress *ipAdapterDnsServerAdapter
	_                     *uint16
	_                     *uint16
	_                     *uint16
	_                     [syscall.MAX_ADAPTER_ADDRESS_LENGTH]byte
	_                     uint32
	_                     uint32
	_                     uint32
	_                     uint32
	operStatus            uint32
	_                     uint32
	_                     [16]uint32
	_                     *ipAdapterPrefix
	_                     uint64
	_                     uint64
	_                     *ipAdapterWinsServerAddress
	firstGatewayAddress   *ipAdapterGatewayAddress
	// Additional fields not needed here
}

type ipAdapterUnicastAddress struct {
	// The order of fields DOES matter since they are read raw
	// from a bytes buffer. However, we are not interested in
	// the value of any field, so they are all named as "_".
	_ uint32
	_ uint32
	_ *ipAdapterUnicastAddress
	_ ipAdapterSocketAddress
	_ int32
	_ int32
	_ int32
	_ uint32
	_ uint32
	_ uint32
	_ uint8
}

type ipAdapterAnycastAddress struct {
	// The order of fields DOES matter since they are read raw
	// from a bytes buffer. However, we are not interested in
	// the value of any field, so they are all named as "_".
	_ uint32
	_ uint32
	_ *ipAdapterAnycastAddress
	_ ipAdapterSocketAddress
}

type ipAdapterMulticastAddress struct {
	// The order of fields DOES matter since they are read raw
	// from a bytes buffer. However, we are only interested in
	// a few select fields, so unneeded fields are named as "_".
	_ uint32
	_ uint32
	_ *ipAdapterMulticastAddress
	_ ipAdapterSocketAddress
}

type ipAdapterDnsServerAdapter struct {
	// The order of fields DOES matter since they are read raw
	// from a bytes buffer. However, we are only interested in
	// a few select fields, so unneeded fields are named as "_".
	_       uint32
	_       uint32
	next    *ipAdapterDnsServerAdapter
	address ipAdapterSocketAddress
}

type ipAdapterPrefix struct {
	_ uint32
	_ uint32
	_ *ipAdapterPrefix
	_ ipAdapterSocketAddress
	_ uint32
}

type ipAdapterWinsServerAddress struct {
	_ uint32
	_ uint32
	_ *ipAdapterWinsServerAddress
	_ ipAdapterSocketAddress
}

type ipAdapterGatewayAddress struct {
	_ uint32
	_ uint32
	_ *ipAdapterGatewayAddress
	_ ipAdapterSocketAddress
}

type ipAdapterSocketAddress struct {
	rawSockAddrAny *syscall.RawSockaddrAny
}
