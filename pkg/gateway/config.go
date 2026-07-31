package gateway

type MQTTConfig struct {
	Enabled              bool
	URL                  string
	ClientID             string
	User                 string
	Password             string
	CleanSession         bool
	Store                string
	DiscoveryTopicPrefix string
	GatewayTopicPrefix   string
	CAFile               string
}

type TydomConfig struct {
	Hostname  string
	GatewayID string
	Password  string
}

type WebUITLSConfig struct {
	Enabled  bool
	CertFile string
	KeyFile  string
}

type WebUIConfig struct {
	Enabled  bool
	Addr     string
	Username string
	Password string
	TLS      *WebUITLSConfig
}

type Config struct {
	MQTT  *MQTTConfig
	Tydom *TydomConfig
	WebUI *WebUIConfig
}
