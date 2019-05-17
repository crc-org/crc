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
	<domain name='crc.testing' localOnly='yes'/>
	<dns>
	  <srv service='etcd-server-ssl' protocol='tcp' domain='crc.testing' target='etcd-0.crc.testing' port='2380' weight='10'/>
	  <host ip='{{ .IP }}'>
		<hostname>api.crc.testing</hostname>
		<hostname>api-int.crc.testing</hostname>
		<hostname>etcd-0.crc.testing</hostname>
	  </host>
	</dns>
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
