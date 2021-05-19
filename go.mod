module github.com/code-ready/crc

go 1.14

require (
	github.com/Masterminds/semver v1.5.0
	github.com/Microsoft/go-winio v0.4.15-0.20200908182639-5b44b70ab3ab
	github.com/YourFin/binappend v0.0.0-20181105185800-0add4bf0b9ad
	github.com/asaskevich/govalidator v0.0.0-20200907205600-7a23bdc65eef
	github.com/cavaliercoder/grab v2.0.0+incompatible
	github.com/cheggaaa/pb/v3 v3.0.5
	github.com/code-ready/admin-helper v0.0.0-20210317085515-2d7f65e95985
	github.com/code-ready/clicumber v0.0.0-20210201104241-cecb794bdf9a
	github.com/code-ready/gvisor-tap-vsock v0.0.0-20210308122700-d61f9aac135c
	github.com/code-ready/machine v0.0.0-20210122113819-281ccfbb4566
	github.com/cucumber/godog v0.9.0
	github.com/cucumber/messages-go/v10 v10.0.3
	github.com/docker/go-units v0.4.0
	github.com/fatih/color v1.10.0 // indirect
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/gorilla/handlers v1.4.2
	github.com/h2non/filetype v1.1.2-0.20210202002709-95e28344e08c
	github.com/hectane/go-acl v0.0.0-20190604041725-da78bae5fc95
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/jinzhu/copier v0.2.8
	github.com/klauspost/compress v1.11.7
	github.com/klauspost/cpuid v1.3.1
	github.com/kofalt/go-memoize v0.0.0-20200917044458-9b55a8d73e1c
	github.com/libvirt/libvirt-go-xml v6.10.0+incompatible
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mattn/go-colorable v0.1.8
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/mapstructure v1.4.0 // indirect
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/openshift/api v0.0.0-20201216151826-78a19e96f9eb
	github.com/openshift/client-go v0.0.0-20201214125552-e615e336eb49
	github.com/openshift/oc v0.0.0-alpha.0.0.20210319134016-c8d1f56fb3e2
	github.com/pbnjay/memory v0.0.0-20201129165224-b12e5d931931
	github.com/pborman/uuid v1.2.1
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/pkg/browser v0.0.0-20201207095918-0426ae3fba23
	github.com/pkg/errors v0.9.1
	github.com/segmentio/analytics-go v3.1.0+incompatible
	github.com/segmentio/backo-go v0.0.0-20200129164019-23eae7c10bd3 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/afero v1.4.1 // indirect
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8
	github.com/xtgo/uuid v0.0.0-20140804021211-a0b114877d4c // indirect
	github.com/zalando/go-keyring v0.1.1
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb // indirect
	golang.org/x/oauth2 v0.0.0-20201203001011-0b49973bad19 // indirect
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/sys v0.0.0-20201204225414-ed752295db88
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1
	golang.org/x/text v0.3.4
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/AlecAivazis/survey.v1 v1.8.8
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
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
