package lib

import "regexp"

type Route struct {
	Name    string `yaml:"name"`
	Pattern string `yaml:"pattern"`
	URL     string `yaml:"url"`
}

func (r *Route) Match(topic string) (ok bool, err error) {
	return regexp.MatchString(r.Pattern, topic)
}
