package conf

var Config config

type config struct {
	GCP struct {
		ProjectId string `src:"image_artifactRegistry_project" required:"true"`
	}
}
