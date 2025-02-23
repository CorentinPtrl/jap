package jap

import (
	"fmt"
	"github.com/Letsu/jap/utils"
	"github.com/google/go-cmp/cmp"
	"regexp"
	"strconv"
)

type Eigrp struct {
	ProcessID int
	Network   []EigrpNetwork `reg:"network.*" cmd:"network"`
}

type EigrpNetwork struct {
	NetworkNumber string `reg:"network (\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3})(?: (\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}))?" cmd:" %s"`
	WildCard      string `reg:"network (?:\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}) (\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3})" cmd:" %s"`
}

func (o *Eigrp) Parse(part string) error {
	re := regexp.MustCompile(`router eigrp (\d+)`)
	head := re.FindStringSubmatch(part)
	o.ProcessID, _ = strconv.Atoi(head[1])

	err := processParse(part, o)
	if err != nil {
		return err
	}

	return nil
}

func (o Eigrp) Generate() (string, error) {
	config := fmt.Sprintf("router eigrp %d\n", o.ProcessID)

	generated, err := Generate(o)
	if err != nil {
		return "", err
	}
	config = config + generated

	return config, nil
}

func (o Eigrp) Diff(b Eigrp) ([]string, error) {
	var r utils.DiffReporter
	cmp.Equal(o, b, cmp.Reporter(&r))
	return r.Fields(), nil
}

func (o Eigrp) Merge(b Eigrp) (string, error) {
	diff, err := o.Diff(b)
	if err != nil {
		return "", err
	}
	config := fmt.Sprintf("router eigrp %d\n", o.ProcessID)
	for _, s := range diff {
		cmd, err := GenerateFieldByPath(b, s)
		if err != nil {
			return "", err
		}
		config = config + fmt.Sprintf("  no %s\n", cmd)
	}
	generated, err := Generate(o)
	if err != nil {
		return "", err
	}
	config = config + generated

	return config, nil
}
