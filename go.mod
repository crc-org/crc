module github.com/code-ready/crc

go 1.16

require (
	github.com/AlecAivazis/survey/v2 v2.3.2
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/Microsoft/go-winio v0.5.0
	github.com/RedHatQE/gowinx v0.0.3
	github.com/StackExchange/wmi v1.2.1
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/YourFin/binappend v0.0.0-20181105185800-0add4bf0b9ad
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/cavaliercoder/grab v2.0.0+incompatible
	github.com/cheggaaa/pb/v3 v3.0.8
	github.com/code-ready/admin-helper v0.0.7
	github.com/code-ready/clicumber v0.0.0-20210201104241-cecb794bdf9a
	github.com/code-ready/machine v0.0.0-20210616065635-eff475d32b9a
	github.com/containers/gvisor-tap-vsock v0.1.1-0.20210816082554-ebd241aab69f
	github.com/coreos/go-systemd/v22 v22.3.2
	github.com/cucumber/godog v0.9.0
	github.com/cucumber/messages-go/v10 v10.0.3
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/docker/go-units v0.4.0
	github.com/fatih/color v1.13.0 // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/godbus/dbus/v5 v5.0.5 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/handlers v1.5.1
	github.com/h2non/filetype v1.1.2-0.20210917125640-7fafb18134ff
	github.com/hectane/go-acl v0.0.0-20190604041725-da78bae5fc95
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/jinzhu/copier v0.3.2
	github.com/klauspost/compress v1.13.6
	github.com/klauspost/cpuid/v2 v2.0.9
	github.com/kofalt/go-memoize v0.0.0-20210721235729-46a601ff34b8
	github.com/kr/pretty v0.3.0 // indirect
	github.com/libvirt/libvirt-go-xml v7.4.0+incompatible
	github.com/mattn/go-colorable v0.1.11
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mdlayher/vsock v0.0.0-20210303205602-10d591861736
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.16.0
	github.com/openshift/api v0.0.0-20210521075222-e273a339932a
	github.com/openshift/client-go v0.0.0-20210521082421-73d9475a9142
	github.com/openshift/oc v0.0.0-alpha.0.0.20210920131146-88e7eba36ac7
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58
	github.com/pborman/uuid v1.2.1
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8
	github.com/pkg/errors v0.9.1
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/segmentio/analytics-go v3.2.0+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.9.0
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8
	github.com/zalando/go-keyring v0.1.1
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/net v0.0.0-20210929193557-e81a3d93ecf6
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20211004093028-2c5d950f24ef
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211
	golang.org/x/text v0.3.7
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
)

replace (
	github.com/apcera/gssapi => github.com/openshift/gssapi v0.0.0-20161010215902-5fb4217df13b

	k8s.io/apimachinery => github.com/openshift/kubernetes-apimachinery v0.0.0-20210521074607-b6b98f7a1855
	k8s.io/client-go => github.com/openshift/kubernetes-client-go v0.0.0-20210521075216-71b63307b5df
	k8s.io/kubectl => github.com/openshift/kubernetes-kubectl v0.0.0-20210521075729-633333dfccda
)

replace github.com/mdlayher/vsock => github.com/cfergeau/vsock v0.0.0-20210707100525-4def293f672e
