package testutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Events(t *testing.T) {
	em := NewEventManager("compose")
	em.AddEvent(&Event{
		Name:      "rhel-9.2",
		Suite:     "compose",
		ClassName: "commit",
		Start:     time.Now().Add(-100 * time.Second),
		End:       time.Now(),
		SystemOut: "test output",
	})
	em.AddEvent(&FailedEvent{
		Event: Event{
			Name:      "centos9",
			Suite:     "compose",
			ClassName: "image-download",
			Start:     time.Now().Add(-10 * time.Second),
			End:       time.Now(),
			SystemOut: "test output",
		},
		Message: "something wrong",
		Content: "very wrong",
	})

	em.AddEvent(&SkippedEvent{
		Event: Event{
			Name:      "rhel-9.3",
			Suite:     "compose",
			ClassName: "commit",
			Start:     time.Now().Add(-10 * time.Second),
			End:       time.Now(),
			SystemOut: "test output",
		},
		Message: "everything is fine, as expected, commit already on disk",
	})
	time.Sleep(1 * time.Second)
	fmt.Printf("%v\n", em)
	junit := em.GetJUnit()
	assert.NotNil(t, junit)

	assert.Equal(t, "compose", junit.Name)
	assert.Equal(t, 1, junit.Time)
	assert.Len(t, junit.TestSuites, 1)

	suite0 := junit.TestSuites[0]
	assert.Equal(t, "compose", suite0.Name)
	assert.Equal(t, 1, suite0.Time)
	assert.Equal(t, 3, suite0.Tests)
	assert.Equal(t, 1, suite0.Failures)
	assert.Equal(t, 1, suite0.Skipped)

	test0 := suite0.TestCases[0]
	assert.Equal(t, "rhel-9.2", test0.Name)
	assert.Equal(t, "commit", test0.ClassName)

	test1 := suite0.TestCases[1]
	assert.Equal(t, "centos9", test1.Name)
	assert.Equal(t, "image-download", test1.ClassName)

	test2 := suite0.TestCases[2]
	assert.Equal(t, "rhel-9.3", test2.Name)
	assert.Equal(t, "commit", test2.ClassName)

	_, err := junit.Marshal()
	assert.NoError(t, err)
}
