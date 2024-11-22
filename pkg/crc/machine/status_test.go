package machine

import (
	"context"
	"testing"

	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	"github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/stretchr/testify/assert"
)

func TestCreateClusterStatusResultShouldSetOpenShiftStatusAsExpected(t *testing.T) {
	tests := []struct {
		name                  string
		vmStatus              state.State
		vmBundleType          preset.Preset
		expectedClusterStatus types.ClusterStatusResult
	}{
		{
			"MicroShift cluster running", state.Running, preset.Microshift, types.ClusterStatusResult{
				CrcStatus:            "Running",
				OpenshiftStatus:      "Running",
				OpenshiftVersion:     "v4.5.1",
				DiskUse:              int64(16),
				DiskSize:             int64(32),
				RAMUse:               int64(8),
				RAMSize:              int64(12),
				PersistentVolumeUse:  16,
				PersistentVolumeSize: 32,
				Preset:               preset.Microshift,
			},
		},
		{
			"MicroShift cluster stopped", state.Stopped, preset.Microshift, types.ClusterStatusResult{
				CrcStatus:            "Stopped",
				OpenshiftStatus:      "Stopped",
				OpenshiftVersion:     "v4.5.1",
				Preset:               preset.Microshift,
				DiskUse:              int64(0),
				DiskSize:             int64(0),
				RAMUse:               int64(0),
				RAMSize:              int64(0),
				PersistentVolumeUse:  0,
				PersistentVolumeSize: 0,
			},
		},
		{
			"MicroShift cluster error state", state.Error, preset.Microshift, types.ClusterStatusResult{
				CrcStatus:            "Error",
				OpenshiftStatus:      "Stopped",
				OpenshiftVersion:     "v4.5.1",
				Preset:               preset.Microshift,
				DiskUse:              int64(0),
				DiskSize:             int64(0),
				RAMUse:               int64(0),
				RAMSize:              int64(0),
				PersistentVolumeUse:  0,
				PersistentVolumeSize: 0,
			},
		},
		{
			"MicroShift cluster stopping state", state.Stopping, preset.Microshift, types.ClusterStatusResult{
				CrcStatus:            "Stopping",
				OpenshiftStatus:      "Stopped",
				OpenshiftVersion:     "v4.5.1",
				Preset:               preset.Microshift,
				DiskUse:              int64(0),
				DiskSize:             int64(0),
				RAMUse:               int64(0),
				RAMSize:              int64(0),
				PersistentVolumeUse:  0,
				PersistentVolumeSize: 0,
			},
		},
		{
			"MicroShift cluster starting state", state.Starting, preset.Microshift, types.ClusterStatusResult{
				CrcStatus:            "Starting",
				OpenshiftStatus:      "Stopped",
				OpenshiftVersion:     "v4.5.1",
				Preset:               preset.Microshift,
				DiskUse:              int64(0),
				DiskSize:             int64(0),
				RAMUse:               int64(0),
				RAMSize:              int64(0),
				PersistentVolumeUse:  0,
				PersistentVolumeSize: 0,
			},
		},
		{
			"OpenShift cluster running", state.Running, preset.OpenShift, types.ClusterStatusResult{
				CrcStatus:            "Running",
				OpenshiftStatus:      "Running",
				OpenshiftVersion:     "v4.5.1",
				Preset:               preset.OpenShift,
				DiskUse:              int64(16),
				DiskSize:             int64(32),
				RAMUse:               int64(8),
				RAMSize:              int64(12),
				PersistentVolumeUse:  0,
				PersistentVolumeSize: 0,
			},
		},
		{
			"OpenShift cluster stopped", state.Stopped, preset.OpenShift, types.ClusterStatusResult{
				CrcStatus:            "Stopped",
				OpenshiftStatus:      "Stopped",
				OpenshiftVersion:     "v4.5.1",
				Preset:               preset.OpenShift,
				DiskUse:              int64(0),
				DiskSize:             int64(0),
				RAMUse:               int64(0),
				RAMSize:              int64(0),
				PersistentVolumeUse:  0,
				PersistentVolumeSize: 0,
			},
		},
		{
			"OpenShift cluster errored", state.Error, preset.OpenShift, types.ClusterStatusResult{
				CrcStatus:            "Error",
				OpenshiftStatus:      "Stopped",
				OpenshiftVersion:     "v4.5.1",
				Preset:               preset.OpenShift,
				DiskUse:              int64(0),
				DiskSize:             int64(0),
				RAMUse:               int64(0),
				RAMSize:              int64(0),
				PersistentVolumeUse:  0,
				PersistentVolumeSize: 0,
			},
		},
		{
			"OpenShift cluster stopping state", state.Stopping, preset.OpenShift, types.ClusterStatusResult{
				CrcStatus:            "Stopping",
				OpenshiftStatus:      "Stopped",
				OpenshiftVersion:     "v4.5.1",
				Preset:               preset.OpenShift,
				DiskUse:              int64(0),
				DiskSize:             int64(0),
				RAMUse:               int64(0),
				RAMSize:              int64(0),
				PersistentVolumeUse:  0,
				PersistentVolumeSize: 0,
			},
		},
		{
			"OpenShift cluster starting state", state.Starting, preset.OpenShift, types.ClusterStatusResult{
				CrcStatus:            "Starting",
				OpenshiftStatus:      "Stopped",
				OpenshiftVersion:     "v4.5.1",
				Preset:               preset.OpenShift,
				DiskUse:              int64(0),
				DiskSize:             int64(0),
				RAMUse:               int64(0),
				RAMSize:              int64(0),
				PersistentVolumeUse:  0,
				PersistentVolumeSize: 0,
			},
		},
		{
			"OpenShift/OKD cluster running", state.Running, preset.OKD, types.ClusterStatusResult{
				CrcStatus:            "Running",
				OpenshiftStatus:      "Running",
				OpenshiftVersion:     "v4.5.1",
				Preset:               preset.OKD,
				DiskUse:              int64(16),
				DiskSize:             int64(32),
				RAMUse:               int64(8),
				RAMSize:              int64(12),
				PersistentVolumeUse:  0,
				PersistentVolumeSize: 0,
			},
		},
		{
			"OpenShift/OKD cluster stopped", state.Stopped, preset.OKD, types.ClusterStatusResult{
				CrcStatus:            "Stopped",
				OpenshiftStatus:      "Stopped",
				OpenshiftVersion:     "v4.5.1",
				Preset:               preset.OKD,
				DiskUse:              int64(0),
				DiskSize:             int64(0),
				RAMUse:               int64(0),
				RAMSize:              int64(0),
				PersistentVolumeUse:  0,
				PersistentVolumeSize: 0,
			},
		},
		{
			"OpenShift/OKD cluster errored", state.Error, preset.OKD, types.ClusterStatusResult{
				CrcStatus:            "Error",
				OpenshiftStatus:      "Stopped",
				OpenshiftVersion:     "v4.5.1",
				Preset:               preset.OKD,
				DiskUse:              int64(0),
				DiskSize:             int64(0),
				RAMUse:               int64(0),
				RAMSize:              int64(0),
				PersistentVolumeUse:  0,
				PersistentVolumeSize: 0,
			},
		},
		{
			"OpenShift/OKD cluster stopping state", state.Stopping, preset.OKD, types.ClusterStatusResult{
				CrcStatus:            "Stopping",
				OpenshiftStatus:      "Stopped",
				OpenshiftVersion:     "v4.5.1",
				Preset:               preset.OKD,
				DiskUse:              int64(0),
				DiskSize:             int64(0),
				RAMUse:               int64(0),
				RAMSize:              int64(0),
				PersistentVolumeUse:  0,
				PersistentVolumeSize: 0,
			},
		},
		{
			"OpenShift/OKD cluster starting state", state.Starting, preset.OKD, types.ClusterStatusResult{
				CrcStatus:            "Starting",
				OpenshiftStatus:      "Stopped",
				OpenshiftVersion:     "v4.5.1",
				Preset:               preset.OKD,
				DiskUse:              int64(0),
				DiskSize:             int64(0),
				RAMUse:               int64(0),
				RAMSize:              int64(0),
				PersistentVolumeUse:  0,
				PersistentVolumeSize: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			// When
			actualClusterStatusResult, err := createClusterStatusResult(tt.vmStatus, tt.vmBundleType, "v4.5.1", "127.0.0.1", 32, 16, 12, 8, 16, 32, func(context.Context, string) types.OpenshiftStatus { return types.OpenshiftRunning })

			// Then
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedClusterStatus.CrcStatus, actualClusterStatusResult.CrcStatus)
			assert.Equal(t, tt.expectedClusterStatus.OpenshiftStatus, actualClusterStatusResult.OpenshiftStatus)
			assert.Equal(t, tt.expectedClusterStatus.Preset, actualClusterStatusResult.Preset)
			assert.Equal(t, tt.expectedClusterStatus.OpenshiftVersion, actualClusterStatusResult.OpenshiftVersion)
			assert.Equal(t, tt.expectedClusterStatus.RAMSize, actualClusterStatusResult.RAMSize)
			assert.Equal(t, tt.expectedClusterStatus.RAMUse, actualClusterStatusResult.RAMUse)
			assert.Equal(t, tt.expectedClusterStatus.DiskSize, actualClusterStatusResult.DiskSize)
			assert.Equal(t, tt.expectedClusterStatus.DiskUse, actualClusterStatusResult.DiskUse)
			assert.Equal(t, tt.expectedClusterStatus.PersistentVolumeSize, actualClusterStatusResult.PersistentVolumeSize)
			assert.Equal(t, tt.expectedClusterStatus.PersistentVolumeUse, actualClusterStatusResult.PersistentVolumeUse)
		})
	}
}
