package config

var (
	DefaultReverseProxy = ReverseProxy{
		Port: 9200,
		Aws:  DefaultsAws,
	}
	EnvReverseProxy = ReverseProxy{
		Aws: EnvAws,
	}
)

type ReverseProxy struct {
	Port int `hcl:"port"`
	Aws  Aws `hcl:"aws"`
}

func MergeReverseProxy(cfgs ...ReverseProxy) ReverseProxy {
	rv := ReverseProxy{}
	for _, cur := range cfgs {
		if cur.Port > 0 {
			rv.Port = cur.Port
		}
		rv.Aws = MergeAws(rv.Aws, cur.Aws)
	}
	return rv
}
