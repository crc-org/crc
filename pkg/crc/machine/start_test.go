package machine

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/containers/common/pkg/strongunits"
	"github.com/stretchr/testify/assert"
)

type MockedSSHRunner struct {
	mockedSSHCommandToOutputMap map[string]string
	mockedSSHCommandToError     map[string]error
	mockedSSHCommandToArgsMap   map[string]string
}

func (r *MockedSSHRunner) RunPrivileged(reason string, cmdAndArgs ...string) (string, string, error) {
	if r.mockedSSHCommandToError[reason] != nil {
		return "", "", r.mockedSSHCommandToError[reason]
	}
	r.mockedSSHCommandToArgsMap[reason] = strings.Join(cmdAndArgs, ",")
	output, ok := r.mockedSSHCommandToOutputMap[reason]
	if !ok {
		r.mockedSSHCommandToOutputMap[reason] = ""
	}
	return output, "", nil
}

func (r *MockedSSHRunner) Run(_ string, _ ...string) (string, string, error) {
	// No-op
	return "", "", nil
}

func (r *MockedSSHRunner) RunPrivate(_ string, _ ...string) (string, string, error) {
	// No-op
	return "", "", nil
}

func NewMockedSSHRunner() *MockedSSHRunner {
	return &MockedSSHRunner{
		mockedSSHCommandToOutputMap: map[string]string{},
		mockedSSHCommandToArgsMap:   map[string]string{},
		mockedSSHCommandToError:     map[string]error{},
	}
}

func TestGrowLVForMicroShift_WhenResizeAttemptFails_ThenThrowErrorExplainingWhereItFailed(t *testing.T) {
	testCases := []struct {
		name                                    string
		volumeGroupSizeOutput                   string
		logicalVolumeSizeOutput                 string
		physicalVolumeListCommandFailed         error
		physicalVolumeGroupSizeCommandFailed    error
		logicalVolumeSizeCommandFailed          error
		extendingLogicalVolumeSizeCommandFailed error
		expectedReturnedError                   error
	}{
		{"when listing physical volume devices failed, then throw error", "", "", errors.New("listing volumes failed"), nil, nil, nil, errors.New("listing volumes failed")},
		{"when fetching volume group size failed, then throw error", "", "", nil, errors.New("fetching volume group size failed"), nil, nil, errors.New("fetching volume group size failed")},
		{"when parsing volume group size failed, then throw error", "invalid", "", nil, nil, nil, nil, errors.New("strconv.Atoi: parsing \"invalid\": invalid syntax")},
		{"when fetching lv size failed, then throw error", "42966450176", "", nil, nil, errors.New("fetching lvm size failed"), nil, errors.New("fetching lvm size failed")},
		{"when parsing lv size failed, then throw error", "42966450176", "invalid", nil, nil, nil, nil, errors.New("strconv.Atoi: parsing \"invalid\": invalid syntax")},
		{"when extending lv size failed, then throw error", "42966450176", "16106127360", nil, nil, nil, errors.New("extending lv failed"), errors.New("extending lv failed")},
	}

	// Loop through each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			sshRunner := NewMockedSSHRunner()
			sshRunner.mockedSSHCommandToError["Resizing the physical volume(PV)"] = tc.physicalVolumeListCommandFailed
			sshRunner.mockedSSHCommandToError["Get the volume group size"] = tc.physicalVolumeGroupSizeCommandFailed
			sshRunner.mockedSSHCommandToOutputMap["Get the volume group size"] = tc.volumeGroupSizeOutput
			sshRunner.mockedSSHCommandToError["Get the size of root logical volume"] = tc.logicalVolumeSizeCommandFailed
			sshRunner.mockedSSHCommandToOutputMap["Get the size of root logical volume"] = tc.logicalVolumeSizeOutput
			sshRunner.mockedSSHCommandToError["Extending and resizing the logical volume(LV)"] = tc.extendingLogicalVolumeSizeCommandFailed

			// When
			err := growLVForMicroshift(sshRunner, "rhel/root", "/dev/vda5", 15)

			// Then
			assert.EqualError(t, err, tc.expectedReturnedError.Error(), "Error messages should match")
		})
	}
}

func TestGrowLVForMicroShift_WhenPhysicalVolumeAvailableForResize_ThenSizeToIncreaseIsCalculatedAndGrown(t *testing.T) {
	testCases := []struct {
		name                    string
		existingVolumeGroupSize strongunits.B
		oldPersistentVolumeSize strongunits.B
		persistentVolumeSize    int
		expectPVSizeToGrow      bool
		expectedIncreaseInSize  strongunits.B
	}{
		{"when disk size can not accommodate persistent volume size growth, then do NOT grow lv", strongunits.GiB(31).ToBytes(), strongunits.GiB(15).ToBytes(), 20, false, strongunits.B(0)},
		{"when disk size can accommodate persistent volume size growth, then grow lv", strongunits.GiB(41).ToBytes(), strongunits.GiB(15).ToBytes(), 20, true, strongunits.GiB(6).ToBytes()},
		{"when requested persistent volume size less than present persistent volume size, then do NOT grow lv", strongunits.GiB(31).ToBytes(), strongunits.GiB(20).ToBytes(), 15, false, strongunits.B(0)},
		{"when requested disk size less than present persistent volume size, then do NOT grow lv", strongunits.GiB(31).ToBytes(), strongunits.GiB(20).ToBytes(), 41, false, strongunits.B(0)},
		// https://github.com/crc-org/crc/issues/4186#issuecomment-2656129015
		{"when disk size has more space than requested persistent volume growth, then lv growth larger than requested", strongunits.GiB(45).ToBytes(), strongunits.GiB(15).ToBytes(), 20, true, strongunits.GiB(10).ToBytes()},
	}

	// Loop through each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			sshRunner := NewMockedSSHRunner()
			sshRunner.mockedSSHCommandToOutputMap["resizing the physical volume(PV)"] = "Physical volume \"/dev/vda5\" changed"
			sshRunner.mockedSSHCommandToOutputMap["Get the volume group size"] = fmt.Sprintf("%d", tc.existingVolumeGroupSize.ToBytes())
			sshRunner.mockedSSHCommandToOutputMap["Get the size of root logical volume"] = fmt.Sprintf("%d", tc.oldPersistentVolumeSize.ToBytes())

			// When
			err := growLVForMicroshift(sshRunner, "rhel/root", "/dev/vda5", tc.persistentVolumeSize)

			// Then
			assert.NoError(t, err)
			_, lvExpanded := sshRunner.mockedSSHCommandToOutputMap["Extending and resizing the logical volume(LV)"]
			assert.Equal(t, tc.expectPVSizeToGrow, lvExpanded)
			if tc.expectPVSizeToGrow {
				assert.Contains(t, sshRunner.mockedSSHCommandToArgsMap["Extending and resizing the logical volume(LV)"], fmt.Sprintf("+%db", tc.expectedIncreaseInSize.ToBytes()))
			}
		})
	}
}
