package correlator

type HotEvent struct {
	Type      uint8
	Priority  uint8
	Flags     uint16
	GuildIdx  uint32
	ActorIdx  uint32
	TargetID  uint64
	Metadata  uint64
	Timestamp int64
}

type HotCounters struct {
	Count    uint32
	LastTime int64
	Velocity uint32
	Reserved uint32
}

type FastPathData struct {
	Event       HotEvent
	Counters    HotCounters
	Thresholds  ThresholdSet
	TriggerMask uint32
	_           [28]byte
}

func NewFastPathData() *FastPathData {
	return &FastPathData{}
}

func (fpd *FastPathData) Reset() {
	fpd.Event = HotEvent{}
	fpd.Counters = HotCounters{}
	fpd.TriggerMask = 0
}

func (fpd *FastPathData) SetTrigger(flag uint32) {
	fpd.TriggerMask |= flag
}

func (fpd *FastPathData) HasTrigger() bool {
	return fpd.TriggerMask != 0
}

func (fpd *FastPathData) GetTriggerCount() uint32 {
	count := uint32(0)
	mask := fpd.TriggerMask
	for mask != 0 {
		count += mask & 1
		mask >>= 1
	}
	return count
}
