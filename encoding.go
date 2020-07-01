package utilz

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

// TranscodeJSON converts the input to json, and then unmarshals it into the
// destination. The destination must be a pointer.
func TranscodeJSON(input interface{}, destinationPointer interface{}) error {
	b, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("error while marshaling input to json: %s", err)
	}

	err = json.Unmarshal(b, destinationPointer)
	if err != nil {
		return fmt.Errorf("error while unmarshaling json to destination: %s", err)
	}
	return nil
}

// TranscodeYAML converts the input to yaml, and then unmarshals it into the
// destination. The destination must be a pointer.
func TranscodeYAML(input interface{}, destinationPointer interface{}) error {
	b, err := yaml.Marshal(input)
	if err != nil {
		return fmt.Errorf("error while marshaling input to yaml: %s", err)
	}

	err = yaml.Unmarshal(b, destinationPointer)
	if err != nil {
		return fmt.Errorf("error while unmarshaling yaml to destination: %s", err)
	}
	return nil
}

func LoadYaml(ptr interface{}, filepath string) error {
	yamlFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("error while reading config file from %q: %s", filepath, err)
	}
	err = yaml.Unmarshal(yamlFile, ptr)
	if err != nil {
		return fmt.Errorf("error while unmarshaling config file: %s", err)
	}
	return nil
}

func SaveAsYaml(v interface{}, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error while OpenFile: %s", err)
	}
	defer file.Close()

	err = file.Truncate(0)
	if err != nil {
		return fmt.Errorf("error while file.Truncate: %s", err)
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error while file.Seek: %s", err)
	}

	enc := yaml.NewEncoder(file)
	err = enc.Encode(v)
	if err != nil {
		return fmt.Errorf("error while enc.Encode: %s", err)
	}
	return nil
}

func LoadJSON(ptr interface{}, filepath string) error {
	jsonFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("error while reading config file from %q: %s", filepath, err)
	}
	err = json.Unmarshal(jsonFile, ptr)
	if err != nil {
		return fmt.Errorf("error while unmarshaling config file: %s", err)
	}

	return nil
}

func SaveAsJSON(v interface{}, filepath string) error {
	d, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("error while marshaling config: %s", err)
	}

	err = ioutil.WriteFile(filepath, d, 0640)
	if err != nil {
		return fmt.Errorf("error while writing config file: %s", err)
	}

	return nil
}

func SaveAsIndentedJSON(v interface{}, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error while OpenFile: %s", err)
	}
	defer file.Close()

	err = file.Truncate(0)
	if err != nil {
		return fmt.Errorf("error while file.Truncate: %s", err)
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error while file.Seek: %s", err)
	}

	enc := json.NewEncoder(file)
	enc.SetIndent("", "   ")
	err = enc.Encode(v)
	if err != nil {
		return fmt.Errorf("error while enc.Encode: %s", err)
	}
	return nil
}

func Itoa(i int) string {
	return strconv.Itoa(i)
}

func Atoi(s string) (int, error) {
	return strconv.Atoi(s)
}

func MustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}
