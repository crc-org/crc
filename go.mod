module github.com/code-ready/crc

go 1.15

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/Microsoft/go-winio v0.5.0
	github.com/StackExchange/wmi v0.0.0-20210224194228-fe8f1750fd46
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/YourFin/binappend v0.0.0-20181105185800-0add4bf0b9ad
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/cavaliercoder/grab v2.0.0+incompatible
	github.com/cheggaaa/pb/v3 v3.0.8
	github.com/code-ready/admin-helper v0.0.0-20210602075518-d22477037442
	github.com/code-ready/clicumber v0.0.0-20210201104241-cecb794bdf9a
	github.com/code-ready/gvisor-tap-vsock v0.0.0-20210308122700-d61f9aac135c
	github.com/code-ready/machine v0.0.0-20210616065635-eff475d32b9a
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf
	github.com/cucumber/godog v0.9.0
	github.com/cucumber/messages-go/v10 v10.0.3
	github.com/docker/go-units v0.4.0
	github.com/fatih/color v1.12.0 // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/godbus/dbus/v5 v5.0.4 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/gorilla/handlers v1.5.1
	github.com/h2non/filetype v1.1.2-0.20210312132201-8345d11fd249
	github.com/hectane/go-acl v0.0.0-20190604041725-da78bae5fc95
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/jinzhu/copier v0.3.0
	github.com/klauspost/compress v1.12.3
	github.com/klauspost/cpuid/v2 v2.0.6
	github.com/kofalt/go-memoize v0.0.0-20200917044458-9b55a8d73e1c
	github.com/libvirt/libvirt-go-xml v6.10.0+incompatible
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-colorable v0.1.8
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mdlayher/vsock v0.0.0-20210303205602-10d591861736
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/onsi/ginkgo v1.16.2
	github.com/onsi/gomega v1.12.0
	github.com/openshift/api v0.0.0-20210105115604-44119421ec6b
	github.com/openshift/client-go v0.0.0-20210112165513-ebc401615f47
	github.com/openshift/library-go v0.0.0-20210205203934-9eb0d970f2f4 // indirect
	github.com/openshift/oc v0.0.0-alpha.0.0.20210411025417-95881afb5df0
	github.com/pbnjay/memory v0.0.0-20201129165224-b12e5d931931
	github.com/pborman/uuid v1.2.1
	github.com/pelletier/go-toml v1.9.1 // indirect
	github.com/pkg/browser v0.0.0-20210115035449-ce105d075bb4
	github.com/pkg/errors v0.9.1
	github.com/segmentio/analytics-go v1.2.1-0.20201110202747-0566e489c7b9
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/afero v1.5.1 // indirect
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8
	github.com/zalando/go-keyring v0.1.1
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5 // indirect
	golang.org/x/oauth2 v0.0.0-20201203001011-0b49973bad19 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210525143221-35b2ab0089ea
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1
	golang.org/x/text v0.3.6
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/AlecAivazis/survey.v1 v1.8.8
	gopkg.in/ini.v1 v1.62.0 // indirect
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.20.1
)

replace (
	github.com/apcera/gssapi => github.com/openshift/gssapi v0.0.0-20161010215902-5fb4217df13b
	github.com/containers/image => github.com/openshift/containers-image v0.0.0-20190130162819-76de87591e9d
	// Taking changes from https://github.com/moby/moby/pull/40021 to accomodate new version of golang.org/x/sys.
	// Although the PR lists c3a0a3744636069f43197eb18245aaae89f568e5 as the commit with the fixes,
	// d1d5f6476656c6aad457e2a91d3436e66b6f2251 is more suitable since it does not break fsouza/go-clientdocker,
	// yet provides the same fix.
	github.com/docker/docker => github.com/docker/docker v1.4.2-0.20191121165722-d1d5f6476656

	k8s.io/apimachinery => github.com/openshift/kubernetes-apimachinery v0.0.0-20210108114224-194a87c5b03a
	k8s.io/client-go => github.com/openshift/kubernetes-client-go v0.0.0-20210108114446-0829bdd68114
	k8s.io/kubectl => github.com/openshift/kubernetes-kubectl v0.0.0-20210108115031-c0d78c0aeda3
)

replace github.com/mdlayher/vsock => github.com/cfergeau/vsock v0.0.0-20210707084117-4d87f8a20ba8
