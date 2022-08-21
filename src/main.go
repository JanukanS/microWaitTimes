package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type WaitData struct {
	gorm.Model
	UnixTime uint
	Location string
	City     string
	Category string
	WaitTime uint
	Note     string
	Valid    bool
}

type locationData struct {
	Name        string `json:"Name"`
	Category    string `json:"Category"`
	WaitTime    string `json:"WaitTime"`
	Url         string `json:"URL"`
	Note        string `json:"Note"`
	Unavailable bool   `json:"TimesUnavailable"`
}

func (l locationData) convert(dataTime time.Time, city string) WaitData {
	var waitVal uint
	if !l.Unavailable {
		timeList := strings.Split(l.WaitTime, " ")
		hourVal, hourErr := strconv.Atoi(timeList[0])
		minVal, minErr := strconv.Atoi(timeList[2])

		if hourErr != nil && minErr != nil {
			waitVal = uint(60*hourVal + minVal)
		}
	}
	return WaitData{
		UnixTime: uint(dataTime.Unix()),
		Location: l.Name,
		City:     city,
		Category: l.Category,
		WaitTime: waitVal,
		Note:     l.Note,
		Valid:    !l.Unavailable,
	}
}

type cityData struct {
	Emergency []locationData `json:"Emergency"`
	Urgent    []locationData `json:"Urgent"`
}

func (cd cityData) encodeData(dataTime time.Time, cityName string) []WaitData {
	var wdSlice []WaitData
	for _, lData := range cd.Emergency {
		wdSlice = append(wdSlice, lData.convert(dataTime, cityName))
	}
	for _, lData := range cd.Urgent {
		wdSlice = append(wdSlice, lData.convert(dataTime, cityName))
	}
	return wdSlice
}

//type ahsData map[string]cityData

//func (ad ahsData) encodeCities(dataTime time.Time) []WaitData {
//	var wdSlice []WaitData
//	for cityName, cdInst := range ad {
//		wdSlice = append(wdSlice, cdInst.encodeData(dataTime, cityName)...)
//	}
//	return wdSlice
//}

type ahsApiData struct {
	timeReceived time.Time
	data         map[string]cityData
}

func (aad ahsApiData) encodeResp() []WaitData {
	var wdSlice []WaitData
	for cityName, cdInst := range aad.data {
		wdSlice = append(wdSlice, cdInst.encodeData(aad.timeReceived, cityName)...)
	}
	return wdSlice
}

func dumpData(dbInst *gorm.DB, wdSlice []WaitData) {
	for _, wd := range wdSlice {
		dbInst.Create(&wd)
	}
}

func collectData(dbInst *gorm.DB) {
	resp, err := http.Get("https://www.albertahealthservices.ca/Webapps/WaitTimes/api/waittimes")
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	data1 := ahsApiData{timeReceived: time.Now()}
	jsonErr := json.Unmarshal(body, &data1.data)
	if jsonErr != nil {
		log.Fatalln(jsonErr)
	}

	dumpData(dbInst, data1.encodeResp())
}

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatalln("failed to connect database")
	}

	err = db.AutoMigrate(&WaitData{})
	if err != nil {
		log.Fatalln(err)
	}
	for range time.Tick(time.Second * 120) {
		collectData(db)
		fmt.Println("Occurance")
	}
}
