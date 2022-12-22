package config

import (
	"fmt"
	"io"

	"github.com/9seconds/chore/internal/vcs"
)

type Config struct {
	Description string
	Git         vcs.GitAccessMode
	Network     bool
	AsUser      string
	Parameters  map[string]Parameter
	Flags       map[string]bool
}

func Parse(reader io.Reader) (Config, error) { //nolint: cyclop
	raw, err := parseRaw(reader)
	if err != nil {
		return Config{}, err
	}

	gitMode, err := vcs.GetGitAccessMode(raw.Git)
	if err != nil {
		return Config{}, fmt.Errorf("cannot parse git access mode: %w", err)
	}

	conf := Config{
		Description: raw.Description,
		Network:     raw.Network,
		AsUser:      raw.AsUser,
		Git:         gitMode,
		Parameters:  make(map[string]Parameter),
		Flags:       raw.Flags,
	}

	for name, param := range raw.Parameters {
		name := NormalizeName(name)

		var (
			value Parameter
			err   error
		)

		switch param.Type {
		case ParameterInteger:
			value, err = NewInteger(param.Required, param.Spec)
		case ParameterString:
			value, err = NewString(param.Required, param.Spec)
		case ParameterFloat:
			value, err = NewFloat(param.Required, param.Spec)
		case ParameterURL:
			value, err = NewURL(param.Required, param.Spec)
		case ParameterEmail:
			value, err = NewEmail(param.Required, param.Spec)
		case ParameterEnum:
			value, err = NewEnum(param.Required, param.Spec)
		case ParameterBase64:
			value, err = NewBase64(param.Required, param.Spec)
		case ParameterHex:
			value, err = NewHex(param.Required, param.Spec)
		case ParameterHostname:
			value, err = NewHostname(param.Required, param.Spec)
		case ParameterMac:
			value, err = NewMac(param.Required, param.Spec)
		case ParameterJSON:
			value, err = NewJSON(param.Required, param.Spec)
		case ParameterXML:
			value, err = NewXML(param.Required, param.Spec)
		case ParameterUUID:
			value, err = NewUUID(param.Required, param.Spec)
		case ParameterDirectory:
			value, err = NewDirectory(param.Required, param.Spec)
		case ParameterFile:
			value, err = NewFile(param.Required, param.Spec)
		case ParameterSemver:
			value, err = NewSemver(param.Required, param.Spec)
		case ParameterDatetime:
			value, err = NewDatetime(param.Required, param.Spec)
		default:
			return conf, fmt.Errorf("unknown parameter type %s for parameter %s", param.Type, name)
		}

		if err != nil {
			return conf, fmt.Errorf("cannot initialize parameter %s: %w", name, err)
		}

		conf.Parameters[name] = value
	}

	return conf, nil
}
