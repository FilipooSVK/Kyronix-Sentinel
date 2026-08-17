package domain

// DiskStats represents filesystem capacity information for one mount point.
type DiskStats struct {
	MountPoint     string `json:"mount_point"`
	Device         string `json:"device"`
	FilesystemType string `json:"filesystem_type"`

	TotalBytes     uint64  `json:"total_bytes"`
	UsedBytes      uint64  `json:"used_bytes"`
	AvailableBytes uint64  `json:"available_bytes"`
	UsedPercent    float64 `json:"used_percent"`
}
