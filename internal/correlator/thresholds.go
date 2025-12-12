package correlator

type ThresholdSet struct {
	BanThreshold        uint32
	ChannelThreshold    uint32
	RoleThreshold       uint32
	WebhookThreshold    uint32
	VelocityThreshold   uint32
	MultiActorThreshold uint32
	_                   [40]byte
}

var DefaultThresholds = ThresholdSet{
	BanThreshold:        5,
	ChannelThreshold:    3,
	RoleThreshold:       3,
	WebhookThreshold:    10,
	VelocityThreshold:   15,
	MultiActorThreshold: 50,
}

type ThresholdTable [8192]ThresholdSet

var GlobalThresholds *ThresholdTable

func InitThresholds() {
	GlobalThresholds = &ThresholdTable{}
	for i := range GlobalThresholds {
		GlobalThresholds[i] = DefaultThresholds
	}
}

func GetThresholds() *ThresholdTable {
	if GlobalThresholds == nil {
		InitThresholds()
	}
	return GlobalThresholds
}

func (tt *ThresholdTable) Get(index uint32) *ThresholdSet {
	return &tt[index&8191]
}

func (tt *ThresholdTable) Set(index uint32, thresholds ThresholdSet) {
	tt[index&8191] = thresholds
}
