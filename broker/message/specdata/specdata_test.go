package specdata

import "testing"

func TestPackage(t *testing.T) {
	for name := range _bindata {
		t.Run(name, func(t *testing.T) {
			_, err := Asset(name)
			if err != nil {
				t.Error(err)
			}
		})
	}
}
