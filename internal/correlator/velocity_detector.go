package correlator

type VelocityTracker struct {
	lastCount uint32
	lastTime  int64
	velocity  uint32
	_         [52]byte
}

type VelocityTable [8192]VelocityTracker

var GlobalVelocity *VelocityTable

func InitVelocity() {
	GlobalVelocity = &VelocityTable{}
}

func GetVelocity() *VelocityTable {
	if GlobalVelocity == nil {
		InitVelocity()
	}
	return GlobalVelocity
}

func (vt *VelocityTable) Update(index, currentCount uint32, currentTime int64) uint32 {
	tracker := &vt[index&8191]

	if tracker.lastTime == 0 {
		tracker.lastCount = currentCount
		tracker.lastTime = currentTime
		tracker.velocity = 0
		return 0
	}

	deltaCount := currentCount - tracker.lastCount
	deltaTime := currentTime - tracker.lastTime

	if deltaTime > 0 {
		velocity := uint32((int64(deltaCount) * 1000000000) / deltaTime)
		tracker.velocity = velocity
	}

	tracker.lastCount = currentCount
	tracker.lastTime = currentTime

	return tracker.velocity
}

func (vt *VelocityTable) Get(index uint32) uint32 {
	return vt[index&8191].velocity
}

func (vt *VelocityTable) Reset(index uint32) {
	vt[index&8191] = VelocityTracker{}
}
