/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package proto

const (
	RpcKaktusNodeCapabilities               = "NodeCapabilities"
	RpcKaktusGetLibvirtEndpoint             = "GetLibvirtEndpoint"
	RpcKaktusCreateInstance                 = "CreateInstance"
	RpcKaktusGetInstance                    = "GetInstance"
	RpcKaktusUpdateInstance                 = "UpdateInstance"
	RpcKaktusDeleteInstance                 = "DeleteInstance"
	RpcKaktusMigrateInstance                = "MigrateInstance"
	RpcKaktusGetInstanceState               = "GetInstanceState"
	RpcKaktusGetInstanceRemoteConnectionUrl = "GetInstanceRemoteConnectionUrl"
	RpcKaktusInstanceIsRunning              = "InstanceIsRunning"
	RpcKaktusInstanceOperation              = "InstanceOperation"
	RpcKaktusGetStoragePoolStats            = "GetStoragePoolStats"
	RpcKaktusGetVolumeInfos                 = "GetVolumeInfos"
	RpcKaktusResizeVolume                   = "ResizeVolume"
	RpcKaktusCreateRawVolume                = "CreateRawVolume"
	RpcKaktusCreateTemplateVolume           = "CreateTemplateVolume"
	RpcKaktusCreateOsVolume                 = "CreateOsVolume"
	RpcKaktusCreateIsoVolume                = "CreateIsoVolume"
	RpcKaktusDeleteVolume                   = "DeleteVolume"
	RpcKaktusUpdateIsoVolume                = "UpdateIsoVolume"
	RpcKaktusListFileSystems                = "ListFileSystems"
	RpcKaktusListFsSubVolumes               = "ListFsSubVolumes"
	RpcKaktusCreateFsSubVolume              = "CreateFsSubVolume"
	RpcKaktusCreateNfsBackends              = "CreateNfsBackends"
	RpcKaktusUpdateNfsBackends              = "UpdateNfsBackends"
	RpcKaktusDeleteFsSubVolume              = "DeleteFsSubVolume"
	RpcKaktusDeleteNfsBackends              = "DeleteNfsBackends"
)

/*
 * RPC NodeCapabilities()
 */

type KaktusNodeCapabilitiesArgs struct{}
type KaktusNodeCapabilitiesReply struct {
	Arch             string
	Vendor           string
	Model            string
	Sockets          uint8
	Cores            uint16
	Threads          uint16
	Memory           uint64
	GuestEmulator    string
	GuestMachineName string
}

/*
 * RPC GetLibvirtEndpoint()
 *
 * Reports how another Kaktus node can reach this node's libvirtd over the
 * network, for peer-to-peer live migration (see MigrateInstance()). Always
 * queried live, never cached: this is transient node config, not a
 * capability worth persisting.
 */

type KaktusGetLibvirtEndpointArgs struct{}
type KaktusGetLibvirtEndpointReply struct {
	Protocol string // "tcp" or "tls"
	Address  string
	Port     int
}

/*
 * RPC CreateInstance()
 */

type KaktusCreateInstanceArgs struct {
	Name string
	XML  string
}
type KaktusCreateInstanceReply struct{}

/*
 * RPC GetInstance()
 */

type KaktusGetInstanceArgs struct {
	Name       string
	Migratable bool
}

type KaktusGetInstanceReply struct {
	XML string
}

/*
 * RPC DeleteInstance()
 */

type KaktusDeleteInstanceArgs struct {
	Name string
}

type KaktusDeleteInstanceReply struct{}

/*
 * RPC UpdateInstance()
 */

type KaktusUpdateInstanceArgs struct {
	Name string
	XML  string
}

type KaktusUpdateInstanceReply struct{}

/*
 * RPC MigrateInstance()
 *
 * Sent to the *source* Kaktus node hosting the instance. DestinationURI is
 * a full libvirt connection URI (e.g. "qemu+tls://kaktus-042.internal:16514/system"),
 * built by the caller from a prior GetLibvirtEndpoint() call against the
 * destination node. This triggers a live, peer-to-peer migration: the
 * source libvirtd itself dials the destination and drives the handshake,
 * so no second connection is required by the agent.
 */

type KaktusMigrateInstanceArgs struct {
	Name           string
	DestinationURI string
}

type KaktusMigrateInstanceReply struct{}

/*
 * RPC GetInstanceState()
 */

type KaktusGetInstanceStateArgs struct {
	Name string
}

type KaktusGetInstanceStateReply struct {
	State  string
	Reason string
}

/*
 * RPC InstanceIsRunning()
 */

type KaktusInstanceIsRunningArgs struct {
	Name string
}

type KaktusInstanceIsRunningReply struct {
	Running bool
}

/*
 * RPC GetInstanceRemoteConnectionUrl()
 */

type KaktusGetInstanceRemoteConnectionUrlArgs struct {
	Name string
}

type KaktusGetInstanceRemoteConnectionUrlReply struct {
	URL string
}

/*
 * RPC InstanceOperation()
 */

type KaktusInstanceOperationArgs struct {
	Name   string
	Action KaktusInstanceOperation
}

type KaktusInstanceOperationReply struct{}

type KaktusInstanceOperation int

const (
	KaktusInstanceOpAutoStart KaktusInstanceOperation = iota
	KaktusInstanceOpStart
	KaktusInstanceOpSoftReboot
	KaktusInstanceOpHardReboot
	KaktusInstanceOpHardShutdown
	KaktusInstanceOpSoftShutdown
	KaktusInstanceOpPmSuspend
	KaktusInstanceOpPmResume
)

/*
 * RPC GetStoragePoolStats()
 */

type KaktusGetStoragePoolStatsArgs struct {
	Pool string
}

type KaktusGetStoragePoolStatsReply struct {
	Allocated uint64
	Available uint64
	Capacity  uint64
}

/*
 * RPC CreateRawVolume()
 */

type KaktusCreateRawVolumeArgs struct {
	Pool   string
	Volume string
	Size   uint64
}

type KaktusCreateRawVolumeReply struct{}

/*
 * RPC CreateTemplateVolume()
 */

type KaktusCreateTemplateVolumeArgs struct {
	Pool      string
	Volume    string
	SourceURL string
}

type KaktusCreateTemplateVolumeReply struct {
	Size uint64
}

/*
 * RPC CreateOsVolume()
 */

type KaktusCreateOsVolumeArgs struct {
	Pool     string
	Volume   string
	Size     uint64
	Template string
}

type KaktusCreateOsVolumeReply struct{}

/*
 * RPC CreateIsoVolume()
 */

type KaktusCreateIsoVolumeArgs struct {
	Pool    string
	Volume  string
	Size    uint64
	Content []byte
}

type KaktusCreateIsoVolumeReply struct{}

/*
 * RPC UpdateIsoVolume()
 */

type KaktusUpdateIsoVolumeArgs struct {
	Pool    string
	Volume  string
	Size    uint64
	Content []byte
}

type KaktusUpdateIsoVolumeReply struct{}

/*
 * RPC GetVolumeInfos()
 */

type KaktusGetVolumeInfosArgs struct {
	Pool   string
	Volume string
}

type KaktusGetVolumeInfosReply struct {
	Size uint64
}

/*
 * RPC ResizeVolume()
 */

type KaktusResizeVolumeArgs struct {
	Pool   string
	Volume string
	Size   uint64
}

type KaktusResizeVolumeReply struct{}

/*
 * RPC DeleteVolume()
 */

type KaktusDeleteVolumeArgs struct {
	Pool          string
	Volume        string
	WithSnapshots bool
}

type KaktusDeleteVolumeReply struct{}

// CephFS

/*
 * RPC ListFileSystems()
 */

type KaktusListFileSystemsArgs struct{}

type KaktusListFileSystemsReply struct {
	FS []string
}

/*
 * RPC ListFsSubVolumes()
 */

type KaktusListFsSubVolumesArgs struct {
	FS string
}

type KaktusListFsSubVolumesReply struct {
	SubVolumes []string
}

/*
 * RPC CreateFsSubVolume()
 */

type KaktusCreateFsSubVolumeArgs struct {
	FS        string
	SubVolume string
}

type KaktusCreateFsSubVolumeReply struct {
	Path      string
	BytesUsed int64
}

/*
 * RPC DeleteFsSubVolume()
 */

type KaktusDeleteFsSubVolumeArgs struct {
	FS        string
	SubVolume string
}

type KaktusDeleteFsSubVolumeReply struct{}

// NFS

/*
 * RPC CreateNfsBackends()
 */

type KaktusCreateNfsBackendsArgs struct {
	ID        string
	Name      string
	FS        string
	Path      string
	Access    string
	Protocols []int32
	Clients   []string
	Backends  []string
	Port      int
}

type KaktusCreateNfsBackendsReply struct{}

/*
 * RPC UpdateNfsBackends()
 */

type KaktusUpdateNfsBackendsArgs struct {
	ID        string
	Name      string
	FS        string
	Path      string
	Access    string
	Protocols []int32
	Clients   []string
	Backends  []string
	Port      int
}
type KaktusUpdateNfsBackendsReply struct{}

/*
 * RPC DeleteNfsBackends()
 */

type KaktusDeleteNfsBackendsArgs struct {
	ID        string
	Name      string
	FS        string
	Path      string
	Access    string
	Protocols []int32
	Clients   []string
	Backends  []string
	Port      int
}

type KaktusDeleteNfsBackendsReply struct{}
