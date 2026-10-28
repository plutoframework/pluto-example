package config

type CacheConfig struct {
	Backend       string
	Environment   string
	URL           string
	AllowInsecure bool
}
