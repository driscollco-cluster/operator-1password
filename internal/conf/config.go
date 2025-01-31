package conf

var Config config

type config struct {
	OnePassword struct {
		Api struct {
			Url   string
			Token string
		}
	}
	Secrets struct {
		Refresh struct {
			MinIntervalSeconds int
		}
	}
}
