package main

import (
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"strings"
)

type assignByName struct {
	Total int
	Scale float64
	Names []name
}

func (a *assignByName) load(r readers) error {
	var err error
	a.Names, a.Scale, err = loadName(r.name)
	if err != nil {
		return err
	}
	rand.Shuffle(len(a.Names), func(i, j int) { a.Names[i], a.Names[j] = a.Names[j], a.Names[i] })
	return nil
}

func (a *assignByName) assign() {
	var name name
	total, scale := a.Total, a.Scale
	for i := 0; i < len(a.Names); i++ {
		name, a.Names = a.Names[0], a.Names[1:]
		name.Count = randCeilFloor(float64(total) / scale * name.Scale)
		a.Names = append(a.Names, name)
		total -= name.Count
		scale -= name.Scale
	}
}

func (a assignByName) export(w io.Writer) error {
	if math.Mod(float64(a.Total), a.Scale) == 0 {
		fmt.Printf("Total: %d, Assign: %d(Scale: %g), Per number: %g\n", a.Total, len(a.Names), a.Scale, float64(a.Total)/a.Scale)
	} else {
		fmt.Printf("Total: %d, Assign: %d(Scale: %g), Per number: %g-%g\n",
			a.Total, len(a.Names), a.Scale, math.Floor(float64(a.Total)/a.Scale), math.Ceil(float64(a.Total)/a.Scale))
	}
	for _, i := range a.Names {
		fmt.Printf("%s: %d\n", i.Name, i.Count)
		if _, err := io.WriteString(w, strings.Repeat(i.Name+"\n", i.Count)); err != nil {
			return err
		}
	}
	return nil
}

func randCeilFloor(n float64) int {
	if rand.N(2) == 0 {
		return int(math.Ceil(n))
	}
	return int(math.Floor(n))
}
