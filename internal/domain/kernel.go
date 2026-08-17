package domain

// OOMStats represents Out-Of-Memory events observed by Sentinel.
type OOMStats struct {
	SystemKillCount *uint64 `json:"system_kill_count,omitempty"`
	CgroupKillCount *uint64 `json:"cgroup_kill_count,omitempty"`
	CgroupOOMCount  *uint64 `json:"cgroup_oom_count,omitempty"`
}

// KernelStats represents kernel-level health information.
type KernelStats struct {
	OOM OOMStats `json:"oom"`

	// FilesystemErrors is reserved for the filesystem/kernel event
	// collector. It is not populated by the basic OOM collector.
	FilesystemErrors uint64 `json:"filesystem_errors"`
}
