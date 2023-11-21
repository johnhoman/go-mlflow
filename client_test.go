package mlflow

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestClient_CreateExperiment(t *testing.T) {
	exp := &Experiment{Name: "create-experiment-" + uuid.NewString()[:7]}
	c := &client{address: mustParseURL("http://localhost:5000")}
	assert.NoError(t, c.CreateExperiment(exp))
	assert.NotEmpty(t, exp.ExperimentID)
	assert.False(t, exp.GetCreationTimestamp().IsZero())
	assert.True(t, exp.GetTags().Contains("metadata.namespace"))
	assert.False(t, strings.HasPrefix(exp.GetName(), "default"))
}

func TestClient_DeleteExperiment(t *testing.T) {

}
