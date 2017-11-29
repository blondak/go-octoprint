package octoprint

import (
	"encoding/json"
	"strings"
	"time"
)

type Axis string

const (
	XAxis Axis = "x"
	YAxis Axis = "y"
	ZAxis Axis = "z"
)

// FullStateResponse contains informantion about the current state of the printer.
type FullStateResponse struct {
	//Temperature is the printer’s temperature state data.
	Temperature TemperatureState `json:"temperature"`
	// SD is the printer’s sd state data.
	SD SDState `json:"sd"`
	// State is the printer’s general state.
	State PrinterState `json:"state"`
}

// JobResponse is the response from a job command.
type JobResponse struct {
	// Job contains information regarding the target of the current print job.
	Job JobInformation `json:"job"`
	// Progress contains information regarding the progress of the current job.
	Progress ProgressInformation `json:"progress"`
}

// JobInformation contains information regarding the target of the current job.
type JobInformation struct {
	// File is the file that is the target of the current print job.
	File FileInformation `json:"file"`
	// EstimatedPrintTime is the estimated print time for the file, in seconds.
	EstimatedPrintTime int `json:"estimatedPrintTime"`
	// LastPrintTime is the print time of the last print of the file, in seconds.
	LastPrintTime int `json:"lastPrintTime"`
	// Filament contains Information regarding the estimated filament
	// usage of the print job.
	Filament struct {
		// Length of filament used, in mm
		Length int `json:"length"`
		// Volume of filament used, in cm³
		Volume float64 `json:"volume"`
	} `json:"filament"`
	FilePosition int `json:"filepos"`
}

// FileInformation contains information regarding a file.
type FileInformation struct {
	// Name is name of the file without path. E.g. “file.gco” for a file
	// “file.gco” located anywhere in the file system.
	Name string `json:"name"`
	// Path is the path to the file within the location. E.g.
	//“folder/subfolder/file.gco” for a file “file.gco” located within “folder”
	// and “subfolder” relative to the root of the location.
	Path string `json:"path"`
	// Type of file. model or machinecode. Or folder if it’s a folder, in
	// which case the children node will be populated.
	Type string `json:"type"`
	// TypePath path to type of file in extension tree. E.g. `["model", "stl"]`
	// for .stl files, or `["machinecode", "gcode"]` for .gcode files.
	// `["folder"]` for folders.
	TypePath string `json:"typePath"`
}

// ProgressInformation contains information regarding the progress of the
// current print job.
type ProgressInformation struct {
	// Completion percentage of completion of the current print job.
	Completion float64 `json:"completion"`
	// FilePosition current position in the file being printed, in bytes
	// from the beginning.
	FilePosition int `json:"filepos"`
	// PrintTime is time already spent printing, in seconds
	PrintTime int `json:"printTime"`
	// PrintTimeLeft is estimate of time left to print, in seconds
	PrintTimeLeft int `json:"printTimeLeft"`
}

// TemperatureState is the printer’s temperature state data.
type TemperatureState temperatureState
type temperatureState struct {
	// Current temperature stats.
	Current map[string]TemperatureData `json:"current"`
	// Temperature history.
	History []*HistoricTemperatureData `json:"history"`
}

func (r *TemperatureState) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	history := raw["history"]
	delete(raw, "history")
	b, _ = json.Marshal(map[string]interface{}{
		"current": raw,
		"history": history,
	})

	i := &temperatureState{}
	if err := json.Unmarshal(b, i); err != nil {
		return err
	}

	*r = TemperatureState(*i)
	return nil
}

// TemperatureData is temperature stats for a tool.
type TemperatureData struct {
	// Actual current temperature.
	Actual float64 `json:"actual"`
	// Target temperature, may be nil if no target temperature is set.
	Target float64 `json:"target"`
	// Offset currently configured temperature offset to apply, will be left
	// out for historic temperature information.
	Offset float64 `json:"offset"`
}

// PrinterState current state of the printer.
type PrinterState struct {
	Text  string `json:"text"`
	Flags struct {
		Operations    bool `json:"operational"`
		Puased        bool `json:"paused"`
		Printing      bool `json:"printing"`
		SDReady       bool `json:"sdReady"`
		Error         bool `json:"error"`
		Ready         bool `json:"ready"`
		ClosedOnError bool `json:"closedOrError"`
	} `json:"flags"`
}

// SDState is the state of the sd reader.
type SDState struct {
	Ready bool `json:"ready"`
}

// HistoricTemperatureData is temperature historic stats for a tool.
type HistoricTemperatureData historicTemperatureData
type historicTemperatureData struct {
	// Time of this data point.
	Time time.Time `json:"time"`
	// Tools is temperature stats a set of tools.
	Tools map[string]TemperatureData `json:"tools"`
}

func (h *HistoricTemperatureData) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	ts := raw["time"]
	delete(raw, "time")
	b, _ = json.Marshal(map[string]interface{}{
		"time":  time.Unix(int64(ts.(float64)), 0),
		"tools": raw,
	})

	i := &historicTemperatureData{}
	if err := json.Unmarshal(b, i); err != nil {
		return err
	}

	*h = HistoricTemperatureData(*i)
	return nil
}

// VersionResponse is the response from a job command.
type VersionResponse struct {
	// API is the API version.
	API string `json:"api"`
	// Server is the server version.
	Server string `json:"server"`
}

type ConnectionState string

const (
	Operational ConnectionState = "Operational"
)

// The states are  based on:
// https://github.com/foosel/OctoPrint/blob/77753ca02602d3a798d6b0a22535e6fd69ff448a/src/octoprint/util/comm.py#L549
// no comments :(

func (s ConnectionState) IsOperational() bool {
	return strings.HasPrefix(string(s), "Operational") ||
		strings.HasPrefix(string(s), "Transfering") ||
		strings.HasPrefix(string(s), "Paused")
}

func (s ConnectionState) IsPrinting() bool {
	return strings.HasPrefix(string(s), "Printing") ||
		strings.HasPrefix(string(s), "Sending") ||
		strings.HasPrefix(string(s), "Paused")
}

func (s ConnectionState) IsOffline() bool {
	return strings.HasPrefix(string(s), "Offline") ||
		strings.HasPrefix(string(s), "Closed")
}

func (s ConnectionState) IsError() bool {
	return strings.HasPrefix(string(s), "Error") ||
		strings.HasPrefix(string(s), "Unknown")
}

func (s ConnectionState) IsConnecting() bool {
	return strings.HasPrefix(string(s), "Opening") ||
		strings.HasPrefix(string(s), "Detecting") ||
		strings.HasPrefix(string(s), "Connecting") ||
		strings.HasPrefix(string(s), "Detecting")
}

// ConnectionResponse is the response from a connection command.
type ConnectionResponse struct {
	Current struct {
		// State current state of the connection.
		State ConnectionState `json:"state"`
		// Port to connect to.
		Port string `json:"port"`
		// BaudRate speed of the connection.
		BaudRate int `json:"baudrate"`
		// PrinterProfile profile to use for connection.
		PrinterProfile string `json:"printerProfile"`
	}
	Options struct {
		// Ports list of available ports.
		Ports []string `json:"ports"`
		// BaudRates list of available speeds.
		BaudRates []int `json:"baudrates"`
		// PrinterProfile list of available profiles.
		PrinterProfiles []*Profile `json:"printerProfiles"`
		// PortPreference default port.
		PortPreference string `json:"portPreference"`
		// BaudRatePreference default speed.
		BaudRatePreference int `json:"baudratePreference"`
		// PrinterProfilePreference default profile.
		PrinterProfilePreference string `json:"printerProfilePreference"`
		// Autoconnect whether to automatically connect to the printer on
		// OctoPrint’s startup in the future.
		Autoconnect bool `json:"autoconnect"`
	}
}

// Profile describe a printer profile.
type Profile struct {
	// ID is the identifier of the profile.
	ID string `json:"id"`
	// Name is the display name of the profile.
	Name string `json:"name"`
}
