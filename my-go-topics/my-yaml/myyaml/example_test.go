package myyaml_test

import (
	"fmt"
	"os"
	"testing"

	"example.com/myyaml"

	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

func buildCfg(t *testing.T) *myyaml.Config {
	t.Helper()

	return &myyaml.Config{
		BaseConfig: myyaml.BaseConfig{
			Logger: myyaml.Logger{
				Source: true,
				Level:  "info",
			},
			Application: myyaml.Application{
				BuildInfo: myyaml.BuildInfo{
					Component: myyaml.Component{
						Version: "0.3.0",
					},
				},
			},
		},
	}
}

func TestSimple(t *testing.T) {

	t.Run("Config", func(t *testing.T) {
		filename := "config3.yaml"
		f, err := os.Create(filename)
		require.NoError(t, err)

		bytes, err := yaml.Marshal(buildCfg(t))
		require.NoError(t, err)

		fmt.Printf("%s", bytes)

		_, err = f.Write(bytes)
		require.NoError(t, err)

		defer f.Close()
		// defer os.Remove(filename)
	})

}
