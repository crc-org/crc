package libvirt

const (
	NetworkTemplate = `<network>
	<name>{{ .NetworkName }}</name>
	<uuid>49eee855-d342-46c3-9ed3-b8d1758814cd</uuid>
	<forward mode='nat'>
	  <nat>
		<port start='1024' end='65535'/>
	  </nat>
	</forward>
	<bridge name='crc' stp='on' delay='0'/>
	<mac address='52:54:00:fd:be:d0'/>
	<ip family='ipv4' address='192.168.130.1' prefix='24'>
	  <dhcp>
		<host mac='{{ .MAC }}' name='{{ .HostName }}' ip='{{ .IP }}'/>
	  </dhcp>
	</ip>
  </network>`
)

type NetworkConfig struct {
	NetworkName string
	HostName    string
	MAC         string
	IP          string
}
