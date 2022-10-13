package pkg

import (
	"errors"

	"github.com/sensu/sensu-plugin-sdk/sensu"
)

func GetResponseCodeFromThresholds(value int, thresholdCritical int, thresholdWarning int, thresholdDirection int) (int, error) {
	if thresholdDirection == 0 {
		if value == thresholdCritical {
			return sensu.CheckStateOK, nil
		} else {
			return sensu.CheckStateCritical, nil
		}
	} else if thresholdDirection < 0 {
		if thresholdWarning < thresholdCritical {
			return sensu.CheckStateCritical, errors.New("threshold direction is < 0, but warning threshold is bigger than critical threshold")
		}

		if value < thresholdCritical {
			return sensu.CheckStateCritical, nil
		} else if value < thresholdWarning {
			return sensu.CheckStateWarning, nil
		} else {
			return sensu.CheckStateOK, nil
		}
	} else {
		if thresholdWarning > thresholdCritical {
			return sensu.CheckStateCritical, errors.New("threshold direction is > 0, but warning threshold is less than critical threshold")
		}

		if value > thresholdCritical {
			return sensu.CheckStateCritical, nil
		} else if value > thresholdWarning {
			return sensu.CheckStateWarning, nil
		} else {
			return sensu.CheckStateOK, nil
		}
	}
}
