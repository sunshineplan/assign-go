package main

import (
	"bufio"
	"io"
	"log"
	"strconv"
	"strings"
)

type assign interface {
	load(r ...io.Reader) error
	assign()
	export(w io.Writer) error
}

type name struct {
	Name   string
	Scale  float64
	Count  int
	Number int
}

func loadName(r io.Reader) ([]name, float64, error) {
	var names []name
	var scale float64
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line[:1] == "#" {
			continue
		} else {
			switch content := strings.Split(line, "*"); len(content) {
			case 0:
				continue
			case 1:
				names = append(names, name{Name: content[0], Scale: 1})
				scale++
			case 2:
				n, err := strconv.ParseFloat(content[1], 64)
				if err != nil || n < 0 {
					log.Println("Can't parse scale:", scanner.Text())
					continue
				}
				names = append(names, name{Name: content[0], Scale: n})
				scale += n
			default:
				log.Println("Can't parse line:", scanner.Text())
			}
		}
	}
	return names, scale, scanner.Err()
}
