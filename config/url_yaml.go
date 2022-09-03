package config

import "net/url"

type urlYaml struct {
	*url.URL
}

func (j *urlYaml) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}
	url, err := url.Parse(s)
	j.URL = url
	return err
}
