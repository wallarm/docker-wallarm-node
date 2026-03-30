//go:build functional

package shared

import "time"

// ConditionFunc returns true when a condition is met.
type ConditionFunc func() (bool, error)

// Poll calls a condition function repeatedly on a polling interval until it returns true, returns an error
// or the timeout is reached. If the condition function returns true or an error before the timeout, Poll
// immediately returns with the true value or the error. If the timeout is exceeded, Poll returns false.
func Poll(interval time.Duration, timeout time.Duration, condition ConditionFunc) (bool, error) {
	timeoutCh := time.After(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCh:
			return false, nil
		case <-ticker.C:
			success, err := condition()
			if err != nil {
				return false, err
			}
			if success {
				return true, nil
			}
		}
	}
}
