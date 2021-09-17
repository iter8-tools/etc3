package debug

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortIter8Logs(t *testing.T) {
	il := []Iter8Log{
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         3,
			Message:             "hello world",
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         0,
			Message:             "hello world",
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         2,
			Message:             "hello world",
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         1,
			Message:             "hello world",
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         1,
			Message:             "hello world again",
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         1,
			Message:             "hello world again and again",
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         4,
			Message:             "hello world",
		},
	}

	sortedIl := []Iter8Log{
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         0,
			Message:             "hello world",
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         1,
			Message:             "hello world",
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         1,
			Message:             "hello world again",
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         1,
			Message:             "hello world again and again",
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         2,
			Message:             "hello world",
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         3,
			Message:             "hello world",
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         4,
			Message:             "hello world",
		},
	}

	// sort logs by precedence
	sort.Sort(byPrecedence(il))

	assert.Equal(t, il, sortedIl)
}
