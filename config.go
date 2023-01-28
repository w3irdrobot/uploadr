package main

import (
	"fmt"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	flag "github.com/spf13/pflag"
)

func getConfiguration(f *flag.FlagSet) (*koanf.Koanf, error) {
	k := koanf.New(".")

	cFiles, _ := f.GetStringSlice("conf")
	for _, c := range cFiles {
		if err := k.Load(file.Provider(c), toml.Parser()); err != nil {
			return nil, fmt.Errorf("error loading file: %w", err)
		}
	}

	if err := k.Load(posflag.Provider(f, ".", k), nil); err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	return k, nil
}
