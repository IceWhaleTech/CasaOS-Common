package external_test

// func TestGPUInfo(t *testing.T) {
// 	t.Skip()
// 	result, err := external.GPUInfoList()
// 	assert.NilError(t, err)
// 	assert.Equal(t, len(result), 1)
// }

// func TestGPUTwoImplementInfo(t *testing.T) {
// 	t.Skip()
// 	result, err := external.NvidiaGPUInfoListWithSMI()
// 	assert.NilError(t, err)
// 	result2, err := external.NvidiaGPUInfoListWithNVML()
// 	assert.NilError(t, err)

// 	assert.Equal(t, len(result), len(result2))
// 	for i := range result {
// 		assert.Equal(t, result[i].Name, result2[i].Name)
// 		assert.Equal(t, result[i].DriverVersion, result2[i].DriverVersion)
// 		assert.Equal(t, result[i].Name, result2[i].Name)
// 		assert.Equal(t, result[i].DisplayMode, result2[i].DisplayMode)
// 		assert.Equal(t, result[i].PowerLimit, result2[i].PowerLimit)
// 		assert.Equal(t, result[i].MemoryUtilization, result2[i].MemoryUtilization)
// 		assert.Equal(t, result[i].Utilization, result2[i].Utilization)
// 	}
// }
