package config

func Init(name, configFile string) error {
	return ViperInit(name, configFile)
}

func Unmarshal(k string, v interface{}) error {
	return ViperUnmarshal(k, v)
}
