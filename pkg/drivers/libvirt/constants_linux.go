package libvirt

const (
	DriverName    = "libvirt"
	DriverVersion = "0.12.8"

	connectionString = "qemu:///system"
	dnsmasqStatus    = "/var/lib/libvirt/dnsmasq/%s.status"
	DefaultMemory    = 8096
	DefaultCPUs      = 4
	DefaultNetwork   = "crc"
	DefaultCacheMode = "default"
	DefaultIOMode    = "threads"
	DefaultSSHUser   = "core"
	DefaultSSHPort   = 22
	DomainTemplate   = `<domain type='kvm'>
  <name>{{ .DomainName }}</name>
  <memory unit='MB'>{{ .Memory }}</memory>
  <vcpu placement='static'>{{ .CPU }}</vcpu>
  <features><acpi/><apic/><pae/></features>
  <cpu mode='host-passthrough'>
    <feature policy="disable" name="rdrand"/>
  </cpu>
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
    </disk>
    <graphics type='vnc' autoport='yes' listen='127.0.0.1'>
      <listen type='address' address='127.0.0.1'/>
    </graphics>
    <interface type='network'>
      <mac address='52:fd:fc:07:21:82'/>
      <source network='{{.Network}}'/>
      <model type='virtio'/>
    </interface>
    <console type='pty'></console>
    <channel type='pty'>
      <target type='virtio' name='org.qemu.guest_agent.0'/>
    </channel>
    <rng model='virtio'>
      <backend model='random'>/dev/urandom</backend>
    </rng>
  </devices>
</domain>`
)
