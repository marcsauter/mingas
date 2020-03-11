package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func main() {
	var (
		// AMV is the Atemminutenvolumen
		amv = 30.0
		// bottles volume in liters
		bottles = []int{10, 12, 15, 18, 20, 24}
		// depthStart for calculations
		depthStart = 60
		// depthEnd for calculations
		depthEnd = 0
		// depthStep for calculations
		depthStep = 5
	)

	if len(os.Args) == 2 {
		v, err := strconv.ParseFloat(os.Args[1], 64)
		if err != nil {
			log.Fatal(err)
		}

		amv = v
	}

	data := make(map[int][]*point)

	for _, v := range bottles {
		points := []*point{}

		if depthStart < depthEnd {
			depthStart, depthEnd = depthEnd, depthStart
		}

		for d := depthStart; d > depthEnd; d -= depthStep {
			p := &point{
				Depth: float64(d),
			}

			p.Calc(amv)

			points = append(points, p)
		}

		data[v] = points
	}

	if err := plotChart(amv, bottles, data); err != nil {
		log.Fatal(err)
	}
}

type point struct {
	Depth  float64
	MinGas float64
}

func (p *point) Calc(amv float64) {
	const (
		// time at depth while trying to solve the problem
		tdepth = 1
	)

	stepsMin := p.Depth / 2 / 3
	stepsMax := stepsMin

	if int(p.Depth/2)%3 > 0 {
		stepsMax = stepsMin + 1
	}

	// air consumption at depth while trying to solve the problem
	air0 := 2.0 * amv * tdepth * (1 + (p.Depth / 10))

	// air consumption during the ascent to half the depth
	avDepth11 := (stepsMin*3 + p.Depth) / 2.0
	tascent11 := (p.Depth - stepsMin*3) / 10.0
	air1 := 2.0 * amv * tascent11 * (1 + (avDepth11 / 10.0))

	avDepth12 := (stepsMax*3 + p.Depth) / 2.0
	tascent12 := (p.Depth - stepsMax*3) / 10.0
	air2 := 2.0 * amv * tascent12 * (1 + (avDepth12 / 10.0))

	p.MinGas = air0 + air1 + air2
}

func plotChart(amv float64, bottles []int, data map[int][]*point) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	p.Title.Text = fmt.Sprintf("mingas for AMV = %.0f l/min", amv)
	p.X.Label.Text = "depth [m]"
	p.Y.Label.Text = "mingas [bar]"
	g := plotter.NewGrid()
	p.Add(g)

	lp := []interface{}{}
	for _, v := range bottles {
		lp = append(lp, fmt.Sprintf("%dl", v))
		points := plotter.XYs{}

		for _, p := range data[v] {
			points = append(points, plotter.XY{
				X: p.Depth,
				Y: p.MinGas / float64(v),
			})
		}

		lp = append(lp, points)
	}

	if err := plotutil.AddLinePoints(p, lp...); err != nil {
		return err
	}

	// Save the plot to a PNG file.
	if err := p.Save(180*vg.Millimeter, 140*vg.Millimeter, "mingas.png"); err != nil {
		panic(err)
	}

	return err
}
