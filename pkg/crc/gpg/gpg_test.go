package gpg

import (
	"testing"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/stretchr/testify/assert"
)

const (
	testMsg = `-----BEGIN PGP SIGNED MESSAGE-----
Hash: SHA256

a8267d09eac58e3c7f0db093f3cba83091390e5ac623b0a0282f6f55102b7681  crc_hyperv_4.12.9_amd64.crcbundle
f57ab331ad092d8cb1f354b4308046c5ffd15bd143b19f841cb64b0fda89db67  crc_libvirt_4.12.9_amd64.crcbundle
7e83b6a4c4da6766b6b4981655d4bb38fd8f9da36ef1a5d16d017cd07d6ee7e9  crc_vfkit_4.12.9_amd64.crcbundle
412d20e4969e872c24b14e55cbaa892848a1657b95a20f4af8ad4629ffdf73ab  crc_vfkit_4.12.9_arm64.crcbundle
-----BEGIN PGP SIGNATURE-----
Version: GnuPG v1

iQIVAwUBZCQdgxmeL5H9Qx1RAQhUrxAAiLMQqfTZliHzSTmLsgTYj5YOoRly8ax1
lEGxJZ6jTmrkDaC60S76wAoobD7zbVUifNp0XDYs8f7x6VVANpARoKp/p7NDtWP5
K19NhlG6Hip7/RmzUnEfEMH6sYcrNr0mbNQpsTJpje13u7HuwEXEpiceyU3GcDaA
VcdARo6hcP7bZIieHwJtpWrb8knc/OE2lCMKjanBbeC4+vZtj0xU1kpZTsBq8q84
WFdkR7C76XXhdRnuQqBBTbTuXGTI0OjB963Rx0+3Ej1liiDokH7JWvlUVaX7MKS3
IRf4X7q/Q3acsBtfJ9aNjrTCZJoMNg+F/eCbj+iVpXUK5rdOEzDKQitkMrB8VYgF
SCMe+FMyNqDS6MewR+wJzzKoVZDf6SDuAlSej1FthJ4QZAljuXlpRrDbFq2SRiFz
WgVipg15Bl6vTftjOnCkwHg/GQt1bQuzudGyulRlTu/Nnezvf9/sej++91z2aA86
nJnofMLLw3KkA1HCCgIiE5upMvjzLUobKAFxYyaDal6gNPG9S/l8N91UHujhKCpg
otaQ9Awtg7Z/F8pCLH08Nen4J/CqYaG+JHRORm/i17eD3qBEc+EgIZ7m0sLvm1aV
1Tg7G/6pu+LdPIYvJQKSgGuI8eP/p1zc8zzgHtaSWd2AVL3M/iOjteaba8eU5VCd
uaTo5yjgVLY=
=x4q3
-----END PGP SIGNATURE-----`
	expectedMsg = `a8267d09eac58e3c7f0db093f3cba83091390e5ac623b0a0282f6f55102b7681  crc_hyperv_4.12.9_amd64.crcbundle
f57ab331ad092d8cb1f354b4308046c5ffd15bd143b19f841cb64b0fda89db67  crc_libvirt_4.12.9_amd64.crcbundle
7e83b6a4c4da6766b6b4981655d4bb38fd8f9da36ef1a5d16d017cd07d6ee7e9  crc_vfkit_4.12.9_amd64.crcbundle
412d20e4969e872c24b14e55cbaa892848a1657b95a20f4af8ad4629ffdf73ab  crc_vfkit_4.12.9_arm64.crcbundle`
)

func TestVerify(t *testing.T) {
	msg, err := GetVerifiedClearsignedMsgV3(constants.RedHatReleaseKey, testMsg)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsg, msg)
}
