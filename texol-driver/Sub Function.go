package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"

	"github.com/texol/texol-ble-driver/logfile"
)

type sensorJSON struct {
	BridgeUUID   string
	Channel      string
	Pipe         string
	BLEUUID      string
	SensorID     string
	ModuleName   string
	BLEFWVersion string
	RSSI         string
	Feature      interface{}
}

type Feature_Single struct {
	RPM                     uint16
	OA                      float32
	Peak                    float32
	PtoP                    float32
	CF                      float32
	PIB01                   float32
	PIB02                   float32
	PIB03                   float32
	PIB04                   float32
	PIB05                   float32
	PIB06                   float32
	PIB07                   float32
	PIB08                   float32
	PIB09                   float32
	PIB10                   float32
	Bearing                 float32
	Losseness               float32
	Misalignment            float32
	Unbalance               float32
	Gear_mesh               float32
	Van_Pass                float32
	Bearing_percentage      float32
	Losseness_percentage    float32
	Misalignment_percentage float32
	Unbalance_percentage    float32
	Gear_mesh_percentage    float32
	Van_Pass_percentage     float32
	OA_S                    int8
	Peak_S                  int8
	PtoP_S                  int8
	CF_S                    int8
	PIB01_S                 int8
	PIB02_S                 int8
	PIB03_S                 int8
	PIB04_S                 int8
	PIB05_S                 int8
	PIB06_S                 int8
	PIB07_S                 int8
	PIB08_S                 int8
	PIB09_S                 int8
	PIB10_S                 int8
}

type Feature_Tri struct {
	RPM                     uint16
	XOA                     float32
	XPIB01                  float32
	XPIB02                  float32
	XPIB03                  float32
	YOA                     float32
	YPIB01                  float32
	YPIB02                  float32
	YPIB03                  float32
	ZOA                     float32
	ZPIB01                  float32
	ZPIB02                  float32
	ZPIB03                  float32
	Bearing                 float32
	Misalignment            float32
	Unbalance               float32
	Bearing_percentage      float32
	Misalignment_percentage float32
	Unbalance_percentage    float32
	XOA_S                   int8
	XPIB01_S                int8
	XPIB02_S                int8
	XPIB03_S                int8
	YOA_S                   int8
	YPIB01_S                int8
	YPIB02_S                int8
	YPIB03_S                int8
	ZOA_S                   int8
	ZPIB01_S                int8
	ZPIB02_S                int8
	ZPIB03_S                int8
}

type Alive struct {
	Alive bool
}

var Chan_bridgeJoin = make(chan string, 10)
var Chan_SensorData = make(chan string, 10)
var Chan_SensorAlive = make(chan string, 10)

func init() {
	go read_bridgeJoin()
	go read_SensorData()
	go read_SensorAlive()
}

func read_bridgeJoin() {
	for payload := range Chan_bridgeJoin {
		EventChan <- eventPro{
			Message: "Pub",
			Value:   bridgeJoin(payload),
		}
	}
}

func read_SensorData() {
	for payload := range Chan_SensorData {
		pub_payload, err := sensor_dataandalive("Data", payload)
		if !err {
			EventChan <- eventPro{
				Message: "Pub",
				Value:   pub_payload,
			}
		}
	}
}

func read_SensorAlive() {
	for payload := range Chan_SensorAlive {
		pub_payload, err := sensor_dataandalive("Alive", payload)
		if !err {
			EventChan <- eventPro{
				Message: "Pub",
				Value:   pub_payload,
			}
		}
	}
}

func bridgeJoin(payload_bridgeJoin string) publishPro {
	gatewayuuid := strings.Split(payload_bridgeJoin, ",")[1][:12]
	topic := Topic_bridgeJoin + "/" + gatewayuuid
	Payload := "S,162b3a11-148f-40b1-8c45-47e1a824941f," + GatewayIP + ",10,4,4,-80,-80,3,65244,64475,64731*F0"
	logfile.Println("Bridge Join: " + gatewayuuid)
	return publishPro{
		Topic:   topic,
		Payload: Payload,
	}
}

// dataoralive input Data or Alive. Error return true
func sensor_dataandalive(dataoralive string, payload string) (publishPro, bool) {
	var pub_payload publishPro
	var info sensorJSON
	var mainboard string
	err := message_dataoralive(payload, &info, &mainboard)
	if err {
		return pub_payload, true
	}

	topic := ""
	switch dataoralive {
	case "Alive":
		{
			topic = "/heartbeat"
			if sensorAlive(mainboard, &info) {
				return pub_payload, true
			}
		}
	case "Data":
		{
			topic = ""
			if sensorData(mainboard, &info) {
				return pub_payload, true
			}
		}
	}

	pub_payload = publishPro{
		Topic:   fmt.Sprintf("/texol/%s/%s%s", info.ModuleName, info.SensorID, topic),
		Payload: toJSON(info),
	}

	return pub_payload, false
}

// error return true
func message_dataoralive(payload string, info *sensorJSON, mainboard *string) bool {
	//Start Byte is 55AA and payload length is 100
	if strings.ToUpper(strTOHexstr(payload[0:2])) != "55AA" || len(payload) != 100 { //Header Start and len check
		write_error_log("payload Error: payload is ", payload)
		return true
	}

	*info = sensorJSON{
		BridgeUUID: strings.ToUpper(strTOHexstr(payload[90:96])),
		Channel:    channel(strTOHexstr(payload[97:98])),
		Pipe:       strTOHexstr(payload[5:6]),
		BLEUUID:    strings.ToLower(strTOHexstr(payload[8:14])),
		//SensorID:     SensorID,
		BLEFWVersion: strings.ToUpper(strTOHexstr(payload[2:4])),
		RSSI:         fmt.Sprintf("-%d", []byte(payload[96:97])[0]),
	}

	*mainboard = payload[16 : 16+74] //data長度74

	return false
}

// error return true
func sensorData(payload string, info *sensorJSON) bool {
	//Check StartByte and StopByte
	if payload[:2] != string([]byte{0x53, 0x53}) || payload[len(payload)-2:] != string([]byte{0x53, 0x54}) {
		write_error_log("StartByte OR StopByte Error: payload is ", payload)
		return true
	}

	//Check SensorID 出現非大寫或數字
	SensorID := payload[2 : 2+6]
	matched, _ := regexp.MatchString("^[0-9A-Z]{6}$", SensorID)
	if !matched || SensorID == " BANK1" {
		write_error_log("SensorID Error: payload is ", payload)
		return true
	}

	if data_repeatedly(SensorID, payload) {
		write_error_log("Repeated Update: payload is ", payload)
		return true
	}

	var rpm uint16
	var dataArray []float32
	var QCs []int8
	if strtoData(payload[2+6:len(payload)-2], &rpm, &dataArray, &QCs) { //StartByte(2)+SensorID(6)........+StopByte(2)
		write_error_log("Value Error: ", payload)
		return true
	}
	//fmt.Printf("RPM: %d data: %f QC:%d\n", rpm, dataArray, QCs)

	var Feature interface{}
	ModuleName := moduleName(SensorID)
	switch ModuleName {
	case "211HM1-B1":
		Feature = Feature_Single{
			RPM:     rpm,
			OA:      dataArray[0],
			Peak:    dataArray[1],
			PtoP:    dataArray[2],
			CF:      dataArray[3],
			PIB01:   dataArray[4],
			PIB02:   dataArray[5],
			PIB03:   dataArray[6],
			PIB04:   dataArray[7],
			PIB05:   dataArray[8],
			PIB06:   dataArray[9],
			PIB07:   dataArray[10],
			PIB08:   dataArray[11],
			PIB09:   dataArray[12],
			PIB10:   dataArray[13],
			OA_S:    QCs[0],
			Peak_S:  QCs[1],
			PtoP_S:  QCs[2],
			CF_S:    QCs[3],
			PIB01_S: QCs[4],
			PIB02_S: QCs[5],
			PIB03_S: QCs[6],
			PIB04_S: QCs[7],
			PIB05_S: QCs[8],
			PIB06_S: QCs[9],
			PIB07_S: QCs[10],
			PIB08_S: QCs[11],
			PIB09_S: QCs[12],
			PIB10_S: QCs[13],
		}
	case "213MM1-B1":
		Feature = Feature_Tri{
			RPM:      rpm,
			XOA:      dataArray[0],
			XPIB01:   dataArray[1],
			XPIB02:   dataArray[2],
			XPIB03:   dataArray[3],
			YOA:      dataArray[4],
			YPIB01:   dataArray[5],
			YPIB02:   dataArray[6],
			YPIB03:   dataArray[7],
			ZOA:      dataArray[8],
			ZPIB01:   dataArray[9],
			ZPIB02:   dataArray[10],
			ZPIB03:   dataArray[11],
			XOA_S:    QCs[0],
			XPIB01_S: QCs[1],
			XPIB02_S: QCs[2],
			XPIB03_S: QCs[3],
			YOA_S:    QCs[4],
			YPIB01_S: QCs[5],
			YPIB02_S: QCs[6],
			YPIB03_S: QCs[7],
			ZOA_S:    QCs[8],
			ZPIB01_S: QCs[9],
			ZPIB02_S: QCs[10],
			ZPIB03_S: QCs[11],
		}
	}

	info.SensorID = SensorID
	info.ModuleName = ModuleName
	info.Feature = Feature

	return false
}

// error return true
func sensorAlive(payload string, info *sensorJSON) bool {
	//StartByte(2)+SensorID(6)........+StopByte(2)
	//fmt.Printf("%x \n", payload)

	//Check StartByte and StopByte
	if payload[:2] != string([]byte{0x53, 0x53}) || payload[len(payload)-2:] != string([]byte{0x53, 0x54}) {
		write_error_log("StartByte OR StopByte Error(Alive): payload is ", payload)
		return true
	}

	//Check SensorID 出現非大寫或數字
	SensorID := payload[2 : 2+6]
	matched, _ := regexp.MatchString("^[0-9A-Z]{6}$", SensorID)
	if !matched || SensorID == " BANK1" {
		write_error_log("SensorID Error(Alive): payload is ", payload)
		return true
	}

	info.SensorID = SensorID
	ModuleName := moduleName(SensorID)
	info.ModuleName = ModuleName
	info.Feature = "Alive"

	return false
}

// error return true
func strtoData(payload string, rpm *uint16, dataArray *[]float32, QCs *[]int8) bool {
	byte_array := []byte(payload)
	//fmt.Printf("%x\n", byte_array)
	s := 0

	end := s + 2
	if BytesToInt(byte_array[s:end], rpm) {
		return true
	}
	s = end

	for index := 0; index < 14; index++ {
		var data float32
		end = s + 4
		err := BytesToFloat(byte_array[s:end], &data)
		if err || feature_check(data) {
			return true
		}
		*dataArray = append(*dataArray, data)
		s = end
	}

	end = s + 6
	bys := byte_array[s:end] //[]byte{0x1c, 0xa8, 0x23, 0x54, 0x20, 0x14}
	if BytesToQC(bys, QCs) {
		return true
	}
	//fmt.Println(QCs)

	//fmt.Printf("rpm: %d x:%f", &rpm, dataArray)
	return false
}

// error return true
func BytesToInt(bys []byte, data *uint16) bool {
	bytebuff := bytes.NewBuffer(bys)
	err := binary.Read(bytebuff, binary.BigEndian, data)
	if err != nil {
		return true
	}
	return false
}

// error return true
func BytesToFloat(bys []byte, data *float32) bool {
	buf := bytes.NewReader(bys)
	err := binary.Read(buf, binary.LittleEndian, data)
	if err != nil {
		return true
	}
	return false
}

// error return true
func BytesToQC(bys []byte, QCs *[]int8) bool {
	aa := make([]int8, 14)
	for index := 0; index < 3; index++ {
		var bit uint16
		err := BytesToInt(bys[index*2:index*2+2], &bit)
		if err {
			return true
		}
		//fmt.Printf("t %d: %016b\n", index, bit)
		bit = bit << 2
		//fmt.Printf("t %d: %016b\n", index, bit)
		for i := 0; i < len(aa); i++ {
			if bit&0x8000 == 0x8000 {
				//fmt.Printf("%d: %016b\n", i, bit)
				aa[i] = int8(index + 1)
			}
			bit = bit << 1
		}
		//fmt.Println(aa)
	}
	*QCs = aa
	return false
}

func strTOHexstr(str string) string {
	return fmt.Sprintf("%x", str)
}

func channel(str string) string {
	switch str {
	case "01":
		return "R"
	case "02":
		return "L"
	}
	return ""
}

func moduleName(SensorID string) string {
	switch SensorID[:2] {
	case "00":
		return "211HM1-B1"
	case "01":
		return "213MM1-B1"
	}
	return ""
}

// error return true
func feature_check(feature float32) bool {
	if feature < 0 || feature > 1000000 {
		return true
	}
	return false
}

type data_temporary struct {
	sensorID string
	payload  string
}

var data_temporarys []data_temporary

// Repeatedly return true
func data_repeatedly(SensorID string, payload string) bool {
	repe := false
	for _, d := range data_temporarys {
		if d.sensorID == SensorID { //找到sensor
			if d.payload == payload { //確認是否重複
				//重複
				repe = true
			} else {
				//不重複
				repe = false
			}
			d.payload = payload
			return repe
		}
	}
	//sensor 不存在
	data_temporarys = append(data_temporarys, data_temporary{SensorID, payload})
	//fmt.Println(data_temporarys)

	return repe
}

func write_error_log(mesg string, payload string) {
	logfile.Println(mesg + strTOHexstr(payload))
}
