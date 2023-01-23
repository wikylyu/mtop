package config

func Init(prefix string) error {
	return ViperInit(prefix)
}

func Unmarshal(k string, v interface{}) error {
	return ViperUnmarshal(k, v)
}
