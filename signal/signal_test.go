package signal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSignal(t *testing.T) {

	hbSub := newHeartBeat()
	toSub := newTimeoutSub()

	toOb := newTimeoutObserver()
	hbOb := newHeartBeatObserver(toSub)

	hbSub.attach(hbOb)
	toSub.attach(toOb)

	toSub.randomTimeoutForever()

	const count = 10

	for i := 0; i < count; i++ {
		hbSub.Beat()
	}

	assert.Equal(t, count, *hbOb.c)

}
