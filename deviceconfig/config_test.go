package deviceconfig

import "testing"

func TestSortedAndNonOverlappingRootFiles(t *testing.T) {
	for dev, cfg := range DeviceConfigs {
		t.Run(dev, func(t *testing.T) {
			var startIncl, endExcl int64
			for _, rootDev := range cfg.RootDeviceFiles {
				if rootDev.Offset < startIncl {
					t.Fatalf("Unsorted rootfiles for %s", rootDev.Name)
				}
				if rootDev.Offset < endExcl {
					t.Fatalf("Overlap in rootfiles for %s (offset = %d, end offset of previous = %d)",
						rootDev.Name, rootDev.Offset, endExcl,
					)
				}
				if rootDev.Offset+rootDev.MaxLength > 8192*512 {
					t.Fatalf("Root file %s [%d, %d) overlaps boot/data partitions (starts at %d)",
						rootDev.Name, rootDev.Offset, rootDev.Offset+rootDev.MaxLength, 8192*512,
					)
				}
				startIncl = rootDev.Offset
				endExcl = rootDev.Offset + rootDev.MaxLength

				// GPT entries span LBA 0-33
				if rootDev.Offset <= 34*512 && 0 <= rootDev.Offset+rootDev.MaxLength {
					if !cfg.MBROnlyWithoutGPT {
						t.Fatalf("Root file %s [%d, %d) overlaps GPT header, but MBROnlyWithoutGPT is not set",
							rootDev.Name, rootDev.Offset, rootDev.Offset+rootDev.MaxLength)
					}
				}
			}
		})
	}
}
