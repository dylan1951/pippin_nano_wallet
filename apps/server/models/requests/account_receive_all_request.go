package requests

type AccountReceiveAllRequest struct {
	BaseRequest `mapstructure:",squash"`
	Account     string `json:"account" mapstructure:"account"`
}
