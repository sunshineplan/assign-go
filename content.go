package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/sunshineplan/utils/export"
)

type content struct {
	ID     string
	Number int
	Name   string
}

type assignByContent struct {
	Total    int
	Scale    float64
	Contents []content
	Names    []name
}

func (a *assignByContent) load(r ...io.Reader) error {
	done := make(chan error, 2)
	go func() {
		var err error
		a.Names, a.Scale, err = loadName(r[0])
		if err != nil {
			done <- err
			return
		}
		sort.Slice(a.Names, func(i, j int) bool { return a.Names[i].Scale > a.Names[j].Scale })
		done <- nil
	}()
	go func() {
		r := csv.NewReader(r[1])
		r.FieldsPerRecord = 2
		id, number, err := readCSVLine(r)
		if err == nil {
			a.Contents = append(a.Contents, content{ID: id, Number: number})
			a.Total += number
		}
		for {
			id, number, err := readCSVLine(r)
			if err == io.EOF {
				break
			}
			if err != nil {
				done <- err
				return
			}
			a.Contents = append(a.Contents, content{ID: id, Number: number})
			a.Total += number
		}
		sort.Slice(a.Contents, func(i, j int) bool { return a.Contents[i].Number > a.Contents[j].Number })
		done <- nil
	}()
	for i := 0; i < 2; i++ {
		if err := <-done; err != nil {
			return err
		}
	}
	return nil
}

func (a *assignByContent) assign() {
	contents := a.Contents
	a.Contents = nil
	for len(contents) > 0 {
		var content content
		var names []struct {
			Index int
			Name  name
		}
		refer := float64(a.Names[0].Number+contents[0].Number) / a.Names[0].Scale
		average := namesAverage(a.Names)
		for i, item := range a.Names {
			if len(contents) > 0 {
				if num := float64(item.Number) / item.Scale; num > refer && num > average {
					names = append(names, struct {
						Index int
						Name  name
					}{Index: i, Name: item})
				} else {
					content, contents = contents[0], contents[1:]
					content.Name = item.Name
					a.Contents = append(a.Contents, content)
					a.Names[i].Number += content.Number
					a.Names[i].Count++
				}
			}
		}
		for i := len(names) - 1; i >= 0; i-- {
			content, contents = contents[len(contents)-1], contents[:len(contents)-1]
			content.Name = names[i].Name.Name
			a.Contents = append(a.Contents, content)
			a.Names[names[i].Index].Number += content.Number
			a.Names[names[i].Index].Count++
		}
		sort.Slice(a.Names, func(i, j int) bool {
			return float64(a.Names[i].Number)/a.Names[i].Scale < float64(a.Names[j].Number)/a.Names[j].Scale
		})
	}
	sort.Slice(a.Contents, func(i, j int) bool {
		id1, err1 := strconv.Atoi(a.Contents[i].ID)
		id2, err2 := strconv.Atoi(a.Contents[j].ID)
		if err1 != nil && err2 != nil {
			return a.Contents[i].ID < a.Contents[j].ID
		} else if err1 == nil && err2 == nil {
			return id1 < id2
		} else if err1 != nil {
			return false
		} else {
			return true
		}
	})
}

func (a assignByContent) export(w io.Writer) error {
	fmt.Printf("Total: %d, Assign: %d(Scale: %g), Average: %d\n", a.Total, len(a.Names), a.Scale, int(float64(a.Total)/a.Scale))
	for _, i := range a.Names {
		fmt.Printf("%s\t%d\t%d\n", i.Name, i.Number, i.Count)
	}
	return export.CSVWithUTF8BOM([]string{"ID", "Number", "Name"}, a.Contents, w)
}

func readCSVLine(r *csv.Reader) (id string, number int, err error) {
	var record []string
	record, err = r.Read()
	if err != nil {
		return
	}
	id = record[0]
	number, err = strconv.Atoi(record[1])
	return
}

func namesAverage(names []name) float64 {
	length := len(names)
	if length-2 > 0 {
		names = names[1 : length-1]
		length -= 2
	}
	var total float64
	for _, i := range names {
		total += float64(i.Number) / i.Scale
	}
	return total / float64(length)
}
