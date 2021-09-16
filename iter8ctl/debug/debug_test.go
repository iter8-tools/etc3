package debug

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortIter8Logs(t *testing.T) {
	il := []Iter8Log{
		{
			Iter8Log:            true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         3,
			Message:             "hello world",
			Priority:            1,
		},
		{
			Iter8Log:            true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         0,
			Message:             "hello world",
			Priority:            1,
		},
		{
			Iter8Log:            true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         2,
			Message:             "hello world",
			Priority:            1,
		},
		{
			Iter8Log:            true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         1,
			Message:             "hello world",
			Priority:            1,
		},
		{
			Iter8Log:            true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         1,
			Message:             "hello world again",
			Priority:            1,
		},
		{
			Iter8Log:            true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         1,
			Message:             "hello world again and again",
			Priority:            1,
		},
		{
			Iter8Log:            true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         4,
			Message:             "hello world",
			Priority:            1,
		},
	}

	sortedIl := []Iter8Log{
		{
			Iter8Log:            true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         0,
			Message:             "hello world",
			Priority:            1,
		},
		{
			Iter8Log:            true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         1,
			Message:             "hello world",
			Priority:            1,
		},
		{
			Iter8Log:            true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         1,
			Message:             "hello world again",
			Priority:            1,
		},
		{
			Iter8Log:            true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         1,
			Message:             "hello world again and again",
			Priority:            1,
		},
		{
			Iter8Log:            true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         2,
			Message:             "hello world",
			Priority:            1,
		},
		{
			Iter8Log:            true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         3,
			Message:             "hello world",
			Priority:            1,
		},
		{
			Iter8Log:            true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			ActionIndex:         4,
			Message:             "hello world",
			Priority:            1,
		},
	}

	// sort logs by precedence
	sort.Sort(byPrecedence(il))

	assert.Equal(t, il, sortedIl)
}
