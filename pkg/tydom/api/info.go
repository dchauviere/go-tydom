package api

import (
	"net/http"

	"github.com/dchauviere/go-tydom/pkg/tydom"
)

//nolint:tagliatelle
type Info struct {
	ProductName     string                `json:"productName"`
	Mac             string                `json:"mac"`
	Config          string                `json:"config"`
	BddEmpty        bool                  `json:"bddEmpty"`
	BddStatus       int                   `json:"bddStatus"`
	X3dEmpty        bool                  `json:"x3dEmpty"`
	APIMode         bool                  `json:"apiMode"`
	MainVersionSw   string                `json:"mainVersionSW"`
	MainVersionHw   string                `json:"mainVersionHW"`
	MainID          string                `json:"mainId"`
	MainReference   string                `json:"mainReference"`
	KeyVersionSw    string                `json:"keyVersionSW"`
	KeyVersionHw    string                `json:"keyVersionHW"`
	KeyVersionStack string                `json:"keyVersionStack"`
	KeyReference    string                `json:"keyReference"`
	BootReference   string                `json:"bootReference"`
	BootVersion     string                `json:"bootVersion"`
	TydomDat        int                   `json:"TYDOM.dat"`
	ConfigJSON      int                   `json:"config.json"`
	MomJSON         int                   `json:"mom.json"`
	GatewayDat      int                   `json:"gateway.dat"`
	BddJSON         int                   `json:"bdd.json"`
	CollectJSON     int                   `json:"collect.json"`
	GroupsJSON      int                   `json:"groups.json"`
	MomAPIJSON      int                   `json:"mom_api.json"`
	ScenarioJSON    int                   `json:"scenario.json"`
	SiteJSON        int                   `json:"site.json"`
	BddMigJSON      int                   `json:"bdd_mig.json"`
	InfoMigJSON     int                   `json:"info_mig.json"`
	InfoColJSON     int                   `json:"info_col.json"`
	AbsenceJSON     int                   `json:"absence.json"`
	AnticipJSON     int                   `json:"anticip.json"`
	TriggerJSON     int                   `json:"trigger.json"`
	URLMediation    string                `json:"urlMediation"`
	PltRegistered   bool                  `json:"pltRegistered"`
	UpdateAvailable bool                  `json:"updateAvailable"`
	PasswordEmpty   bool                  `json:"passwordEmpty"`
	Maintenance     map[string]string     `json:"maintenance"`
	Geoloc          infoGeoLoc            `json:"geoloc"`
	Clock           infoClock             `json:"clock"`
	Moments         map[string]infoMoment `json:"moments"`
	LocalClaim      infoLocalClaim        `json:"local_claim"`
	Protocols       []infoProtocol        `json:"protocols"`
}

type infoGeoLoc struct {
	Longitude int `json:"longitude"`
	Latitude  int `json:"latitude"`
}

type infoClock struct {
	Clock        string `json:"clock"`
	Source       string `json:"source"`
	Timezone     int    `json:"timezone"`
	SummerOffset string `json:"summerOffset"`
}

type infoMoment struct {
	To int `json:"to"`
}

type infoLocalClaim struct {
	Status     string `json:"status"`
	LastAccess string `json:"lastAccess"`
}

type infoProtocol struct {
	Protocol      string `json:"protocol"`
	Available     bool   `json:"available"`
	Installed     bool   `json:"installed,omitempty"`
	Ready         bool   `json:"ready,omitempty"`
	Status        string `json:"status,omitempty"`
	InstallStatus string `json:"installStatus,omitempty"`
}

func parseInfo(resp *http.Response) (*Info, error) {
	var data Info

	err := parse(resp, &data)
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to parse Info response", Err: err}
	}

	return &data, nil
}

// Get Tydom gateway info.
func (ac *ClientAPI) GetInfo() (*Info, error) {
	resp, err := ac.Client.HTTPRequest("GET", "/info", "")
	if err != nil {
		return nil, &tydom.Error{Msg: "failed to send Info request"}
	}

	return parseInfo(resp)
}

func (asc *ClientAsyncAPI) GetInfo(hook func(data *Info)) error {
	err := asc.Client.HTTPAsyncRequest("GET", "/info", "", func(resp *http.Response) {
		data, err := parseInfo(resp)
		if err != nil {
			asc.Client.Logger.Error("bad Info response")

			return
		}

		hook(data)
	})
	if err != nil {
		return &tydom.Error{Msg: "failed to send async request info"}
	}

	return nil
}
