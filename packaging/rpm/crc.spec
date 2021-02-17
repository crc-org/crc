# https://github.com/code-ready/admin-helper
%global goipath         github.com/code-ready/crc
Version:                1.22.0
%global 	 	openshift_suffix -4.6.15

%gometa

# debuginfo is not supported on RHEL with Go packages
%global debug_package %{nil}
%global _enable_debug_package 0

%global common_description %{expand:
CodeReady Container's executable}


%global golicenses    LICENSE
%global godocs        *.md


Name:           %{goname}
Release:        1%{?dist}
Summary:        CodeReady Container's executable
License:        APL 2.0
ExcludeArch:    armv7hl i686
URL:            %{gourl}
ExcludeArch:    s390x
Source0:        %{gosource}

#generate_buildrequires
#go_generate_buildrequires

BuildRequires: git-core
BuildRequires: go-srpm-macros
BuildRequires: make

Provides: bundled(golang(github.com/Azure/go-ansiterm)) = 0.0.0-0.20170929gitd6e3b3328b78
Provides: bundled(golang(github.com/Masterminds/semver)) = 1.5.0
Provides: bundled(golang(github.com/Microsoft/hcsshim)) = 0.8.10
Provides: bundled(golang(github.com/RangelReale/osincli)) = 0.0.0-0.20160924gitfababb0555f2
Provides: bundled(golang(github.com/VividCortex/ewma)) = 1.1.1
Provides: bundled(golang(github.com/YourFin/binappend)) = 0.0.0-0.20181105git0add4bf0b9ad
Provides: bundled(golang(github.com/alexbrainman/sspi)) = 0.0.0-0.20180613gite580b900e9f5
Provides: bundled(golang(github.com/apparentlymart/go-cidr)) = 1.1.0
Provides: bundled(golang(github.com/asaskevich/govalidator)) = 0.0.0-0.20200907git7a23bdc65eef
Provides: bundled(golang(github.com/cavaliercoder/grab)) = 2.0.0
Provides: bundled(golang(github.com/cheggaaa/pb/v3)) = 3.0.5
Provides: bundled(golang(github.com/code-ready/clicumber)) = 0.0.0-0.20210201gitcecb794bdf9a
Provides: bundled(golang(github.com/code-ready/gvisor-tap-vsock)) = 0.0.0-0.20210128gite5f886c34c9f
Provides: bundled(golang(github.com/code-ready/machine)) = 0.0.0-0.20210122git281ccfbb4566
Provides: bundled(golang(github.com/cucumber/gherkin-go/v11)) = 11.0.0
Provides: bundled(golang(github.com/cucumber/godog)) = 0.9.0
Provides: bundled(golang(github.com/cucumber/messages-go/v10)) = 10.0.3
Provides: bundled(golang(github.com/danieljoos/wincred)) = 1.1.0
Provides: bundled(golang(github.com/davecgh/go-spew)) = 1.1.1
Provides: bundled(golang(github.com/docker/docker)) = 1.4.2-0.20191121gitd1d5f6476656
Provides: bundled(golang(github.com/docker/docker)) = 1.4.2-0.20191121gitd1d5f6476656
Provides: bundled(golang(github.com/docker/go-units)) = 0.4.0
Provides: bundled(golang(github.com/docker/spdystream)) = 0.0.0-0.20160310git449fdfce4d96
Provides: bundled(golang(github.com/fatih/color)) = 1.10.0
Provides: bundled(golang(github.com/fsnotify/fsnotify)) = 1.4.9
Provides: bundled(golang(github.com/go-logr/logr)) = 0.2.0
Provides: bundled(golang(github.com/godbus/dbus/v5)) = 5.0.3
Provides: bundled(golang(github.com/gofrs/uuid)) = 3.2.0
Provides: bundled(golang(github.com/gogo/protobuf)) = 1.3.1
Provides: bundled(golang(github.com/golang/protobuf)) = 1.4.3
Provides: bundled(golang(github.com/google/btree)) = 1.0.0
Provides: bundled(golang(github.com/google/go-cmp)) = 0.5.4
Provides: bundled(golang(github.com/google/gofuzz)) = 1.1.0
Provides: bundled(golang(github.com/google/gopacket)) = 1.1.16
Provides: bundled(golang(github.com/google/tcpproxy)) = 0.0.0-0.20200125gitb6bb9b5b8252
Provides: bundled(golang(github.com/google/uuid)) = 1.1.2
Provides: bundled(golang(github.com/h2non/filetype)) = 1.1.2-0.20210202git95e28344e08c
Provides: bundled(golang(github.com/hashicorp/hcl)) = 1.0.0
Provides: bundled(golang(github.com/hectane/go-acl)) = 0.0.0-0.20190604gitda78bae5fc95
Provides: bundled(golang(github.com/hpcloud/tail)) = 1.0.0
Provides: bundled(golang(github.com/imdario/mergo)) = 0.3.11
Provides: bundled(golang(github.com/inconshreveable/mousetrap)) = 1.0.0
Provides: bundled(golang(github.com/json-iterator/go)) = 1.1.10
Provides: bundled(golang(github.com/kballard/go-shellquote)) = 0.0.0-0.20180428git95032a82bc51
Provides: bundled(golang(github.com/klauspost/compress)) = 1.11.7
Provides: bundled(golang(github.com/libvirt/libvirt-go-xml)) = 6.10.0
Provides: bundled(golang(github.com/linuxkit/virtsock)) = 0.0.0-0.20180830git8e79449dea07
Provides: bundled(golang(github.com/magiconair/properties)) = 1.8.4
Provides: bundled(golang(github.com/mattn/go-colorable)) = 0.1.8
Provides: bundled(golang(github.com/mattn/go-isatty)) = 0.0.12
Provides: bundled(golang(github.com/mattn/go-runewidth)) = 0.0.9
Provides: bundled(golang(github.com/mdlayher/vsock)) = 0.0.0-0.20200508git7ad3638b3fbc
Provides: bundled(golang(github.com/mgutz/ansi)) = 0.0.0-0.20200706gitd51e80ef957d
Provides: bundled(golang(github.com/miekg/dns)) = 1.1.35
Provides: bundled(golang(github.com/mitchellh/go-wordwrap)) = 1.0.0
Provides: bundled(golang(github.com/mitchellh/mapstructure)) = 1.4.0
Provides: bundled(golang(github.com/moby/term)) = 0.0.0-0.20200312git672ec06f55cd
Provides: bundled(golang(github.com/modern-go/concurrent)) = 0.0.0-0.20180306gitbacd9c7ef1dd
Provides: bundled(golang(github.com/modern-go/reflect2)) = 1.0.1
Provides: bundled(golang(github.com/onsi/ginkgo)) = 1.12.0
Provides: bundled(golang(github.com/onsi/gomega)) = 1.9.0
Provides: bundled(golang(github.com/openshift/api)) = 0.0.0-0.20200930gitdb52bc4ef99f
Provides: bundled(golang(github.com/openshift/containers-image)) = 0.0.0-0.20190130git76de87591e9d
Provides: bundled(golang(github.com/openshift/gssapi)) = 0.0.0-0.20161010git5fb4217df13b
Provides: bundled(golang(github.com/openshift/gssapi)) = 0.0.0-0.20161010git5fb4217df13b
Provides: bundled(golang(github.com/openshift/kubernetes)) = 1.20.0-0.alpha.20200922git4700daee7399
Provides: bundled(golang(github.com/openshift/kubernetes-apimachinery)) = 0.0.0-0.20200831gitc0eb43ac4a3e
Provides: bundled(golang(github.com/openshift/kubernetes-cli-runtime)) = 0.0.0-0.20200831git852eec47b608
Provides: bundled(golang(github.com/openshift/kubernetes-client-go)) = 0.0.0-0.20200908git9409de4c95e0
Provides: bundled(golang(github.com/openshift/kubernetes-kubectl)) = 0.0.0-0.20200922git1f5b2cd472a9
Provides: bundled(golang(github.com/openshift/library-go)) = 0.0.0-0.20201007gitfcceeb075980
Provides: bundled(golang(github.com/openshift/oc)) = 0.0.0-0.alpha.20201126git299b6af535d1
Provides: bundled(golang(github.com/pbnjay/memory)) = 0.0.0-0.20201129gitb12e5d931931
Provides: bundled(golang(github.com/pborman/uuid)) = 1.2.1
Provides: bundled(golang(github.com/pelletier/go-toml)) = 1.8.1
Provides: bundled(golang(github.com/pkg/browser)) = 0.0.0-0.20201207git0426ae3fba23
Provides: bundled(golang(github.com/pkg/errors)) = 0.9.1
Provides: bundled(golang(github.com/pmezard/go-difflib)) = 1.0.0
Provides: bundled(golang(github.com/segmentio/analytics-go)) = 3.1.0
Provides: bundled(golang(github.com/segmentio/backo-go)) = 0.0.0-0.20200129git23eae7c10bd3
Provides: bundled(golang(github.com/sirupsen/logrus)) = 1.7.0
Provides: bundled(golang(github.com/spf13/afero)) = 1.4.1
Provides: bundled(golang(github.com/spf13/cast)) = 1.3.1
Provides: bundled(golang(github.com/spf13/cobra)) = 1.1.1
Provides: bundled(golang(github.com/spf13/jwalterweatherman)) = 1.1.0
Provides: bundled(golang(github.com/spf13/pflag)) = 1.0.5
Provides: bundled(golang(github.com/spf13/viper)) = 1.7.1
Provides: bundled(golang(github.com/stretchr/testify)) = 1.7.0
Provides: bundled(golang(github.com/subosito/gotenv)) = 1.2.0
Provides: bundled(golang(github.com/xi2/xz)) = 0.0.0-0.20171230git48954b6210f8
Provides: bundled(golang(github.com/xtgo/uuid)) = 0.0.0-0.20140804gita0b114877d4c
Provides: bundled(golang(github.com/zalando/go-keyring)) = 0.1.1
Provides: bundled(golang(golang.org/x/crypto)) = 0.0.0-0.20201203gitbe400aefbc4c
Provides: bundled(golang(golang.org/x/net)) = 0.0.0-0.20201202gitc7110b5ffcbb
Provides: bundled(golang(golang.org/x/oauth2)) = 0.0.0-0.20201203git0b49973bad19
Provides: bundled(golang(golang.org/x/sync)) = 0.0.0-0.20201020git67f06af15bc9
Provides: bundled(golang(golang.org/x/sys)) = 0.0.0-0.20201204gited752295db88
Provides: bundled(golang(golang.org/x/term)) = 0.0.0-0.20201126git7de9c90e9dd1
Provides: bundled(golang(golang.org/x/text)) = 0.3.4
Provides: bundled(golang(golang.org/x/time)) = 0.0.0-0.20200630git3af7569d3a1e
Provides: bundled(golang(golang.org/x/xerrors)) = 0.0.0-0.20200804git5ec99f83aff1
Provides: bundled(golang(google.golang.org/appengine)) = 1.6.7
Provides: bundled(golang(google.golang.org/protobuf)) = 1.25.1-0.20201020gitd3470999428b
Provides: bundled(golang(gopkg.in/AlecAivazis/survey.v1)) = 1.8.8
Provides: bundled(golang(gopkg.in/fsnotify.v1)) = 1.4.7
Provides: bundled(golang(gopkg.in/inf.v0)) = 0.9.1
Provides: bundled(golang(gopkg.in/ini.v1)) = 1.62.0
Provides: bundled(golang(gopkg.in/tomb.v1)) = 1.0.0-0.20141024gitdd632973f1e7
Provides: bundled(golang(gopkg.in/yaml.v2)) = 2.4.0
Provides: bundled(golang(gopkg.in/yaml.v3)) = 3.0.0-0.20210107git496545a6307b
Provides: bundled(golang(gvisor.dev/gvisor)) = 0.0.0-0.20210106git99534ddb4e66
Provides: bundled(golang(howett.net/plist)) = 0.0.0-0.20201203git1454fab16a06
Provides: bundled(golang(k8s.io/api)) = 0.19.0
Provides: bundled(golang(k8s.io/apiextensions-apiserver)) = 0.19.0
Provides: bundled(golang(k8s.io/apiserver)) = 0.19.0
Provides: bundled(golang(k8s.io/cloud-provider)) = 0.19.0
Provides: bundled(golang(k8s.io/cluster-bootstrap)) = 0.19.0
Provides: bundled(golang(k8s.io/code-generator)) = 0.19.0
Provides: bundled(golang(k8s.io/component-base)) = 0.19.0
Provides: bundled(golang(k8s.io/cri-api)) = 0.19.0
Provides: bundled(golang(k8s.io/csi-translation-lib)) = 0.19.0
Provides: bundled(golang(k8s.io/klog/v2)) = 2.3.0
Provides: bundled(golang(k8s.io/kube-aggregator)) = 0.19.0
Provides: bundled(golang(k8s.io/kube-controller-manager)) = 0.19.0
Provides: bundled(golang(k8s.io/kube-proxy)) = 0.19.0
Provides: bundled(golang(k8s.io/kube-scheduler)) = 0.19.0
Provides: bundled(golang(k8s.io/kubelet)) = 0.19.0
Provides: bundled(golang(k8s.io/legacy-cloud-providers)) = 0.19.0
Provides: bundled(golang(k8s.io/metrics)) = 0.19.0
Provides: bundled(golang(k8s.io/node-api)) = 0.19.0
Provides: bundled(golang(k8s.io/sample-apiserver)) = 0.19.0
Provides: bundled(golang(k8s.io/sample-cli-plugin)) = 0.19.0
Provides: bundled(golang(k8s.io/sample-controller)) = 0.19.0
Provides: bundled(golang(k8s.io/utils)) = 0.0.0-0.20200729gitd5654de09c73
Provides: bundled(golang(sigs.k8s.io/structured-merge-diff/v4)) = 4.0.1
Provides: bundled(golang(sigs.k8s.io/yaml)) = 1.2.0


%description
%{common_description}

%gopkg

%prep
# order of these 3 steps is important, build breaks if they are moved around
%global archivename crc-%{version}%{?openshift_suffix}
%autosetup -S git -n crc-%{version}%{?openshift_suffix}
# with fedora macros: goprep -e -k
install -m 0755 -vd "$(dirname %{gobuilddir}/src/%{goipath})"
ln -fs "$(pwd)" "%{gobuilddir}/src/%{goipath}"


%build
export GOFLAGS="-mod=vendor"
make GO_EXTRA_LDFLAGS="-B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \n')" GO_EXTRA_BUILDFLAGS="" cross

%install
# with fedora macros: gopkginstall
install -m 0755 -vd                     %{buildroot}%{_bindir}
install -m 0755 -vp %{gobuilddir}/src/%{goipath}/out/linux-amd64/crc %{buildroot}%{_bindir}/

install -d %{buildroot}%{_datadir}/%{name}-redistributable/{linux,macos,windows}
install -m 0755 -vp %{gobuilddir}/src/%{goipath}/out/linux-amd64/crc %{buildroot}%{_datadir}/%{name}-redistributable/linux/
install -m 0755 -vp %{gobuilddir}/src/%{goipath}/out/windows-amd64/crc.exe %{buildroot}%{_datadir}/%{name}-redistributable/windows/
install -m 0755 -vp %{gobuilddir}/src/%{goipath}/out/macos-amd64/crc %{buildroot}%{_datadir}/%{name}-redistributable/macos/

%check
# with fedora macros: gocheck
export GOFLAGS="-mod=vendor"
# crc uses `go test -race`, which triggers gvisor issues on ppc64le:
# vendor/gvisor.dev/gvisor/pkg/sync/race_unsafe.go:47:6: missing function body
# and a ppc64le implementation is indeed missing on ppc64le:
# https://github.com/google/gvisor/blob/master/pkg/sync/race_unsafe.go
# https://github.com/google/gvisor/tree/master/pkg/sync
%ifnarch ppc64le
make test
%endif

%files
%license %{golicenses}
%doc
%{_bindir}/*
%{_datadir}/%{name}-redistributable/linux/*
%{_datadir}/%{name}-redistributable/macos/*
%{_datadir}/%{name}-redistributable/windows/*

#gopkgfiles

%changelog
* Mon Feb 15 2021 Christophe Fergeau <cfergeau@redhat.com> - 1.22.0-1
- Initial import in Fedora
