package lights

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/jurgen-kluft/go-conbee/conbee"
)

var (
	getAllLightsURL  = "http://%s/api/%s/lights"
	getLightStateURL = "http://%s/api/%s/lights/%d"
	setLightStateURL = "http://%s/api/%s/lights/%d/state"
	setLightAttrsURL = "http://%s/api/%s/lights/%d"
)

type Lights struct {
	Hostname string
	APIkey   string

	// TODO: Do we put an array of lights in this instance?
	//
}

type Light struct {
	Name         string `json:"name"`
	ID           int    `json:"id,omitempty"`
	ETag         string `json:"etag,omitempty"`
	State        State  `json:"state,omitempty"`
	HasColor     bool   `json:"hascolor,omitempty"`
	Type         string `json:"type,omitempty"`
	Manufacturer string `json:"manufacturer,omitempty"`
	ModelID      string `json:"modelid,omitempty"`
	UniqueID     string `json:"uniqueid,omitempty"`
	SWVersion    string `json:"swversion,omitempty"`
}

type State struct {
	On             *bool     `json:"on,omitempty"`     //
	Hue            *uint16   `json:"hue,omitempty"`    //
	Effect         string    `json:"effect,omitempty"` //
	Bri            *uint8    `json:"bri,omitempty"`    // min = 1, max = 254
	Sat            *uint8    `json:"sat,omitempty"`    //
	CT             *uint16   `json:"ct,omitempty"`     // min = 154, max = 500
	XY             []float32 `json:"xy,omitempty"`
	Alert          string    `json:"alert,omitempty"`
	Reachable      *bool     `json:"reachable,omitempty"`
	ColorMode      string    `json:"colormode,omitempty"`
	ColorLoopSpeed *uint8    `json:"colorloopspeed,omitempty"`
	TransitionTime *uint16   `json:"transitiontime,omitempty"`
}

func New(hostname string, apikey string) *Lights {
	return &Lights{
		Hostname: hostname,
		APIkey:   apikey,
	}
}

func (l *Lights) GetLightState(lightID int) (Light, error) {
	var ll Light
	url := fmt.Sprintf(getLightStateURL, l.Hostname, l.APIkey, lightID)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ll, err
	}
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return ll, err
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return ll, err
	}
	err = json.Unmarshal(contents, &ll)
	if err != nil {
		return ll, err
	}
	ll.ID = lightID
	return ll, err
}

func (l *Lights) SetLightAttrs(lightID int, lightName string) ([]conbee.ApiResponse, error) {
	url := fmt.Sprintf(setLightAttrsURL, l.Hostname, l.APIkey, lightID)
	data := fmt.Sprintf("{\"name\": \"%s\"}", lightName)
	postbody := strings.NewReader(data)
	request, err := http.NewRequest("PUT", url, postbody)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var apiResponse []conbee.ApiResponse
	err = json.Unmarshal(contents, &apiResponse)
	if err != nil {
		return nil, err
	}
	return apiResponse, err
}

func (s *State) SetOn(OnOff bool) {
	s.On = new(bool)
	*s.On = OnOff
}

func (s *State) SetCT(Bri int, CT int) {
	s.Bri = new(uint8)
	*s.Bri = uint8(Bri)
	s.CT = new(uint16)
	*s.CT = uint16(CT)
}

func (s *State) SetXY(x, y float32) {
	s.XY = make([]float32, 2, 2)
	s.XY[0] = x
	s.XY[1] = y
}

func (l *Lights) SetLightState(lightID int, state *State) ([]conbee.ApiResponse, error) {
	url := fmt.Sprintf(setLightStateURL, l.Hostname, l.APIkey, lightID)
	stateJSON, err := json.Marshal(&state)
	if err != nil {
		return nil, err
	}
	postbody := strings.NewReader(string(stateJSON))
	request, err := http.NewRequest("PUT", url, postbody)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var apiResponse []conbee.ApiResponse
	err = json.Unmarshal(contents, &apiResponse)
	if err != nil {
		return nil, err
	}
	return apiResponse, err
}

func (l *Lights) GetAllLights() ([]Light, error) {
	url := fmt.Sprintf(getAllLightsURL, l.Hostname, l.APIkey)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	lightsMap := map[string]Light{}
	err = json.Unmarshal(contents, &lightsMap)
	if err != nil {
		return nil, err
	}
	lights := make([]Light, 0, len(lightsMap))
	for lightID, light := range lightsMap {
		light.ID, _ = strconv.Atoi(lightID)
		lights = append(lights, light)
	}

	sort.Slice(lights, func(i, j int) bool { return lights[i].ID < lights[j].ID })
	return lights, err
}

func (l *Light) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("ID:              %d\n", l.ID))
	buffer.WriteString(fmt.Sprintf("UUID:            %s\n", l.UniqueID))
	buffer.WriteString(fmt.Sprintf("Name:            %s\n", l.Name))
	buffer.WriteString(fmt.Sprintf("Type:            %s\n", l.Type))
	buffer.WriteString(fmt.Sprintf("ModelId:         %s\n", l.ModelID))
	buffer.WriteString(fmt.Sprintf("SwVersion:       %s\n", l.SWVersion))
	buffer.WriteString(fmt.Sprintf("State:\n"))
	buffer.WriteString(l.State.String())
	return buffer.String()
}

func (s *State) String() string {
	var buffer bytes.Buffer
	if s.On != nil {
		buffer.WriteString(fmt.Sprintf("On:              %t\n", *s.On))
	}
	if s.Hue != nil {
		buffer.WriteString(fmt.Sprintf("Hue:             %d\n", *s.Hue))
	}
	if len(s.Effect) > 0 {
		buffer.WriteString(fmt.Sprintf("Effect:          %s\n", s.Effect))
	}
	if s.Bri != nil {
		buffer.WriteString(fmt.Sprintf("Bri:             %d\n", *s.Bri))
	}
	if s.Sat != nil {
		buffer.WriteString(fmt.Sprintf("Sat:             %d\n", *s.Sat))
	}
	if s.CT != nil && *s.CT > 0 {
		buffer.WriteString(fmt.Sprintf("CT:              %d\n", *s.CT))
	}
	if len(s.XY) > 0 {
		buffer.WriteString(fmt.Sprintf("XY:              %g, %g\n", s.XY[0], s.XY[1]))
	}
	if len(s.Alert) > 0 {
		buffer.WriteString(fmt.Sprintf("Alert:           %s\n", s.Alert))
	}
	if s.TransitionTime != nil && *s.TransitionTime > 0 {
		buffer.WriteString(fmt.Sprintf("TransitionTime:  %d\n", *s.TransitionTime))
	}
	if s.Reachable != nil {
		buffer.WriteString(fmt.Sprintf("Reachable:       %t\n", *s.Reachable))
	}
	if len(s.ColorMode) > 0 {
		buffer.WriteString(fmt.Sprintf("ColorMode:       %s\n", s.ColorMode))
	}
	if s.ColorLoopSpeed != nil && *s.ColorLoopSpeed > 0 {
		buffer.WriteString(fmt.Sprintf("ColorLoopSpeed:  %d\n", *s.ColorLoopSpeed))
	}
	return buffer.String()
}
