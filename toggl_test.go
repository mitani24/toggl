package toggl

import (
	"testing"
	"time"
)

func TestHasJustStarted(t *testing.T) {
	now := time.Now().Unix()
	var interval int64 = 30

	examples := []struct {
		activity          *Activity
		runningActivities map[int]string
		interval          int64
		expected          bool
	}{
		{&Activity{UserId: 1, Description: "", Duration: 1}, make(map[int]string), interval, false},
		{&Activity{UserId: 1, Description: "", Duration: -now}, make(map[int]string), interval, true},
		{&Activity{UserId: 1, Description: "", Duration: -now + interval/2}, make(map[int]string), interval, true},
		{&Activity{UserId: 1, Description: "", Duration: -now + interval}, make(map[int]string), interval, true},
		{&Activity{UserId: 1, Description: "test", Duration: -now + interval/2}, map[int]string{1: "test"}, interval, false},
		{&Activity{UserId: 1, Description: "test", Duration: -now + interval/2}, map[int]string{1: "test2"}, interval, true},
	}

	for i, e := range examples {
		res := e.activity.hasJustStarted(e.runningActivities, e.interval)
		if res != e.expected {
			t.Errorf("[%v]expected %v but %v", i, e.expected, res)
		}
	}
}

func TestHasJustFinished(t *testing.T) {
	examples := []struct {
		activity          *Activity
		runningActivities map[int]string
		expected          bool
	}{
		{&Activity{UserId: 1, Description: "", Duration: 1}, make(map[int]string), false},
		{&Activity{UserId: 1, Description: "", Duration: 1}, map[int]string{1: "test"}, false},
		{&Activity{UserId: 1, Description: "test", Duration: 1}, map[int]string{1: "test"}, false},
		{&Activity{UserId: 1, Description: "", Duration: 1, Stop: "2016-12-06T07:20:32+00:00"}, make(map[int]string), false},
		{&Activity{UserId: 1, Description: "", Duration: 1, Stop: "2016-12-06T07:20:32+00:00"}, map[int]string{1: "test"}, false},
		{&Activity{UserId: 1, Description: "test", Duration: 1, Stop: "2016-12-06T07:20:32+00:00"}, map[int]string{1: "test"}, true},
	}

	for i, e := range examples {
		res := e.activity.hasJustFinished(e.runningActivities)
		if res != e.expected {
			t.Errorf("[%v]expected %v but %v", i, e.expected, res)
		}
	}
}
