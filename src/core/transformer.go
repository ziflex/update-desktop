package core

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/ziflex/bumblebee-cli/src/core/fs"
	"github.com/ziflex/bumblebee-cli/src/system/logging"
	"strings"
)

const (
	PRIMUSRUN_IDX = len(PRIMUSRUN) + 1
	OPTIRUN_IDX   = len(OPTIRUN) + 1
)

type Transformer struct {
	logger *logging.Logger
	dir    *fs.Directory
}

func NewTransformer(logger *logging.Logger, dir *fs.Directory) *Transformer {
	return &Transformer{logger, dir}
}

func (t *Transformer) Do(entries []*Entry, prefix string, strict bool) ([]string, error) {
	return t.transform(entries, prefix, strict)
}

func (t *Transformer) Revert(entries []*Entry) ([]string, error) {
	return t.transform(entries, "", false)
}

func (t *Transformer) transform(entries []*Entry, prefix string, strict bool) ([]string, error) {
	results := make([]string, 0, len(entries))

	if len(entries) == 0 {
		return results, nil
	}

	names := make([]string, len(entries))

	for i, entry := range entries {
		names[i] = entry.Name
	}

	files, err := t.dir.LoadFiles(names)

	if err != nil {
		return nil, errors.New(err)
	}

	filesToUpdate := make([]*fs.File, 0, len(files))

	for _, file := range files {
		currentValues, failure := file.GetValues()

		if failure != nil {
			err = failure
			break
		}

		if len(currentValues) == 0 {
			continue
		}

		isAddedToTheList := false
		nextValues := make(map[string]string)

		for name, currentValue := range currentValues {
			nextValue, update := t.transformValue(currentValue, prefix, strict)

			if !update {
				continue
			}

			nextValues[name] = nextValue

			if isAddedToTheList {
				continue
			}

			results = append(results, strings.Replace(file.Name(), ".desktop", "", -1))
			isAddedToTheList = true
		}

		if len(nextValues) > 0 {
			err = file.SetValues(nextValues)
			filesToUpdate = append(filesToUpdate, file)
		}
	}

	if err != nil {
		return nil, errors.New(err)
	}

	err = t.dir.SaveFiles(filesToUpdate)

	if err != nil {
		return nil, errors.New(err)
	}

	return results, nil
}

func (t *Transformer) transformValue(currentValue string, prefix string, strict bool) (string, bool) {
	update := false
	nextValue := currentValue

	if prefix != "" {
		if !IsGPUEnabled(currentValue) {
			nextValue = fmt.Sprintf("%s %s", prefix, currentValue)
			update = true
		} else {
			// It doesn't matter which prefix is used
			if !strict {
				return currentValue, false
			}

			// specific prefix must be used
			if strings.HasPrefix(currentValue, prefix) {
				return currentValue, false
			}

			// remove current prefix
			nextValue, update = t.transformValue(currentValue, "", false)

			// well... is shouldn't be false, but... who knows
			if !update {
				return currentValue, false
			}

			// set given prefix
			nextValue, update = t.transformValue(nextValue, prefix, false)
		}
	} else {
		startIndex := -1
		var slicedStr []string

		if strings.HasPrefix(currentValue, PRIMUSRUN) {
			startIndex = PRIMUSRUN_IDX
		} else if strings.HasPrefix(currentValue, OPTIRUN) {
			startIndex = OPTIRUN_IDX
		}

		if startIndex > -1 {
			slicedStr = strings.Split(currentValue, "")
			nextValue = strings.Join(slicedStr[startIndex:], "")
			update = true
		}
	}

	return nextValue, update
}

func IsGPUEnabled(value string) bool {
	if strings.HasPrefix(value, PRIMUSRUN) {
		return true
	}

	if strings.HasPrefix(value, OPTIRUN) {
		return true
	}

	return false
}
