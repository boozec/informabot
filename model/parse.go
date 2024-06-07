package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"golang.org/x/exp/slices"

	"github.com/csunibo/config-parser-go"
	"github.com/csunibo/informabot/utils"
)

const (
	jsonPath      = "./json/"
	groupsFile    = "groups.json"
	configSubpath = "config/"
)

func ParseAutoReplies() (autoReplies []AutoReply, err error) {
	file, err := os.Open("./json/autoreply.json")
	if err != nil {
		return nil, fmt.Errorf("error reading autoreply.json file: %w", err)
	}

	err = json.NewDecoder(file).Decode(&autoReplies)
	if err != nil {
		return nil, fmt.Errorf("error parsing autoreply.json file: %w", err)
	}

	return
}

const COMMAND_MAX_LENGTH = 32

func commandNameFromString(s string) string {
	s = utils.ToSnakeCase(s)
	if len(s) > COMMAND_MAX_LENGTH {
		return s[:COMMAND_MAX_LENGTH]
	}
	return s
}

func commandNameFromTeaching(t cparser.Teaching) string {
	return commandNameFromString(t.Url)
}

func commandNameFromDegree(d cparser.Degree) string {
	return commandNameFromString(d.Id)
}

func ParseTeachings() (teachings map[string]cparser.Teaching, err error) {

	teachingsArray, err := cparser.ParseTeachings(filepath.Join(jsonPath, configSubpath))
	if err != nil {
		return nil, fmt.Errorf("error parsing Teachings: %w", err)
	}
	teachings = make(map[string]cparser.Teaching, len(teachingsArray))
	for _, t := range teachingsArray {
		teachings[commandNameFromTeaching(t)] = t
	}
	return
}

func ParseDegrees() (degrees map[string]cparser.Degree, err error) {
	degreesArray, err := cparser.ParseDegrees(filepath.Join(jsonPath, configSubpath))
	if err != nil {
		return nil, fmt.Errorf("error parsing Degrees: %w", err)
	}
	for i := range degreesArray {
		for j := range degreesArray[i].Teachings {
			degreesArray[i].Teachings[j].Name = commandNameFromString(degreesArray[i].Teachings[j].Name)
		}
	}
	degrees = make(map[string]cparser.Degree, len(degreesArray))
	for _, d := range degreesArray {
		degrees[commandNameFromDegree(d)] = d
	}
	return
}

func ParseSettings() (settings SettingsStruct, err error) {
	file, err := os.Open("./json/settings.json")
	if err != nil {
		return SettingsStruct{}, fmt.Errorf("error reading settings.json file: %w", err)
	}

	err = json.NewDecoder(file).Decode(&settings)
	if err != nil {
		return SettingsStruct{}, fmt.Errorf("error parsing settings.json file: %w", err)
	}

	err = file.Close()
	if err != nil {
		return SettingsStruct{}, fmt.Errorf("error closing settings.json file: %w", err)
	}

	return
}

func ParseActions() (actions []Action, err error) {
	byteValue, err := os.Open("./json/actions.json")
	if err != nil {
		return nil, fmt.Errorf("error reading actions.json file: %w", err)
	}

	actions, err = ParseActionsBytes(byteValue)
	if err != nil {
		return nil, fmt.Errorf("error parsing actions.json file: %w", err)
	}

	return
}

func ParseActionsBytes(reader io.Reader) (actions []Action, err error) {
	var mapData map[string]interface{}

	err = json.NewDecoder(reader).Decode(&mapData)
	if err != nil {
		return
	}

	for key, value := range mapData {
		action := GetActionFromType(key, value.(map[string]interface{})["type"].(string))
		err = mapstructure.Decode(value, &action)
		if err != nil {
			return
		}

		actions = append(actions, action)
	}

	slices.SortFunc(actions, func(a, b Action) int { return strings.Compare(a.Name, b.Name) })
	return
}

func ParseMemeList() (memes []Meme, err error) {
	byteValue, err := os.Open("./json/memes.json")
	if err != nil {
		return nil, fmt.Errorf("error reading memes.json file: %w", err)
	}

	var mapData map[string]string
	err = json.NewDecoder(byteValue).Decode(&mapData)
	if err != nil {
		return nil, fmt.Errorf("error parsing memes.json file: %w", err)
	}

	for key, value := range mapData {
		meme := Meme{Name: key, Text: value}
		memes = append(memes, meme)
	}

	return
}

func ParseOrCreateGroups() (GroupsStruct, error) {
	groups := make(GroupsStruct)

	filepath := filepath.Join(jsonPath, groupsFile)
	byteValue, err := os.ReadFile(filepath)
	if errors.Is(err, os.ErrNotExist) {
		return groups, nil
	} else if err != nil {
		return nil, fmt.Errorf("error reading %s file: %w", filepath, err)
	}

	err = json.Unmarshal(byteValue, &groups)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s file: %w", filepath, err)
	}

	if groups == nil {
		groups = make(GroupsStruct)
	}

	return groups, nil
}

func SaveGroups(groups GroupsStruct) error {
	filepath := filepath.Join(jsonPath, groupsFile)
	return utils.WriteJSONFile(filepath, groups)
}

func ParseTimetables() (timetables map[string]cparser.Timetable, err error) {
	timetables, err = cparser.ParseTimetables(filepath.Join(jsonPath, configSubpath))

	if err != nil {
		return nil, fmt.Errorf("error parsing Timetables: %w", err)
	}
	return
}

func ParseMaintainers() (maintainer []cparser.Maintainer, err error) {
	return cparser.ParseMaintainers(path.Join(jsonPath, configSubpath))
}

func ParseRepresentatives() (map[string]cparser.Representative, error) {
	return cparser.ParseRepresentatives(path.Join(jsonPath, configSubpath))
}
