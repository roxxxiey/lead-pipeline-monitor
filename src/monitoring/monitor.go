package monitoring

import (
	"strings"
	"time"
)

func AddServiceForMonitor(name string, mon map[string][3]int) {
	if _, ok := mon[name]; !ok {
		mon[name] = [3]int{0, 0, 0}
	}
}

var lastErrorCode = map[string]string{}

func CheckCriticalErrors(serviceName, answerCode string, reaction time.Duration, list map[string][3]int) {
	val := list[serviceName]

	if !strings.HasPrefix(answerCode, "2") {
		val[0] = 1

		if answerCode == lastErrorCode[serviceName] {
			val[1]++
		} else {
			val[1] = 1
			lastErrorCode[serviceName] = answerCode
		}
	} else {
		val[0] = 0
		val[1] = 0
		lastErrorCode[serviceName] = ""
	}

	if reaction > 2*time.Second {
		val[2] = 1
	} else {
		val[2] = 0
	}

	list[serviceName] = val
}
