package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dchauviere/go-tydom/internal/config"
	"github.com/dchauviere/go-tydom/pkg/tydom"
)

type WidgetBehaviour struct{}

//nolint:tagliatelle
type Endpoint struct {
	EndpointID        int    `json:"id_endpoint"`
	FirstUsage        string `json:"first_usage"`
	Skill             string `json:"skill,omitempty"`
	DeviceID          int    `json:"id_device"`
	Name              string `json:"name"`
	AnticipationStart bool   `json:"anticipation_start"`
	Picto             string `json:"picto"`
	LastUsage         string `json:"last_usage"`
	//WidgetBehaviour   WidgetBehaviour `json:"widget_behaviour,omitempty"`
}

type Moment struct {
	Color int    `json:"color"`
	Name  string `json:"name"`
	ID    int    `json:"id"`
}

//nolint:tagliatelle
type Group struct {
	GroupAll    bool   `json:"group_all"`
	Usage       string `json:"usage"`
	Name        string `json:"name"`
	ID          int    `json:"id"`
	IsGroupUser bool   `json:"is_group_user"`
	Picto       string `json:"picto"`
}

type Area struct{}

//nolint:tagliatelle
type Scenario struct {
	RuleID string `json:"rule_id"`
	Name   string `json:"name"`
	ID     int    `json:"id"`
	Type   string `json:"type"`
	Picto  string `json:"picto"`
}

type ZigbeeNetwork struct{}

//nolint:tagliatelle
type UserConfig struct {
	Date               int             `json:"date"`
	VersionApplication string          `json:"version_application"`
	Endpoints          []Endpoint      `json:"endpoints"`
	OldTycam           bool            `json:"old_tycam"`
	NewTycam           bool            `json:"new_tycam"`
	Moments            []Moment        `json:"moments"`
	Os                 string          `json:"os"`
	Groups             []Group         `json:"groups"`
	Areas              []Area          `json:"areas"`
	Scenarios          []Scenario      `json:"scenarios"`
	CatalogID          string          `json:"id_catalog"`
	Version            string          `json:"version"`
	ZigbeeNetworks     []ZigbeeNetwork `json:"zigbee_networks,omitempty"`
	CameraInstallDate  int             `json:"camera_install_date"`
}

/*
{
	"date":1725205332,
	"version_application":"4.13.11-1-dd",
	"endpoints":[
		{
			"id_endpoint":0,
			"first_usage":"shutter",
			"skill":"TYDOM_X3D",
			"id_device":0,
			"name":"Fenetre Salon",
			"anticipation_start":false,
			"picto":"picto_shutter",
			"last_usage":"shutter",
			"widget_behavior":{
				"tutorial_id":"Volet_roulant_wellcom"
			}
		},
		{
			"id_endpoint":0,
			"first_usage":"hvac",
			"skill":"TYDOM_X3D",
			"id_device":1,
			"name":"Thermique 1",
			"anticipation_start":false,
			"picto":"picto_thermometer",
			"last_usage":"boiler",
			"widget_behavior":{
				"tutorial_id":"6_RecepteurRF_serie6000"
			}
		},
		{
			"id_endpoint":0,
			"first_usage":"shutter",
			"skill":"TYDOM_X3D",
			"id_device":3,
			"name":"Etage",
			"anticipation_start":false,
			"picto":"picto_shutter",
			"last_usage":"shutter",
			"widget_behavior":{
				"tutorial_id":"7_Tyxia_serie4000"
			}
		}
	],
	"old_tycam":false,
	"moments":[
		{
			"color":9813268,
			"name":"fin chauffage",
			"id":401341867
		},
		{
			"color":9813268,
			"name":"chauffage",
			"id":1446945519
		}
	],
	"os":"android",
	"groups":[
		{
			"group_all":true,
			"usage":"light",
			"name":"TOTAL",
			"id":623866694,
			"is_group_user":false,
			"picto":"picto_lamp"
		},
		{
			"group_all":true,
			"usage":"shutter",
			"name":"TOTAL",
			"id":1635339958,
			"is_group_user":false,
			"picto":"picto_shutter"
		},
		{
			"group_all":true,
			"usage":"awning",
			"name":"TOTAL",
			"id":1532324573,
			"is_group_user":false,
			"picto":"picto_awning_awning"
		},
		{
			"group_all":true,
			"usage":"plug",
			"name":"Total",
			"id":1734452070,
			"is_group_user":false,
			"picto":"default_device"
		}
	],
	"areas":[],
	"scenarios":[
		{
			"rule_id":"",
			"name":"test",
			"id":725442656,
			"type":"NORMAL",
			"picto":"picto_scenario_clap"
		}
	],
	"id_catalog":"F2BD90F93B888DA02C54980F11AE4796DFCC98F447CD3FE326F5A3A964C939BF",
	"version":"1.0.2",
	"zigbee_networks":[]
}
*/

func parseUserConfig(resp *http.Response) (*UserConfig, error) {
	var data UserConfig

	err := parse(resp, &data)
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to parse user config response"}
	}

	return &data, nil
}

// Get user config.
func (ac *ClientAPI) GetUserConfig() (*UserConfig, error) {
	resp, err := ac.Client.HTTPRequest("GET", "/configs/file", "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send user configs request"}
	}

	if resp.StatusCode == http.StatusNotFound {
		return &UserConfig{
			Date:               int(time.Now().Unix()),
			VersionApplication: config.VERSION,
			Endpoints:          []Endpoint{},
			OldTycam:           false,
			Moments:            []Moment{},
			Areas:              []Area{},
			Os:                 "Linux",
			Groups:             []Group{},
			Scenarios:          []Scenario{},
			CatalogID:          "",
			Version:            "1.0.1",
			ZigbeeNetworks:     []ZigbeeNetwork{},
		}, nil
	}

	return parseUserConfig(resp)
}

func (asc *ClientAsyncAPI) GetUserConfig(hook func(data *UserConfig)) error {
	err := asc.Client.HTTPAsyncRequest("GET", "/configs/file", "", func(resp *http.Response) {
		data, err := parseUserConfig(resp)
		if err != nil {
			asc.Client.Logger.Error("bad user configs response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request user configs"}
	}

	return nil
}

// Set user config.
func (ac *ClientAPI) SetUserConfig(userConfig *UserConfig) (*UserConfig, error) {
	body, err := json.Marshal(userConfig)
	if err != nil {
		ac.Client.Logger.Error("failed to encode configs")

		return nil, &tydom.Error{Msg: "failed to encode configs", Err: err}
	}

	resp, err := ac.Client.HTTPRequest("POST", "/configs/file", string(body))
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send user configs set"}
	}

	return parseUserConfig(resp)
}

func (asc *ClientAsyncAPI) SetUserConfig(userConfig *UserConfig, hook func(data *UserConfig)) error {
	body, err := json.Marshal(userConfig)
	if err != nil {
		asc.Client.Logger.Error("failed to encode configs")

		return &tydom.Error{Msg: "failed to encode configs", Err: err}
	}

	err = asc.Client.HTTPAsyncRequest("PUT", "/configs/file", string(body), func(resp *http.Response) {
		data, err := parseUserConfig(resp)
		if err != nil {
			asc.Client.Logger.Error("bad set user configs response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request user configs"}
	}

	return nil
}

// Get moments config.
func (ac *ClientAPI) GetMomentsConfig() (*any, error) {
	resp, err := ac.Client.HTTPRequest("GET", "/moments/file", "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send moments request"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) GetMomentsConfig(hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest("GET", "/moments/file", "", func(resp *http.Response) {
		data, err := parseAny(resp)
		if err != nil {
			asc.Client.Logger.Error("bad moments response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request moments"}
	}

	return nil
}

// Get scenarios config.
func (ac *ClientAPI) GetScenariosConfig() (*any, error) {
	resp, err := ac.Client.HTTPRequest("GET", "/scenarios/file", "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send scenarios request"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) GetScenariosConfig(hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest("GET", "/scenarios/file", "", func(resp *http.Response) {
		data, err := parseAny(resp)
		if err != nil {
			asc.Client.Logger.Error("bad scenarios response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request scenarios"}
	}

	return nil
}

// Get local_claim config.
func (ac *ClientAPI) GetLocalClaimConfig() (*any, error) {
	resp, err := ac.Client.HTTPRequest("GET", "/configs/gateway/local_claim", "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send local_claim request"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) GetLocalClaimConfig(hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest("GET", "/configs/gateway/local_claim", "", func(resp *http.Response) {
		data, err := parseAny(resp)
		if err != nil {
			asc.Client.Logger.Error("bad local_claim response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request local_claim"}
	}

	return nil
}

// Get geoloc config.
func (ac *ClientAPI) GetGeoLocConfig() (*any, error) {
	resp, err := ac.Client.HTTPRequest("GET", "/configs/gateway/geoloc", "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send geoloc request"}
	}

	return parseAny(resp)
}

func (asc *ClientAsyncAPI) GetGeoLocConfig(hook func(data *any)) error {
	err := asc.Client.HTTPAsyncRequest("GET", "/configs/gateway/geoloc", "", func(resp *http.Response) {
		data, err := parseAny(resp)
		if err != nil {
			asc.Client.Logger.Error("bad geoloc response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request geoloc"}
	}

	return nil
}
