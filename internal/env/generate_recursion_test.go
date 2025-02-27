package env_test

import (
	"testing"

	"github.com/9seconds/chore/internal/argparse"
	"github.com/9seconds/chore/internal/env"
	"github.com/stretchr/testify/suite"
)

type GenerateRecursionTestSuite struct {
	EnvBaseTestSuite

	args argparse.ParsedArgs
}

func (suite *GenerateRecursionTestSuite) SetupTest() {
	suite.EnvBaseTestSuite.SetupTest()

	suite.args = argparse.ParsedArgs{
		Parameters: map[string]string{
			"param1": "33",
			"param2": "34 35",
		},
		Flags: map[string]argparse.FlagValue{
			"flag1": argparse.FlagTrue,
			"flag2": argparse.FlagFalse,
		},
		Positional: []string{"pos1", "pos2", "pos3"},
	}
}

func (suite *GenerateRecursionTestSuite) TestEnv() {
	env.GenerateRecursion(
		suite.Context(),
		suite.values,
		suite.wg,
		"namespace2",
		"script1",
		suite.args)

	data := suite.Collect()

	suite.Len(data, 1)
	suite.Contains(data[env.EnvRecursion], "run namespace2 script1")
	suite.Contains(data[env.EnvRecursion], "param1=33")
	suite.Contains(data[env.EnvRecursion], "'param2=34 35'")
	suite.Contains(data[env.EnvRecursion], "+flag1")
	suite.Contains(data[env.EnvRecursion], "-flag2")
	suite.NotContains(data[env.EnvRecursion], "pos1")
	suite.NotContains(data[env.EnvRecursion], "pos2")
	suite.NotContains(data[env.EnvRecursion], "pos3")
}

func TestGenerateRecursion(t *testing.T) {
	suite.Run(t, &GenerateRecursionTestSuite{})
}
