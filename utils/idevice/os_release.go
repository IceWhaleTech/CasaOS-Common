package idevice

import (
	"github.com/IceWhaleTech/CasaOS-Common/utils/file"
)

func GetOSRelease(name []string) (map[string]string, error) {
	osRelease, err := file.ReadOSRelease()
	if err != nil {
		return nil, err
	}
	data := make(map[string]string)
	if len(name) == 0 {
		return osRelease, nil
	}
	for _, v := range name {
		data[v] = osRelease[v]
	}
	return data, nil
}
