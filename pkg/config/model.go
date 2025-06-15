package config

import (
	"errors"
	"net/url"
	"time"

	"gopkg.in/yaml.v3"
)

type Duration struct {
	time.Duration
}

func NewDuration(val time.Duration) Duration {
	return Duration{val}
}

func (d *Duration) UnmarshalYAML(n *yaml.Node) error {
	var v any
	if err := n.Decode(&v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		return err
	default:
		return errors.New("invalid duration")
	}
}

type URL struct {
	*url.URL
}

func NewURL(val *url.URL) URL {
	return URL{val}
}

func (u *URL) UnmarshalYAML(n *yaml.Node) error {
	var v any
	if err := n.Decode(&v); err != nil {
		return err
	}

	switch value := v.(type) {
	case string:
		var err error
		u.URL, err = url.Parse(value)
		return err
	default:
		return errors.New("invalid duration")
	}
}
