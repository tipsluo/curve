package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"time"
)

type EodRecds []EodRecd

const (
	NoValue = -1
)

type Market struct {
	Symbols map[string]int
	Dates   []time.Time
	ERs     []EodRecd
}

//<ticker>,<date>,<open>,<high>,<low>,<close>,<vol>
//A,20190924,78.29,78.39,75.64,75.97,4317500
type EodRecd struct {
	SmblIndx int
	DateIndx int
	Open     float32
	High     float32
	Low      float32
	Close    float32
	Volume   int64
}

const (
	NyseFileType = "NYSE_"
	EodDir       = "edata"
)

func LoadAllMarketData() Market {
	var m Market

	m.Symbols = make(map[string]int)

	var err error
	if m.Dates, err = EodFileDates(); err != nil {
		fmt.Printf("ERROR: failed to get all dates")
		os.Exit(1)
	}

	ReadNyseCsvs(&m)

	Sort(&(m.ERs))

	return m
}

func ParseEodCsv(m *Market, record []string) EodRecd {
	smblIndx, ok := m.Symbols[record[0]]
	if !ok {
		smblIndx = len(m.Symbols)
		m.Symbols[record[0]] = smblIndx
	}

	date, err := time.Parse("20060102", record[1])
	if err != nil {
		log.Fatalln(err)
	}

	open, err := strconv.ParseFloat(record[2], 32)
	if err != nil {
		log.Fatalln(err)
	}

	high, err := strconv.ParseFloat(record[3], 32)
	if err != nil {
		log.Fatalln(err)
	}

	low, err := strconv.ParseFloat(record[4], 32)
	if err != nil {
		log.Fatalln(err)
	}

	close, err := strconv.ParseFloat(record[5], 32)
	if err != nil {
		log.Fatalln(err)
	}

	volume, err := strconv.ParseInt(record[6], 10, 64)
	if err != nil {
		log.Fatalln(err)
	}

	eodRecd := EodRecd{
		SmblIndx: smblIndx,
		DateIndx: m.GetDateIndex(date),
		Open:     float32(open),
		High:     float32(high),
		Low:      float32(low),
		Close:    float32(close),
		Volume:   volume,
	}

	return eodRecd
}

func ReadNyseCsvs(m *Market) {
	var (
		eodRecds, eodRecds1, eodRecds2 []EodRecd
	)

	for i, date := range m.Dates {
		fmt.Printf("\033[2K\r")
		fmt.Printf("%v", date)

		filename := NyseCsvName(date)
		if !IsFileExist(filename) {
			continue
		}

		records, err := ReadCsvFile(filename)
		if err != nil {
			log.Fatalln(err)
		}

		var p *[]EodRecd
		if i%2 == 0 {
			p = &eodRecds1
		} else {
			p = &eodRecds2
		}

		for _, n := range records[1:] {
			eodRecd := ParseEodCsv(m, n)

			*p = append(*p, eodRecd)
		}

		eodRecds = append(eodRecds, *p...)
	}

	fmt.Println()

	m.ERs = eodRecds
}

func NyseCsvName(date time.Time) string {
	return EodDir + "/" + NyseFileType + fmt.Sprintf("%4d%02d%02d", date.Year(), date.Month(), date.Day()) + ".txt"
}

func EodFileDates() ([]time.Time, error) {
	var (
		dates []time.Time
	)

	files := GetDirFileList(EodDir, NyseFileType+"????????.txt")

	sort.Strings(files)

	r, _ := regexp.Compile(EodDir + "/" + NyseFileType + `([\d]{8})\.txt`)

	for i := range files {
		sdate := r.FindStringSubmatch(files[i])
		date, err := time.Parse("20060102", sdate[1])
		if err != nil {
			return nil, err
		}
		dates = append(dates, date)
	}

	return dates, nil
}

func GetDirFileList(dir string, pattern string) []string {
	var files []string

	pattern = dir + "/" + pattern

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		matched, err1 := filepath.Match(pattern, path)
		if err1 != nil {
			log.Fatalln(err1)
		}
		if matched {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	return files
}

func ReadFile(filename string) []byte {
	jsonFile, err := os.Open(filename)
	defer jsonFile.Close()

	if err != nil {
		log.Fatalln(err)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	return byteValue
}

func ReadCsvBytes(buffer []byte) ([][]string, error) {
	r := csv.NewReader(bytes.NewReader(buffer))

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func ReadCsvFile(filename string) ([][]string, error) {
	return ReadCsvBytes(ReadFile(filename))
}

func IsFileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func Sort(eodRecds *([]EodRecd)) {
	sort.Slice(*eodRecds, func(i, j int) bool {
		if (*eodRecds)[i].SmblIndx != (*eodRecds)[j].SmblIndx {
			return (*eodRecds)[i].SmblIndx < (*eodRecds)[j].SmblIndx
		} else {
			return (*eodRecds)[i].DateIndx < (*eodRecds)[j].DateIndx
		}
	})
}

func (mrkt *Market) Len() int {
	return len(mrkt.ERs)
}

func (mrkt *Market) DateLen() uint16 {
	return uint16(len(mrkt.Dates))
}

func (mrkt *Market) SymbolName(id int) string {
	for k, n := range mrkt.Symbols {
		if n == id {
			return k
		}
	}

	return ""
}

func (mrkt *Market) Date(index uint16) time.Time {
	return mrkt.Dates[index]
}

func (mrkt *Market) GetDateIndex(d time.Time) int {
	i := sort.Search(len(mrkt.Dates), func(i int) bool { return !mrkt.Dates[i].Before(d) })
	if i >= len(mrkt.Dates) || mrkt.Dates[i] != d {
		return NoValue
	} else {
		return i
	}
}

func (mrkt Market) PrintStruct(i interface{}) {
	v := reflect.ValueOf(i).Elem()
	t := v.Type()

	for j := 0; j < t.NumField(); j++ {
		fmt.Print("{")
		if t.Field(j).Name == "SmblIndx" {
			fmt.Printf("%s, ", mrkt.SymbolName(v.Field(j).Interface().(int)))
		} else if t.Field(j).Name == "DateIndx" {
			fmt.Printf("%v, ", mrkt.Dates[v.Field(j).Interface().(int)])
		} else {
			fmt.Printf("%v, ", v.Field(j).Interface())
		}
		fmt.Print("}")
	}
}

func (mrkt Market) GetSmblCurve(smbl string, field string) Stock {
	values := make(map[int]float32)

	smblIndx, ok := mrkt.Symbols[smbl]
	if !ok {
		return Stock{
			Symbol:    "",
			CurveBase: CurveBase{},
		}
	}

	for _, eod := range mrkt.ERs {
		if eod.SmblIndx == smblIndx {
			r := reflect.ValueOf(eod)
			f := reflect.Indirect(r).FieldByName(field)
			values[eod.DateIndx] = float32(f.Float())
		}
	}

	return Stock{
		Symbol: smbl,
		CurveBase: CurveBase{
			Values: values,
		},
	}
}
