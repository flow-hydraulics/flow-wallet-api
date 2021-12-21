package common

import (
	"math/rand"
	"time"
)

func SleepRandomMillis(min, max int) {
	random_int := rand.Intn(max-min) + min
	time.Sleep(time.Millisecond * time.Duration(random_int))
}
