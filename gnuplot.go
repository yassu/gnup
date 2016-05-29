package gnuplot

import (
	"fmt"
	"io/ioutil"
	"os"
	// "os/exec"
	"strconv"
	"strings"
)

type Plotter struct {
	configures map[string]string
}

func (p *Plotter) Init() {
	p.configures = map[string]string{}
}

func (p *Plotter) Configure(key, val string) {
	p.configures[key] = val
}

func (p *Plotter) GetC(key string) string {
	return p.configures[key]
}

const DefaultFunction2dSplitNum int = 1000

type Function2d struct {
	plotter  Plotter
	splitNum int
	f        func(float64) float64
}

func (fun *Function2d) Init() {
	fun.splitNum = DefaultFunction2dSplitNum
	fun.plotter.configures = map[string]string{
		"_xMin": "-10.0",
		"_xMax": "10.0"}
}

func (fun *Function2d) UpdatePlotter(plotter *Plotter) {
	for key, val := range plotter.configures {
		fun.plotter.configures[key] = val
	}
}

func (fun *Function2d) GetData() [][2]float64 { // TODO: テスト書く
	xMin, _ := strconv.ParseFloat(fun.plotter.configures["_xMin"], 32)
	xMax, _ := strconv.ParseFloat(fun.plotter.configures["_xMax"], 32)
	var sep = float64(xMax-xMin) / float64(fun.splitNum-1)

	var a [][2]float64
	for j := 0; j < fun.splitNum; j++ {
		t := xMin + float64(j)*sep
		y := fun.f(t)
		a = append(a, [2]float64{t, y})
	}
	return a
}

func (fun *Function2d) getGnuData() string {
	var s string
	for _, xs := range fun.GetData() {
		s += fmt.Sprintf("%f %f\n", xs[0], xs[1])
	}
	return s
}

func (fun *Function2d) SetF(_f func(float64) float64) {
	fun.f = _f
}

func (fun Function2d) gnuplot(filename string) string {
	var s = fmt.Sprintf("\"%v\"", filename)
	for key, val := range fun.plotter.configures {
		if !strings.HasPrefix(key, "_") {
			s += fmt.Sprintf(" %v %v", key, val)
		}
	}
	return s
}

func (fun *Function2d) writeIntoGnufile(f os.File) {
	f.WriteString(fun.getGnuData())
}

const DefaultCurve2dSplitNum int = 100

type Curve2d struct {
	plotter  Plotter
	splitNum int
	c        func(float64) [2]float64
}

func (c *Curve2d) Init() {
	c.splitNum = DefaultCurve2dSplitNum
	c.plotter.configures = map[string]string{
		"_tMin": "-10.0",
		"_tMax": "10.0"}
}

func (c *Curve2d) UpdatePlotter(plotter *Plotter) {
	for key, val := range plotter.configures {
		c.plotter.Configure(key, val)
	}
}

func (c *Curve2d) GetData() [][2]float64 { // TODO: test
	tMin, _ := strconv.ParseFloat(c.plotter.configures["_tMin"], 32)
	tMax, _ := strconv.ParseFloat(c.plotter.configures["_tMax"], 32)
	var sep = float64(tMax-tMin) / float64(c.splitNum-1)

	var a [][2]float64
	for j := 0; j < c.splitNum; j++ {
		var t float64 = tMin + float64(j)*sep
		cs := c.c(tMin + t*float64(j))
		a = append(a, [2]float64{cs[0], cs[1]})
	}
	return a
}

func (c *Curve2d) getGnuData() string {
	fmt.Println("=== In getGnuData()")
	fmt.Println(c)
	var s string
	fmt.Println("before of getData")
	fmt.Println(c.GetData())
	fmt.Println("after of getData")
	for _, xs := range c.GetData() {
		fmt.Println(xs)
		s += fmt.Sprintf("%f %f\n", xs[0], xs[1])
	}
	fmt.Println("=== Out of getGnuData()")
	return s
}

func (c *Curve2d) SetC(_c func(float64) [2]float64) {
	c.c = _c
}

func (c Curve2d) gnuplot(fileName string) string {
	var s = fmt.Sprintf("\"%v\" ", fileName)
	for key, val := range c.plotter.configures {
		if !strings.HasPrefix(key, "_") {
			s += fmt.Sprintf(" %v %v", key, val)
		}
	}
	return s
}

// Graph
type Graph2d struct {
	plotter   Plotter
	functions []Function2d
	curves    []Curve2d
}

func (g *Graph2d) Init() {
	g.plotter.configures = map[string]string{}
}

func (g *Graph2d) AppendFunc(f Function2d) {
	g.functions = append(g.functions, f)
}

func (g *Graph2d) AppendCurve(c Curve2d) {
	fmt.Println(c)
	g.curves = append(g.curves, c)
}

func (g Graph2d) writeIntoFile(data string, f *os.File) {
	f.WriteString(data)
}

func (g *Graph2d) UpdatePlotter(plotter *Plotter) {
	fmt.Println("before of Graph2d.UpdatePlotter")
	for key, val := range plotter.configures {
		g.plotter.Configure(key, val)
	}
	fmt.Println("after of Graph2d.UpdatePlotter")
}

func (g Graph2d) exec_gnuplot() {
	// until
}

func (g Graph2d) gnuplot(funcFilenames []string, curveFilenames []string) string {
	var s string

	for key, val := range g.plotter.configures {
		if !strings.HasPrefix(key, "_") {
			if val == "true" {
				s += fmt.Sprintf("set %v;\n", key)
			} else if val == "false" {
				s += fmt.Sprintf("set no%v;\n", key)
			} else {
				s += fmt.Sprintf("set %v %v;\n", key, val)
			}
		}
	}

	s += "plot "
	for j, _ := range g.functions {
		s += g.functions[j].gnuplot(funcFilenames[j]) + ", "
	}
	fmt.Println("curveFilenames = ")
	fmt.Println(curveFilenames)
	for j, _ := range g.curves {
		fmt.Println(j)
		s += g.curves[j].gnuplot(curveFilenames[j])
		if j != len(g.curves)-1 {
			s += ", "
		}
	}
	s += ";\n"
	s += "pause -1;\n"
	return s
}

func (g *Graph2d) Run() {
	tmpDir := os.TempDir() + "/gnuplot.go/"
	// execFilename := tmpDir + "exec.gnu"
	execFilename := "exec.gnu"

	// それぞれのfunctionのdataをtempファイルに書き込む
	// また, それらのファイルの名前を func_filenames []string に格納する
	var funcFilenames []string
	for _, fun := range g.functions {
		file, _ := ioutil.TempFile(tmpDir, "")
		defer func() {
			file.Close()
		}()
		g.writeIntoFile(fun.getGnuData(), file)
		funcFilenames = append(funcFilenames, file.Name())
	}

	fmt.Println("Run.before of curves")
	// それぞれのcurveのdataをtempファイルに書き込む
	// また, それらのファイルの名前を curve_filenames []stringに格納する
	var curveFilenames []string
	for _, c := range g.curves {
		file, _ := ioutil.TempFile(tmpDir, "")
		defer func() {
			file.Close()
		}()
		fmt.Println(c)
		g.writeIntoFile(c.getGnuData(), file)
		curveFilenames = append(curveFilenames, file.Name())
	}
	fmt.Println("Run.after of curves")

	// 実行するgnuplotの実行ファイルをtempファイルに書き込む
	os.Remove(execFilename)
	execFile, _ := os.OpenFile(execFilename, os.O_CREATE|os.O_WRONLY, 0666)
	defer func() {
		execFile.Close()
	}()
	fmt.Println(funcFilenames)
	execFile.WriteString(g.gnuplot(funcFilenames, curveFilenames))
}
