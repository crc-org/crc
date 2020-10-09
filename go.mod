module github.com/code-ready/crc

go 1.13

require (
	github.com/Masterminds/semver v1.5.0
	github.com/YourFin/binappend v0.0.0-20181105185800-0add4bf0b9ad
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/cavaliercoder/grab v2.0.0+incompatible
	github.com/cheggaaa/pb/v3 v3.0.4
	github.com/code-ready/clicumber v0.0.0-20200728062640-1203dda97f67
	github.com/code-ready/machine v0.0.0-20201008083613-1420a5d7f1e2
	github.com/cucumber/godog v0.9.0
	github.com/cucumber/messages-go/v10 v10.0.3
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/go-units v0.4.0
	github.com/h2non/filetype v1.1.0
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/libvirt/libvirt-go-xml v6.8.0+incompatible
	github.com/mattn/go-colorable v0.1.2
	github.com/openshift/library-go v0.0.0-20200715125344-100bf3ff5a19 // indirect
	github.com/openshift/oc v0.0.0-alpha.0.0.20200716022222-b66f2d3a6893
	github.com/pbnjay/memory v0.0.0-20190104145345-974d429e7ae4
	github.com/pborman/uuid v1.2.0
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cast v1.3.0
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.6.1
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd
	golang.org/x/text v0.3.2
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	gopkg.in/AlecAivazis/survey.v1 v1.8.5
	howett.net/plist v0.0.0-20181124034731-591f970eefbb
	k8s.io/client-go v8.0.0+incompatible
)

replace (
	github.com/Microsoft/hcsshim => github.com/Microsoft/hcsshim v0.8.7
	github.com/apcera/gssapi => github.com/openshift/gssapi v0.0.0-20161010215902-5fb4217df13b
	github.com/containers/image => github.com/openshift/containers-image v0.0.0-20190130162819-76de87591e9d
	// Taking changes from https://github.com/moby/moby/pull/40021 to accomodate new version of golang.org/x/sys.
	// Although the PR lists c3a0a3744636069f43197eb18245aaae89f568e5 as the commit with the fixes,
	// d1d5f6476656c6aad457e2a91d3436e66b6f2251 is more suitable since it does not break fsouza/go-clientdocker,
	// yet provides the same fix.
	github.com/docker/docker => github.com/docker/docker v1.4.2-0.20191121165722-d1d5f6476656

	k8s.io/api => k8s.io/api v0.18.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery => github.com/openshift/kubernetes-apimachinery v0.0.0-20200427132717-228307e8b83c
	k8s.io/apiserver => k8s.io/apiserver v0.18.2
	k8s.io/cli-runtime => github.com/openshift/kubernetes-cli-runtime v0.0.0-20200507115657-2fb95e953778
	k8s.io/client-go => github.com/openshift/kubernetes-client-go v0.0.0-20200507115529-5e2a2d83bced
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.18.2
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.2
	k8s.io/code-generator => k8s.io/code-generator v0.18.2
	k8s.io/component-base => k8s.io/component-base v0.18.2
	k8s.io/cri-api => k8s.io/cri-api v0.18.2
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.2
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.18.2
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.18.2
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.18.2
	k8s.io/kubectl => github.com/openshift/kubernetes-kubectl v0.0.0-20200507115706-2f87de22f81a
	k8s.io/kubelet => k8s.io/kubelet v0.18.2
	k8s.io/kubernetes => github.com/openshift/kubernetes v1.17.0-alpha.0.0.20200427141011-f0879866c662
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.18.2
	k8s.io/metrics => k8s.io/metrics v0.18.2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.18.2
)

replace vbom.ml/util => github.com/fvbommel/util v0.0.2
