package correlator

import (
	"go-antinuke-2.0/internal/detectors"
)

type DetectorBindings struct {
	banDetector        *detectors.BanDetector
	channelDetector    *detectors.ChannelDeleteDetector
	roleDetector       *detectors.RoleDeleteDetector
	permDetector       *detectors.PermissionDetector
	velocityDetector   *detectors.VelocityDetector
	multiActorDetector *detectors.MultiActorDetector
	flagDetector       *detectors.FlagDetector
}

func NewDetectorBindings() *DetectorBindings {
	return &DetectorBindings{
		banDetector:        detectors.NewBanDetector(),
		channelDetector:    detectors.NewChannelDeleteDetector(),
		roleDetector:       detectors.NewRoleDeleteDetector(),
		permDetector:       detectors.NewPermissionDetector(),
		velocityDetector:   detectors.NewVelocityDetector(),
		multiActorDetector: detectors.NewMultiActorDetector(),
		flagDetector:       detectors.NewFlagDetector(),
	}
}

func (db *DetectorBindings) BindToCorrelator(c *Correlator) {
	c.banDetector = db.banDetector
	c.channelDetector = db.channelDetector
	c.roleDetector = db.roleDetector
	c.permDetector = db.permDetector
	c.velocityDetector = db.velocityDetector
	c.multiActorDetector = db.multiActorDetector
	c.flagDetector = db.flagDetector
}

func (db *DetectorBindings) GetBanDetector() *detectors.BanDetector {
	return db.banDetector
}

func (db *DetectorBindings) GetChannelDetector() *detectors.ChannelDeleteDetector {
	return db.channelDetector
}

func (db *DetectorBindings) GetRoleDetector() *detectors.RoleDeleteDetector {
	return db.roleDetector
}
