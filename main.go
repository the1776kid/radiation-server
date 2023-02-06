package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

//go:embed root.html
var rootHTML []byte

//go:embed favicon.png
var favicon []byte

var (
	storageDirectory = "/usr/share/radiation-server/"
)

type Rad struct {
	TS     []string
	Values []float64
}

func ifLessThan10(i int) string {
	var fs string
	if i < 10 {
		fs = fmt.Sprintf("0%d", i)
	} else {
		fs = fmt.Sprintf("%d", i)
	}
	return fs
}

func dayStamp() string {
	now := time.Now()
	return fmt.Sprintf("%s_%s", ifLessThan10(int(now.Month())), ifLessThan10(now.Day()))
}

func timestamp() string {
	now := time.Now()
	return fmt.Sprintf("%s%s", ifLessThan10(now.Hour()), ifLessThan10(now.Minute()))
}

func jsonify(i interface{}) ([]byte, error) {
	jd, err := json.MarshalIndent(i, "", " ")
	if err != nil {
		return nil, fmt.Errorf("error::%v\n", err)
	}
	return jd, nil
}

func readDir(location string) ([]string, error) {
	dir, err := os.ReadDir(location)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s::%v\n", location, err)
	}
	var list []string
	for _, entry := range dir {
		if entry.Name()[len(entry.Name())-5:] == ".json" {
			list = append(list, entry.Name())
		}
	}
	return list, nil
}

func (rad *Rad) fileOperations(action string, filename /*day*/ string) ([]byte, error) {
	switch action {
	case "write":
		jd, err := jsonify(rad)
		if err != nil {
			return nil, fmt.Errorf("error::%v\n", err)
		}
		if err := os.WriteFile(storageDirectory+filename+".json", jd, 0555); err != nil {
			return nil, fmt.Errorf("error::%v\n", err)
		}
		rad.Values = []float64{}
		rad.TS = []string{}
	case "read":
		file, err := os.ReadFile(storageDirectory + filename)
		if err != nil {
			return nil, fmt.Errorf("error::%v\n", err)
		}
		return file, nil
	}
	return nil, nil
}

func (rad *Rad) serve() {
	http.HandleFunc("/favicon", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(favicon); err != nil {
			log.Printf("ERROR::%v\n", err)
			return
		}
	})
	http.HandleFunc("/history_api", func(w http.ResponseWriter, r *http.Request) {
		action := r.FormValue("action")
		switch action {
		case "list":
			dl, err := readDir(storageDirectory)
			if err != nil {
				log.Fatalf("error::%v\n", err)
			}
			mj, err := jsonify(struct {
				List []string `json:"list"`
			}{
				List: dl,
			})
			if err != nil {
				log.Fatalf("error::%v\n", err)
			}
			if _, err := w.Write(mj); err != nil {
				log.Fatalf("error::%v\n", err)
			}
			return
		case "data":
			file := r.FormValue("file")
			fb, err := rad.fileOperations("read", file)
			if err != nil {
				log.Fatalf("error::%v\n", err)
			}
			if _, err := w.Write(fb); err != nil {
				log.Fatalf("error::%v\n", err)
			}
			return
		default:
			log.Println("error::/history_api?action=nil::no action value was provided")
			return
		}
	})
	http.HandleFunc("/chart_api", func(w http.ResponseWriter, r *http.Request) {
		jd, err := jsonify(rad)
		if err != nil {
			log.Printf("error::%v\n", err)
			return
		}
		if _, err := w.Write(jd); err != nil {
			log.Printf("error::%v\n", err)
			return
		}
	})
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		ts := timestamp()
		if ts == "0000" {
			if _, err := rad.fileOperations("write", dayStamp()); err != nil {
				log.Fatalf("error::%v\n", err)
			}
		}
		cpm := r.FormValue("cpm")
		cpmInt, err := strconv.ParseFloat(cpm, 64)
		if err != nil {
			log.Printf("ERROR::%v\n", err)
		}
		if cpmInt < 1 {
			return
		}
		ush := cpmInt / 151
		rad.TS = append(rad.TS, ts)
		rad.Values = append(rad.Values, ush)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(rootHTML); err != nil {
			log.Printf("ERROR::%v\n", err)
			return
		}
	})
	if err := http.ListenAndServe(":8320", nil); err != nil {
		log.Printf("ERROR:HTTP:%v\n", err)
	}
}

func main() {
	r := Rad{}
	r.serve()
}
