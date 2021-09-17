package debug

import (
	"sort"
	"testing"

	"github.com/iter8-tools/etc3/controllers"
	"github.com/stretchr/testify/assert"
)

func TestSortIter8Logs(t *testing.T) {
	il := []controllers.Iter8Log{
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			Message:             "hello world",
			Precedence:          3,
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			Message:             "hello world",
			Precedence:          0,
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			Message:             "hello world",
			Precedence:          2,
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			Message:             "hello world",
			Precedence:          1,
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			Message:             "hello world again",
			Precedence:          1,
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			Message:             "hello world again and again",
			Precedence:          1,
		},
		{
			IsIter8Log:          true,
			ExperimentName:      "hello",
			ExperimentNamespace: "default",
			Source:              "task-runner",
			Message:             "hello world",
			Precedence:          4,
		},
	}

	sortedIl := []controllers.Iter8Log{il[1], il[3], il[4], il[5], il[2], il[0], il[6]}

	// sort logs by precedence
	sort.Sort(byPrecedence(il))

	assert.Equal(t, il, sortedIl)
}
