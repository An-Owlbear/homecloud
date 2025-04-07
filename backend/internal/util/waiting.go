package util

import (
	"errors"
	"time"
)

var TimeoutError = errors.New("timeout error")

func WaitUntil(condition func() (bool, error), timeout time.Duration, interval time.Duration) error {
	timePassed := time.Duration(0)
	for timePassed < timeout {
		startTime := time.Now()
		succeeded, err := condition()
		if err != nil {
			return err
		}
		if succeeded {
			return nil
		}

		time.Sleep(interval)
		timePassed += time.Since(startTime)
	}

	return TimeoutError
}
