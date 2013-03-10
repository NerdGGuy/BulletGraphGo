// bulletgraph - bullet graphs (Design Specification http://www.perceptualedge.com/articles/misc/Bullet_Graph_Design_Spec.pdf)
package bulletgraph

import (
	"fmt"
	"github.com/ajstarks/svgo"
	"strconv"
	"strings"
)

type Bulletgraph struct {
	Top   int     `xml:"top,attr"`
	Left  int     `xml:"left,attr"`
	Right int     `xml:"right,attr"`
	Title string  `xml:"title,attr"`
	Data  []Bdata `xml:"bdata"`
	Note  []Note  `xml:"note"`
	Flags Options
}

type Bdata struct {
	Title    string  `xml:"title,attr"`
	Subtitle string  `xml:"subtitle,attr"`
	Scale    string  `xml:"scale,attr"`
	Qmeasure string  `xml:"qmeasure,attr"`
	Cmeasure float64 `xml:"cmeasure,attr"`
	Measure  float64 `xml:"measure,attr"`
}

type Note struct {
	Text string `xml:",chardata"`
}

type Options struct {
	Width, Height, Fontsize, Barheight, Gutter     int
	Bgcolor, Barcolor, Datacolor, Compcolor, Title string
	Showtitle, Circlemark                          bool
}

func New(canvas *svg.SVG) *Bulletgraph {
	bg := Bulletgraph{}
	bg.Flags.Bgcolor = "white"             //background color
	bg.Flags.Barcolor = "rgb(200,200,200)" //bar color
	bg.Flags.Datacolor = "darkgray"        //data color
	bg.Flags.Compcolor = "black"           //comparative color
	bg.Flags.Width = 1024                  //width
	bg.Flags.Height = 800                  //height 
	bg.Flags.Barheight = 48                //bar height
	bg.Flags.Gutter = 30                   //gutter
	bg.Flags.Fontsize = 18                 //fontsize (px)
	//bg.Flags.Circlemark 				   //circle mark
	//bg.Flags.Showtitle 				   //show title
	//bg.Flags.Title 					   //title
	return &bg
}

// Drawbg draws the bullet graph
func (bg *Bulletgraph) Drawbg(canvas *svg.SVG) {
	qmheight := bg.Flags.Barheight / 3

	if bg.Left == 0 {
		bg.Left = 250
	}
	if bg.Right == 0 {
		bg.Right = 50
	}
	if bg.Top == 0 {
		bg.Top = 50
	}
	if len(bg.Flags.Title) > 0 {
		bg.Title = bg.Flags.Title
	}

	maxwidth := bg.Flags.Width - (bg.Left + bg.Right)
	x := bg.Left
	y := bg.Top
	scalesep := 4
	tx := x - bg.Flags.Fontsize

	canvas.Title(bg.Title)
	// for each data element...
	for _, v := range bg.Data {

		// extract the data from the XML attributes
		sc := strings.Split(v.Scale, ",")
		qm := strings.Split(v.Qmeasure, ",")

		// you must have min,max,increment for the scale, at least one qualitative measure
		if len(sc) != 3 || len(qm) < 1 {
			continue
		}
		// get the qualitative measures
		qmeasures := make([]float64, len(qm))
		for i, q := range qm {
			qmeasures[i], _ = strconv.ParseFloat(q, 64)
		}
		scalemin, _ := strconv.ParseFloat(sc[0], 64)
		scalemax, _ := strconv.ParseFloat(sc[1], 64)
		scaleincr, _ := strconv.ParseFloat(sc[2], 64)

		// label the graph
		canvas.Text(tx, y+bg.Flags.Barheight/3, fmt.Sprintf("%s (%g)", v.Title, v.Measure), "text-anchor:end;font-weight:bold")
		canvas.Text(tx, y+(bg.Flags.Barheight/3)+bg.Flags.Fontsize, v.Subtitle, "text-anchor:end;font-size:75%")

		// draw the scale
		scfmt := "%g"
		if fraction(scaleincr) > 0 {
			scfmt = "%.1f"
		}
		canvas.Gstyle("text-anchor:middle;font-size:75%")
		for sc := scalemin; sc <= scalemax; sc += scaleincr {
			scx := vmap(sc, scalemin, scalemax, 0, float64(maxwidth))
			canvas.Text(x+int(scx), y+scalesep+bg.Flags.Barheight+bg.Flags.Fontsize/2, fmt.Sprintf(scfmt, sc))
		}
		canvas.Gend()

		// draw the qualitative measures
		canvas.Gstyle("fill-opacity:0.5;fill:" + bg.Flags.Barcolor)
		canvas.Rect(x, y, maxwidth, bg.Flags.Barheight)
		for _, q := range qmeasures {
			qbarlength := vmap(q, scalemin, scalemax, 0, float64(maxwidth))
			canvas.Rect(x, y, int(qbarlength), bg.Flags.Barheight)
		}
		canvas.Gend()

		// draw the measure and the comparative measure
		barlength := int(vmap(v.Measure, scalemin, scalemax, 0, float64(maxwidth)))
		canvas.Rect(x, y+qmheight, barlength, qmheight, "fill:"+bg.Flags.Datacolor)
		cmx := int(vmap(v.Cmeasure, scalemin, scalemax, 0, float64(maxwidth)))
		if bg.Flags.Circlemark {
			canvas.Circle(x+cmx, y+bg.Flags.Barheight/2, bg.Flags.Barheight/6, "fill-opacity:0.3;fill:"+bg.Flags.Compcolor)
		} else {
			cbh := bg.Flags.Barheight / 4
			canvas.Line(x+cmx, y+cbh, x+cmx, y+bg.Flags.Barheight-cbh, "stroke-width:3;stroke:"+bg.Flags.Compcolor)
		}

		y += bg.Flags.Barheight + bg.Flags.Gutter // adjust vertical position for the next iteration
	}
	// if requested, place the title below the last bar
	if bg.Flags.Showtitle && len(bg.Title) > 0 {
		y += bg.Flags.Fontsize * 2
		canvas.Text(bg.Left, y, bg.Title, "text-anchor:start;font-size:200%")
	}

	if len(bg.Note) > 0 {
		canvas.Gstyle("font-size:100%;text-anchor:start")
		y += bg.Flags.Fontsize * 2
		leading := 3
		for _, note := range bg.Note {
			canvas.Text(bg.Left, y, note.Text)
			y += bg.Flags.Fontsize + leading
		}
		canvas.Gend()
	}
}

//vmap maps one interval to another
func vmap(value float64, low1 float64, high1 float64, low2 float64, high2 float64) float64 {
	return low2 + (high2-low2)*(value-low1)/(high1-low1)
}

// fraction returns the fractions portion of a floating point number
func fraction(n float64) float64 {
	i := int(n)
	return n - float64(i)
}
