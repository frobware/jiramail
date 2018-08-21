package jiraplus

import (
	"fmt"
)

func List(listFunc func(int) ([]interface{}, error), callbackFunc func(interface{}) error) (int, error) {
	i := 0

	objs, err := listFunc(i)
	if err != nil {
		return i, fmt.Errorf("unable to get objects: %s", err)
	}

	for len(objs) > 0 {
		for _, o := range objs {
			err = callbackFunc(o)
			if err != nil {
				return i, err
			}
		}

		i += len(objs)

		objs, err = listFunc(i)
		if err != nil {
			return i, fmt.Errorf("unable to get objects: %s", err)
		}
	}

	return i, nil
}
