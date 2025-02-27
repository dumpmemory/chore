package config

import (
	"context"
	"fmt"
	"net"
)

const ParameterMac = "mac"

type paramMac struct {
	baseParameter
}

func (p paramMac) Type() string {
	return ParameterMac
}

func (p paramMac) String() string {
	return fmt.Sprintf("%q (required=%t)", p.description, p.required)
}

func (p paramMac) Validate(_ context.Context, value string) error {
	if _, err := net.ParseMAC(value); err != nil {
		return fmt.Errorf("incorrect mac address: %w", err)
	}

	return nil
}

func NewMac(description string, required bool, _ map[string]string) (Parameter, error) {
	return paramMac{
		baseParameter: baseParameter{
			required:    required,
			description: description,
		},
	}, nil
}
