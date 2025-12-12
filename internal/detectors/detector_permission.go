package detectors

const (
	PermAdministrator   uint64 = 1 << 3
	PermManageGuild     uint64 = 1 << 5
	PermManageRoles     uint64 = 1 << 28
	PermManageChannels  uint64 = 1 << 4
	PermManageWebhooks  uint64 = 1 << 29
	PermBanMembers      uint64 = 1 << 2
	PermKickMembers     uint64 = 1 << 1
	PermMentionEveryone uint64 = 1 << 17
)

const CriticalPermMask = PermAdministrator | PermManageGuild | PermManageRoles | PermBanMembers

type PermissionDetector struct {
	permCache map[uint64]uint64
}

func NewPermissionDetector() *PermissionDetector {
	return &PermissionDetector{
		permCache: make(map[uint64]uint64),
	}
}

func (d *PermissionDetector) Detect(roleID, newPerms uint64) (bool, uint64) {
	oldPerms, exists := d.permCache[roleID]
	if !exists {
		d.permCache[roleID] = newPerms
		return false, 0
	}

	diff := oldPerms ^ newPerms
	addedPerms := diff & newPerms

	criticalAdded := addedPerms & CriticalPermMask

	d.permCache[roleID] = newPerms

	return criticalAdded != 0, criticalAdded
}

func (d *PermissionDetector) CheckElevation(oldPerms, newPerms uint64) bool {
	diff := oldPerms ^ newPerms
	addedPerms := diff & newPerms

	return (addedPerms & CriticalPermMask) != 0
}

func (d *PermissionDetector) GetAddedPermissions(oldPerms, newPerms uint64) uint64 {
	diff := oldPerms ^ newPerms
	return diff & newPerms
}

func (d *PermissionDetector) GetRemovedPermissions(oldPerms, newPerms uint64) uint64 {
	diff := oldPerms ^ newPerms
	return diff & oldPerms
}

func HasPermission(perms uint64, perm uint64) bool {
	return (perms & perm) != 0
}

func HasAdministrator(perms uint64) bool {
	return (perms & PermAdministrator) != 0
}

func BranchlessHasPerm(perms, targetPerm uint64) uint32 {
	masked := perms & targetPerm
	return uint32((masked | -masked) >> 63)
}
