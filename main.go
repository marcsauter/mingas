package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

const (
	// xSize is the size for the x axis
	xSize = 180 * vg.Millimeter
	// ySize is the size for the y axis
	ySize = 130 * vg.Millimeter
	// filename for the graph
	filename = "mingas.png"
)

// nolint: gochecknoglobals
var (
	// maxDepth start depth [m] for calculations
	maxDepth = 60
	// minDepth end depth [m] for calculations
	minDepth = 0
	// depthStep for calculations
	depthStep = 5
)

func main() {
	var (
		// AMV is the Atemminutenvolumen
		amv = 30
		// bottle volumes in liters
		bottles = []int{10, 12, 15, 18, 20, 24}
	)

	if len(os.Args) == 2 {
		v, err := strconv.Atoi(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}

		amv = v
	}

	data := make(map[int][]*point)

	for _, v := range bottles {
		points := []*point{}

		if maxDepth < minDepth {
			maxDepth, minDepth = minDepth, maxDepth
		}

		for d := maxDepth; d > minDepth; d -= depthStep {
			p := &point{
				Depth: d,
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
	Depth  int
	MinGas int
}

func (p *point) Calc(amv int) {
	const (
		// time at depth while trying to solve the problem
		tdepth   = 1
		consumer = 2
	)

	// e.g. 40m / 2 = 20m -> 20m / 3 = 6 Steps
	steps := p.Depth / 2 / 3

	// air consumption at depth while trying to solve the problem
	// e.g. 40m: 2 * 30 * 1 * (1 + (40/10)) = 600 l
	air0 := consumer * amv * tdepth * (1 + (p.Depth / 10))

	// air consumption during the ascent to half the depth
	// e.g. 40m->18m: (6 * 3 + 40) / 2.0 = 29
	avDepth1 := float64(steps*3+p.Depth) / 2.0
	// e.g. 40m->18m: (40m - 6 * 3m) / 10.0 = 2.2
	tascent1 := float64(p.Depth-steps*3) / 10.0
	// always round up to the next full minute
	// e.g. 2.2 -> 3.0
	if math.Round(tascent1) < tascent1 {
		tascent1++
	}

	tascent1 = math.Round(tascent1)
	// air consumption
	// e.g. 40m->18m: 2 * 30 * 3.0 * (1 + (29.0 / 10.0)) = 702
	air1 := float64(consumer*amv) * tascent1 * (1 + (avDepth1 / 10.0))

	// air consumption
	// e.g. 18m->0m: (6 * 3) / 2.0 = 9.0
	avDepth2 := float64(steps*3) / 2.0
	// e.g. 18m->0m: 18 / 3 = 6 == steps
	tascent2 := float64(steps)
	// e.g. 18m->0m: 2 * 30 * 3.0 * 6.0 * (1 + (9.0 / 10.0)) = 2052.0
	air2 := float64(consumer*amv) * tascent2 * (1 + (avDepth2 / 10.0))

	fmt.Println(p.Depth, steps, air0, air1, air2)

	p.MinGas = air0 + int(air1) + int(air2)
}

func plotChart(amv int, bottles []int, data map[int][]*point) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	p.Title.Text = fmt.Sprintf("mingas for AMV = %.0f l/min", amv)
	p.X.Label.Text = "depth [m]"
	p.X.Tick.Marker = depthTicks{}
	p.Y.Label.Text = "mingas [bar]"
	p.Y.Tick.Marker = gasTicks{}

	g := plotter.NewGrid()
	p.Add(g)

	lp := []interface{}{}
	for _, v := range bottles {
		lp = append(lp, fmt.Sprintf("%dl", v))
		points := plotter.XYs{}

		for _, p := range data[v] {
			points = append(points, plotter.XY{
				X: float64(p.Depth),
				Y: float64(p.MinGas / v),
			})
		}

		lp = append(lp, points)
	}

	if err := plotutil.AddLinePoints(p, lp...); err != nil {
		return err
	}

	// Save the plot to a PNG file.
	if err := p.Save(xSize, ySize, filename); err != nil {
		panic(err)
	}

	return err
}

// custom ticks for depth axis
type depthTicks struct{}

func (depthTicks) Ticks(min, max float64) []plot.Tick {
	t := []plot.Tick{}

	for i := minDepth; i < maxDepth; i += depthStep {
		t = append(t, plot.Tick{
			Value: float64(i),
			Label: fmt.Sprintf("%d", i),
		})
	}

	return t
}

// custom ticks for gas axis
type gasTicks struct{}

func (gasTicks) Ticks(min, max float64) []plot.Tick {
	t := []plot.Tick{}

	for i := 0; i < 250; i += 10 {
		t = append(t, plot.Tick{
			Value: float64(i),
			Label: fmt.Sprintf("%d", i),
		})
	}

	return t
}
