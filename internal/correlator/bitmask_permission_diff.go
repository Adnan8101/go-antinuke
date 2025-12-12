package correlator

const (
	PermAdmin          uint64 = 1 << 3
	PermManageGuild    uint64 = 1 << 5
	PermManageRoles    uint64 = 1 << 28
	PermBanMembers     uint64 = 1 << 2
	PermKickMembers    uint64 = 1 << 1
	PermManageChannels uint64 = 1 << 4
)

const CriticalMask = PermAdmin | PermManageGuild | PermManageRoles | PermBanMembers

func DiffPermissions(oldPerms, newPerms uint64) uint64 {
	return oldPerms ^ newPerms
}

func GetAddedPermissions(oldPerms, newPerms uint64) uint64 {
	diff := oldPerms ^ newPerms
	return diff & newPerms
}

func GetRemovedPermissions(oldPerms, newPerms uint64) uint64 {
	diff := oldPerms ^ newPerms
	return diff & oldPerms
}

func HasCriticalElevation(oldPerms, newPerms uint64) bool {
	added := GetAddedPermissions(oldPerms, newPerms)
	return (added & CriticalMask) != 0
}

func BranchlessHasCritical(added uint64) uint32 {
	masked := added & CriticalMask
	return uint32((masked | -masked) >> 63)
}

func PermissionScore(perms uint64) uint32 {
	score := uint32(0)

	if (perms & PermAdmin) != 0 {
		score += 100
	}
	if (perms & PermManageGuild) != 0 {
		score += 50
	}
	if (perms & PermManageRoles) != 0 {
		score += 40
	}
	if (perms & PermBanMembers) != 0 {
		score += 30
	}

	return score
}

func BranchlessPermScore(perms uint64) uint32 {
	admin := uint32((perms>>3)&1) * 100
	guild := uint32((perms>>5)&1) * 50
	roles := uint32((perms>>28)&1) * 40
	bans := uint32((perms>>2)&1) * 30
	return admin + guild + roles + bans
}
