package libvirt

const (
	StoragePoolTemplate = `<pool type='dir'>
	<name>{{ .PoolName }}</name>
		<target>
			<path>{{ .Dir }}</path>
		</target>
</pool>`

	NetworkTemplate = `<network>
	<name>{{ .DomainName }}</name>
	<uuid>49eee855-d342-46c3-9ed3-b8d1758814cd</uuid>
	<forward mode='nat'>
	  <nat>
		<port start='1024' end='65535'/>
	  </nat>
	</forward>
	<bridge name='tt0' stp='on' delay='0'/>
	<mac address='52:54:00:fd:be:d0'/>
	<domain name='test1.tt.testing' localOnly='yes'/>
	<dns>
	  <srv service='etcd-server-ssl' protocol='tcp' domain='test1.tt.testing' target='etcd-0.test1.tt.testing' port='2380' weight='10'/>
	  <host ip='{{ .IP }}'>
		<hostname>api.test1.tt.testing</hostname>
		<hostname>etcd-0.test1.tt.testing</hostname>
	  </host>
	</dns>
	<ip family='ipv4' address='192.168.126.1' prefix='24'>
	  <dhcp>
		<host mac='{{ .MAC }}' name='{{ .DomainName }}' ip='{{ .IP }}'/>
	  </dhcp>
	</ip>
  </network>`

	DomainTemplate = `<domain type='kvm' xmlns:qemu='http://libvirt.org/schemas/domain/qemu/1.0'>
  <name>{{ .DomainName }}</name>
  <memory unit='MB'>{{ .Memory }}</memory>
  <vcpu placement='static'>{{ .CPU }}</vcpu>
  <features><acpi/><apic/><pae/></features>
  <cpu mode='host-passthrough'></cpu>
  <os>
    <type arch='x86_64'>hvm</type>
    <boot dev='hd'/>
    <bootmenu enable='no'/>
  </os>
  <features>
    <acpi/>
    <apic/>
    <pae/>
  </features>
  <clock offset='utc'/>
  <on_poweroff>destroy</on_poweroff>
  <on_reboot>restart</on_reboot>
  <on_crash>destroy</on_crash>
  <devices>
    <disk type='file' device='disk'>
      <driver name='qemu' type='qcow2' cache='{{ .CacheMode }}' io='{{ .IOMode }}' />
      <source file='{{ .DiskPath }}'/>
      <target dev='vda' bus='virtio'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x05' function='0x0'/>
    </disk>
    <graphics type='vnc' autoport='yes' listen='127.0.0.1'>
      <listen type='address' address='127.0.0.1'/>
    </graphics>
    <controller type='usb' index='0' model='piix3-uhci'>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x01' function='0x2'/>
    </controller>
    <controller type='pci' index='0' model='pci-root'/>
    <controller type='virtio-serial' index='0'>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x04' function='0x0'/>
    </controller>
    <interface type='network'>
      <mac address='52:fd:fc:07:21:82'/>
      <source network='{{.Network}}'/>
      <model type='virtio'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x03' function='0x0'/>
    </interface>
    <serial type='pty'>
      <target type='isa-serial' port='0'>
        <model name='isa-serial'/>
      </target>
    </serial>
    <console type='pty'>
      <target type='serial' port='0'/>
    </console>
    <channel type='pty'>
      <target type='virtio' name='org.qemu.guest_agent.0'/>
      <address type='virtio-serial' controller='0' bus='0' port='1'/>
    </channel>
    <input type='mouse' bus='ps2'/>
    <input type='keyboard' bus='ps2'/>
    <graphics type='spice' autoport='yes'>
      <listen type='address'/>
    </graphics>
    <video>
      <model type='cirrus' vram='16384' heads='1' primary='yes'/>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x02' function='0x0'/>
    </video>
    <memballoon model='virtio'>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x06' function='0x0'/>
    </memballoon>
    <rng model='virtio'>
      <backend model='random'>/dev/random</backend>
      <address type='pci' domain='0x0000' bus='0x00' slot='0x07' function='0x0'/>
    </rng>
  </devices>
</domain>`
)

type PoolConfig struct {
	PoolName string
	Dir      string
}

type NetworkConfig struct {
	DomainName string
	MAC        string
	IP         string
}

type DomainConfig struct {
	DomainName string
	Memory     int
	CPU        int
	CacheMode  string
	IOMode     string
	DiskPath   string
	Network    string
}
