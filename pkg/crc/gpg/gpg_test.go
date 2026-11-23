package gpg

import (
	"testing"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/stretchr/testify/assert"
)

// testMsg is a v4 clearsigned sha256sum message as published on
// https://mirror.openshift.com/pub/cgw/crc/bundles/openshift/5.0.0/sha256sum.txt.sig
// It is signed with the Red Hat release key (constants.RedHatReleaseKey).
const (
	testMsg = `-----BEGIN PGP SIGNED MESSAGE-----
Hash: SHA256

422d32200318c8334f76eac037b080eb3b374e02006bdb4c69a61f93680283bc  crc_hyperv_5.0.0_amd64.crcbundle
01abffc5d27552ba1427a184e62596a0f2b208e976b0bbf320d8a25c199b7f60  crc_libvirt_5.0.0_amd64.crcbundle
5563d89fc5b89e59bd70dbab3f0f43855d502b345e46e51dc9d54869c664f305  crc_libvirt_5.0.0_arm64.crcbundle
5c189ddc359a6cad068c89773b832d38dda06ffba0e0ff49677ca60f80155355  crc_vfkit_5.0.0_amd64.crcbundle
58a84e8a9232e16c339fb83176c76a521a66e8753c34d3dce386b9c7a0a87921  crc_vfkit_5.0.0_arm64.crcbundle
-----BEGIN PGP SIGNATURE-----

iQIcBAEBCAAGBQJq8vQ1AAoJEBmeL5H9Qx1RGvUP+gPjAeQt/8UosRDe4vKvqkTi
hBKsOxy7J/Gg4DOmhQNNXBxGE+a13O6O6PiKji6LhapOPTeWxM/wM/YJ/Mc58gYH
zizoxpeJsEPNoPaTVgNpn8n0PSSV5FJzLcZBGt5mkYZ5TYz/axjWbGiVEJBJA23K
6XSnQRr06OgNb98wKvtZD1Lkdasr5fIiVomOss6PONi1htd76Wij9qOOH1embz2G
pZHcGcVvqmIkmVSfHQ6PevqAXLWwBc2UM6+/TX9otX5KkG8Bk6Mi2gSeZ2yW/tSE
/NJxBaiXaPyJ+q7koNdhfldYFdDPQPcxLXp0jAPuVGd9c9HEQ3Yc1fec/IW9FgiQ
THUuuEuSuF7ET7OgM8FvEtxjQqAxHKqBdYaHkxR+MnXFsmGLZlVHWwqe8ZYXjLLm
jyxtIZFwnpjzZfTBTepZZLSMA2a0QCNij4sPTN8rD9halnJIWJYKwoHO2GRL0Pyz
wnu5v+f5n1r0XIqcJ1daGLNdhrz6kSoeTqcLISW/BeQSjUu2SU9BPD1C2SbAYjYt
Eqf6zE2VyKjql1TTrpV5w3bX8K37+82h6IWxx81WoEqHT1NEhjDd7fZ9tXqmzr3z
l8lDn6FxQ2HhBPxxyR4qU2UJ51X1+eVbwHJyIEATFFbddMEN/V4WgkbWR18ig8QE
U+5jnCoRAgYftQs/Zd0K
=zpQw
-----END PGP SIGNATURE-----`

	expectedMsg = `422d32200318c8334f76eac037b080eb3b374e02006bdb4c69a61f93680283bc  crc_hyperv_5.0.0_amd64.crcbundle
01abffc5d27552ba1427a184e62596a0f2b208e976b0bbf320d8a25c199b7f60  crc_libvirt_5.0.0_amd64.crcbundle
5563d89fc5b89e59bd70dbab3f0f43855d502b345e46e51dc9d54869c664f305  crc_libvirt_5.0.0_arm64.crcbundle
5c189ddc359a6cad068c89773b832d38dda06ffba0e0ff49677ca60f80155355  crc_vfkit_5.0.0_amd64.crcbundle
58a84e8a9232e16c339fb83176c76a521a66e8753c34d3dce386b9c7a0a87921  crc_vfkit_5.0.0_arm64.crcbundle`
)

func TestVerify(t *testing.T) {
	msg, err := GetVerifiedClearsignedMsg(constants.RedHatReleaseKey, testMsg)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsg, msg)
}
