package correlator

type CounterSet struct {
	BanCount      uint32
	KickCount     uint32
	ChannelDelete uint32
	RoleDelete    uint32
	WebhookCreate uint32
	PermChange    uint32
	_             [40]byte
}

type CounterArray [8192]CounterSet

var GlobalCounters *CounterArray

func InitCounters() {
	GlobalCounters = &CounterArray{}
}

func GetCounters() *CounterArray {
	if GlobalCounters == nil {
		InitCounters()
	}
	return GlobalCounters
}

func (ca *CounterArray) Get(index uint32) *CounterSet {
	return &ca[index&8191]
}

func (ca *CounterArray) Reset(index uint32) {
	ca[index&8191] = CounterSet{}
}

func (ca *CounterArray) ResetAll() {
	for i := range ca {
		ca[i] = CounterSet{}
	}
}
