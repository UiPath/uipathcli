package config

import "net/url"

type urlYaml struct {
	*url.URL
}

func (u urlYaml) MarshalYAML() (interface{}, error) {
	return u.URL.String(), nil
}

func (u *urlYaml) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	err := unmarshal(&s)
	if err != nil {
		return err
	}
	url, err := url.Parse(s)
	u.URL = url
	return err
}
