package jap

import (
	"github.com/mcuadros/go-defaults"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

type TestInterface struct {
	FileName string
}

func TestParseInterface(t *testing.T) {
	var testInters []TestInterface
	// Layer 3 interface
	layer3 := TestInterface{
		FileName: "layer3.txt",
	}
	testInters = append(testInters, layer3)

	// Layer 2 trunk interface
	layer2trunk := TestInterface{
		FileName: "layer2trunk.txt",
	}
	testInters = append(testInters, layer2trunk)

	// Layer 2 access interface
	layer2access := TestInterface{
		FileName: "layer2access.txt",
	}
	testInters = append(testInters, layer2access)

	for _, testInter := range testInters {
		// Open File
		content, err := ioutil.ReadFile("testFiles/interface/" + testInter.FileName)
		if err != nil {
			t.Error(err)
		}
		inter := new(CiscoInterface)
		defaults.SetDefaults(inter)
		err = inter.Parse(string(content))
		if err != nil {
			t.Error(err)
		}

		trunkAllowed := inter.TrunkAllowedVlan
		inter.TrunkAllowedVlan = []int{}
		generated, err := inter.Generate()
		if err != nil {
			t.Error(err)
		}

		re := regexp.MustCompile(`(?m)^\s*switchport trunk allowed vlan( add)? ([\d,-]+)`)
		trunkAllowedFile := re.FindAllStringSubmatch(string(content), -1)
		checkConfig := re.ReplaceAllString(string(content), "")
		re = regexp.MustCompile(`(?m)^\s*$[\r\n]*|[\r\n]+\s+\z`)
		checkConfig = re.ReplaceAllString(checkConfig, "")

		if generated != checkConfig {
			t.Error("Config wrong parsed or generated \n File: \n", checkConfig, "\n Generated: \n", generated)
		}

		var allowedVlans []int
		for _, d := range trunkAllowedFile {
			seperated := strings.Split(d[2], ",")
			for _, number := range seperated {
				if strings.Contains(number, "-") {
					vlanSplit := strings.Split(number, "-")
					from, _ := strconv.Atoi(vlanSplit[0])
					to, _ := strconv.Atoi(vlanSplit[1])
					for j := from; j <= to; j++ {
						allowedVlans = append(allowedVlans, j)
					}
					continue
				}
				vlanI, _ := strconv.Atoi(number)
				allowedVlans = append(allowedVlans, vlanI)
			}
		}

		for i, vlan := range allowedVlans {
			if trunkAllowed[i] != vlan {
				t.Error("Vlan allowed list falsch: vlan parsed-", trunkAllowed[i], "vlan file-", allowedVlans)
				break
			}
		}
	}
}
