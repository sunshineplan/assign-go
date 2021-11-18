package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sunshineplan/utils/csv"
)

var self string

func init() {
	exec, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get self path: %v", err)
	}
	self = filepath.Dir(exec)
	rand.Seed(time.Now().UnixNano())
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	fmt.Println(`
  name total int
        Assign by name with total number.
  content
        Assign by content.`)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	var task assign
	filename := "Result.txt"
	nameList, err := os.Open(filepath.Join(self, "NameList.txt"))
	if err != nil {
		log.Print(err)
		f, err := os.Create(filepath.Join(self, "NameList.txt"))
		if err == nil {
			defer f.Close()
			f.WriteString("Name1\nName2*2\n#Name3\nName4")
			log.Print("Sample NameList.txt is created.")
		}
		return
	}
	defer nameList.Close()
	r := readers{name: nameList}

	if narg := flag.NArg(); narg == 0 || (narg == 1 && flag.Arg(0) == "name") {
		fmt.Print("Please input total number: ")
		var input string
		fmt.Scanln(&input)
		total, err := strconv.Atoi(input)
		if err != nil {
			log.Fatalln("Bad total argument:", input)
		}
		task = &assignByName{Total: total}
	} else if narg == 1 && flag.Arg(0) == "content" {
		task = &assignByContent{}
		contentList, err := os.Open(filepath.Join(self, "ContentList.csv"))
		if err != nil {
			log.Print(err)
			f, err := os.Create(filepath.Join(self, "ContentList.csv"))
			if err == nil {
				defer f.Close()
				csv.ExportUTF8([]string{"ID", "Number"}, []content{{ID: "1", Number: 100}, {ID: "2", Number: 200}}, f)
				log.Print("Sample ContentList.csv is created.")
			}
			return
		}
		defer contentList.Close()
		filename = "Result.csv"
		r.content = contentList
	} else if narg == 2 && flag.Arg(0) == "name" {
		total, err := strconv.Atoi(flag.Arg(1))
		if err != nil {
			log.Fatalln("Unknown argument: name", flag.Arg(1))
		}
		task = &assignByName{Total: total}
	} else {
		log.Fatalln("Unknown arguments:", strings.Join(flag.Args(), " "))
	}

	if err := task.load(r); err != nil {
		log.Fatalln("Failed to load source:", err)
	}
	task.assign()

	var result *os.File
	result, err = os.Create(filepath.Join(self, filename))
	if err != nil {
		log.Println("Failed to save result:", err)
		result, err = ioutil.TempFile(self, strings.ReplaceAll(filename, "Result.", "Result-*."))
		if err != nil {
			log.Fatalln("Failed to create temporary file:", err)
		}
	}
	defer result.Close()
	if err := task.export(result); err != nil {
		log.Fatalln("Failed to export result:", err)
	}

	fmt.Println("Press enter key to exit . . .")
	fmt.Scanln()
}
