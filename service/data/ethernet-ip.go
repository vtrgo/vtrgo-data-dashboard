// ethernet-ip.go
package data

import (
	"fmt"
	"log"
	"strconv"

	"vtarchitect/config"

	"github.com/danomagnum/gologix"
)

type PLC struct {
	client *gologix.Client
}

func NewPLC(ip string) *PLC {
	client := gologix.NewClient(ip)
	return &PLC{client: client}
}

func (plc *PLC) Connect() error {
	return plc.client.Connect()
}

func (plc *PLC) Disconnect() {
	plc.client.Disconnect()
}

func (plc *PLC) ReadTrigger(tagName string) (bool, error) {
	var tagValue bool
	err := plc.client.Read(tagName, &tagValue)
	return tagValue, err
}

func (plc *PLC) WriteResponse(tagName string, value bool) error {
	return plc.client.Write(tagName, value)
}

func (plc *PLC) ReadTagString(tagName string) (string, error) {
	var tagValue string
	err := plc.client.Read(tagName, &tagValue)
	return tagValue, err
}

func (plc *PLC) WriteTagString(tagName string, value string) error {
	return plc.client.Write(tagName, value)
}

func (plc *PLC) ReadTag(tagName string, tagType string, length int) (any, error) {
	switch tagType {
	case "bool":
		var tagValue bool
		err := plc.client.Read(tagName, &tagValue)
		return tagValue, err
	case "int":
		var tagValue int16
		err := plc.client.Read(tagName, &tagValue)
		return tagValue, err
	case "dint":
		var tagValue int32
		err := plc.client.Read(tagName, &tagValue)
		return tagValue, err
	case "real":
		var tagValue float32
		err := plc.client.Read(tagName, &tagValue)
		return tagValue, err
	case "string":
		var tagValue string
		err := plc.client.Read(tagName, &tagValue)
		return tagValue, err
	case "[]int":
		values := make([]uint16, length)
		for i := 0; i < length; i++ {
			elementName := fmt.Sprintf("%s[%d]", tagName, i)
			value, err := plc.ReadTag(elementName, "int", 0)
			if err != nil {
				return nil, fmt.Errorf("problem reading element %d of %s: %v", i, tagName, err)
			}
			intValue, ok := value.(int16)
			if !ok {
				return nil, fmt.Errorf("element %d of %s has incorrect type: %T", i, tagName, value)
			}
			values[i] = uint16(intValue)
		}
		return values, nil
	case "[]dint":
		values := make([]int32, length)
		for i := 0; i < length; i++ {
			elementName := fmt.Sprintf("%s[%d]", tagName, i)
			value, err := plc.ReadTag(elementName, "dint", 0)
			if err != nil {
				return nil, fmt.Errorf("problem reading element %d of %s: %v", i, tagName, err)
			}
			intValue, ok := value.(int32)
			if !ok {
				return nil, fmt.Errorf("element %d of %s has incorrect type: %T", i, tagName, value)
			}
			values[i] = intValue
		}
		return values, nil
	case "[]real":
		values := make([]float32, length)
		for i := 0; i < length; i++ {
			elementName := fmt.Sprintf("%s[%d]", tagName, i)
			value, err := plc.ReadTag(elementName, "real", 0)
			if err != nil {
				return nil, fmt.Errorf("problem reading element %d of %s: %v", i, tagName, err)
			}
			realValue, ok := value.(float32)
			if !ok {
				return nil, fmt.Errorf("element %d of %s has incorrect type: %T", i, tagName, value)
			}
			values[i] = realValue
		}
		return values, nil
	default:
		return nil, fmt.Errorf("unsupported tag type: %s", tagType)
	}
}

func (plc *PLC) WriteTag(tagName string, tagType string, tagValue interface{}) (any, error) {
	switch tagType {
	case "bool":
		return tagValue, plc.client.Write(tagName, tagValue.(bool))
	case "int":
		return tagValue, plc.client.Write(tagName, tagValue.(int16))
	case "dint":
		return tagValue, plc.client.Write(tagName, tagValue.(int32))
	case "real":
		return tagValue, plc.client.Write(tagName, tagValue.(float32))
	case "string":
		return tagValue, plc.client.Write(tagName, tagValue.(string))
	default:
		log.Printf("Incorrect data type")
	}
	return nil, fmt.Errorf("unsupported tag type: %s", tagType)
}

func LoadFromEthernetIPYAML(cfg *config.Config, plc *PLC, yamlPath string) (map[string]interface{}, error) {
	tag := cfg.Values["PLC_TAG"]
	length := 100 // default fallback
	if lstr, ok := cfg.Values["ETHERNET_IP_LENGTH"]; ok {
		if l, err := strconv.Atoi(lstr); err == nil && l > 0 {
			length = l
		}
	}
	rawDataAny, err := plc.ReadTag(tag, "[]int", length)
	if err != nil {
		log.Printf("Error reading from Ethernet/IP: %v", err)
		return nil, err
	}
	rawData, ok := rawDataAny.([]uint16)
	if !ok {
		log.Printf("Invalid data type returned from Ethernet/IP read")
		return nil, err
	}
	return LoadPLCDataMapFromYAML(yamlPath, rawData)
}
