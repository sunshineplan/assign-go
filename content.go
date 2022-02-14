package main

import (
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/sunshineplan/utils/csv"
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

func (a *assignByContent) load(r readers) error {
	done := make(chan error, 2)
	go func() {
		var err error
		defer func() { done <- err }()

		a.Names, a.Scale, err = loadName(r.name)
		if err != nil {
			return
		}
		sort.Slice(a.Names, func(i, j int) bool { return a.Names[i].Scale > a.Names[j].Scale })
	}()
	go func() {
		var err error
		defer func() { done <- err }()

		rows := csv.FromReader(r.content)
		for rows.Next() {
			var id string
			var number int
			if err = rows.Scan(&id, &number); err != nil {
				return
			}
			a.Contents = append(a.Contents, content{ID: id, Number: number})
			a.Total += number
		}
		sort.Slice(a.Contents, func(i, j int) bool { return a.Contents[i].Number > a.Contents[j].Number })
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
					names = append(
						names,
						struct {
							Index int
							Name  name
						}{i, item},
					)
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
	return csv.ExportUTF8(nil, a.Contents, w)
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
