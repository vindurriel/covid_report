package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type stringMap = map[string]interface{}

var yearRegexp = regexp.MustCompile(`/yqtb/(20\d\d)\d\d/(\w+)\.shtml`)
var monthDayRegexp = regexp.MustCompile(`(\d+)月(\d+)日`)

//go:embed provinces.yaml
var provinceConf []byte

type province struct {
	ID     string `yaml:"id" json:"id"`
	Name   string `yaml:"name" json:"name"`
	Index  int    `yaml:"-" json:"index"`
	Travel string `yaml:"travel" json:"travel"`
}

type provinceConfType struct {
	Provinces []*province `yaml:"provinces"`
}

type dayDataType struct {
	TS string `json:"ts"`
	// provinces with new cases
	Data map[int]uint `json:"data"`
	// days of no new case since last time by each province
	Data2 map[int]uint `json:"data2,omitempty"`
}

type covidReportType struct {
	Provinces []*province    `json:"provinces"`
	Days      []*dayDataType `json:"days"`
}

var provinces []string
var provinceSlice []*province
var provinceRegexp *regexp.Regexp
var provinceNameMap = map[int]string{}
var provinceIndexMap = map[string]int{}
var provinceTravelMap = map[int]string{}

const backDays = 14

func init() {
	if err := yaml.Unmarshal(provinceConf, &provinceSlice); err != nil {
		panic(err)
	}
	for i, x := range provinceSlice {
		x.Index = i
		provinces = append(provinces, x.Name)
		provinceNameMap[i] = x.Name
		provinceIndexMap[x.Name] = i
		provinceTravelMap[i] = x.Travel
	}
	provinceRegexp = regexp.MustCompile(fmt.Sprintf("(%v)", strings.Join(provinces, "|")))
}

var domesticRegexp = regexp.MustCompile(`(?:本土病例\d+例|\d+例为本土病例|新增确诊病例\d+例(?:，均为本土病例)?|肺炎确诊病例\d+例|均为本土病例)（(.+?)）`)
var hubeiRegexp = regexp.MustCompile(`湖北新增确诊病例(\d+)例`)
var singleProvince = regexp.MustCompile(`均?在(.+)`)
var provinceCountRegexp = regexp.MustCompile(`(.+?)(\d+)例`)
var allForeignRegexp = regexp.MustCompile(`均为境外输入病例`)

const beijingIndex = 29
const jilinIndex = 12
const hubeiIndex = 14

var fixDataMap = map[string]map[int]uint{
	"2021-11-30": {},
	"2021-08-05": {},
	"2020-07-15": {},
	"2020-06-12": {beijingIndex: 6},
	"2020-06-11": {beijingIndex: 1},
	"2020-06-03": {},
	"2020-06-02": {},
	"2020-05-28": {},
	"2020-05-26": {},
	"2020-05-14": {jilinIndex: 4},
	"2020-05-11": {},
	"2020-05-04": {},
	"2020-05-01": {},
	"2020-04-17": {},
	"2020-04-16": {},
}

func main() {
	var (
		bs      []byte
		err     error
		pageMap = stringMap{}
		yearMap = map[string]string{}
		dateMap = map[string]map[int]uint{}
	)
	if bs, err = ioutil.ReadFile("./crawler/puppeteer/sitemap.json"); err != nil {
		panic(err)
	}
	if err = json.Unmarshal(bs, &pageMap); err != nil {
		panic(err)
	}
	for k := range pageMap {
		groups := yearRegexp.FindStringSubmatch(k)
		year, pageID := groups[1], groups[2]
		yearMap[pageID] = year
	}

	var pageIds []string
	if err = filepath.WalkDir("./crawler/puppeteer/data", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		pageID := d.Name()
		// year
		year := yearMap[pageID]
		if year == "" {
			return fmt.Errorf("year not found: %v", pageID)
		}
		pageIds = append(pageIds, pageID)
		if bs, err = ioutil.ReadFile(path); err != nil {
			return err
		}
		s := string(bs)
		var ts string
		var intVal int64
		var group []string
		if group = monthDayRegexp.FindStringSubmatch(s); len(group) == 0 {
			fmt.Printf("month and day not found: %v", s)
			return nil
		}
		// month and day
		month, day := group[1], group[2]
		if month == "12" && day == "31" {
			if intVal, err = strconv.ParseInt(year, 10, 64); err != nil {
				return err
			}
			year = fmt.Sprint(intVal - 1)
		}
		ts = fmt.Sprintf("%v-%02s-%02s", year, month, day)
		dateMap[ts] = map[int]uint{}
		if year == "2020" && month == "1" {
			return nil
		} else if allForeignRegexp.MatchString(s) {
			return nil
		} else if fixData, ok := fixDataMap[ts]; ok {
			dateMap[ts] = fixData
		} else if group = hubeiRegexp.FindStringSubmatch(s); len(group) > 0 {
			if intVal, err = strconv.ParseInt(group[1], 10, 64); err != nil {
				return err
			}
			dateMap[ts][hubeiIndex] = uint(intVal)
		} else if group = domesticRegexp.FindStringSubmatch(s); len(group) > 0 {
			// province counts
			group2 := singleProvince.FindStringSubmatch(group[1])
			if strings.Index(group[1], "；") < 0 && len(group2) > 0 {
				// single province
				if group2 = provinceRegexp.FindStringSubmatch(group2[1]); len(group2) > 0 {
					p := group2[1]
					dateMap[ts][provinceIndexMap[p]] = 1
				}
			} else {
				// multiple province
				for _, ss := range strings.Split(group[1], "；") {
					for _, sss := range strings.Split(ss, "，") {
						var p, count string
						var index int
						if group = provinceCountRegexp.FindStringSubmatch(sss); len(group) == 0 {
							continue
						}
						p, count = group[1], group[2]
						if group = provinceRegexp.FindStringSubmatch(p); len(group) == 0 {
							continue
						}
						index = provinceIndexMap[group[1]]
						if intVal, err = strconv.ParseInt(count, 10, 64); err != nil {
							return err
						}
						dateMap[ts][index] = uint(intVal)
					}
				}
			}
			if len(dateMap[ts]) == 0 {
				fmt.Printf("zero %v %v %v\n", ts, pageID, s)
			}
		} else {
			fmt.Printf("domestic not found: %v %v \n%v\n\n", ts, pageID, s)
		}
		return nil
	}); err != nil {
		panic(err)
	}
	begin := time.Date(2020, 2, 1, 0, 0, 0, 0, time.Local)
	now := time.Now()
	o := covidReportType{
		Provinces: provinceSlice,
	}
	dayDuration := 24 * time.Hour
	lastCaseDayMap := map[int]time.Time{}
	for day := begin; day.Before(now); day = day.Add(dayDuration) {
		ts := day.Format("2006-01-02")
		dayData := &dayDataType{TS: ts}
		if m := dateMap[ts]; m != nil {
			dayData.Data = m
			for k := range m {
				lastCaseDayMap[k] = day
			}
		}
		noCaseMap := map[int]uint{}
		for i := 0; i < backDays; i++ {
			for k := range provinceSlice {
				if lastCaseDay, ok := lastCaseDayMap[k]; ok {
					if noCaseDays := uint(day.Sub(lastCaseDay) / dayDuration); noCaseDays >= backDays {
						noCaseMap[k] = noCaseDays
					}
				}
			}
		}
		dayData.Data2 = noCaseMap
		o.Days = append(o.Days, dayData)
	}
	if bs, err = json.MarshalIndent(o, "", "  "); err != nil {
		panic(err)
	}
	if err = os.WriteFile("docs/data.json", bs, 0644); err != nil {
		panic(err)
	}
}
