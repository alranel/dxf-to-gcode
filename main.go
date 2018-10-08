package main

import (
    "bufio"
    "github.com/rpaloschi/dxf-go/core"
    "github.com/rpaloschi/dxf-go/document"
    "github.com/rpaloschi/dxf-go/entities"
    "flag"
    "fmt"
    "log"
    "math"
    "os"
    "strings"
)

type Polyline []core.Point

func main() {
    // parse arguments
    Eptr := flag.Float64("E", 1, "extrusion per mm of travel")
    Fptr := flag.Float64("F", 0, "speed")
    centerXptr := flag.Float64("center-x", 0, "X of print area center")
    centerYptr := flag.Float64("center-y", 0, "Y of print area center")
    flag.Parse()
    if flag.Arg(0) == "" {
        fmt.Println(`Usage: dxf-to-gcode -E 1.2 file.dxf`)
        os.Exit(1)
    }
    
    // open the DXF document and start parsing it
    file, err := os.Open(flag.Arg(0)) 
	if err != nil {
		log.Fatal(err)
	}
	doc, err := document.DxfDocumentFromStream(file)
	if err != nil {
		log.Fatal(err)
	}
    
    // collect polylines and check print bounds
    var polylines []Polyline
    min := core.Point{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}
    max := core.Point{math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64}
    for _, entity := range doc.Entities.Entities {
        if polyline, ok := entity.(*entities.Polyline); ok {
            p := NewPolyline(polyline.Vertices)
            p.UpdateBounds(&min, &max)
            polylines = append(polylines, p)
        } else if _, ok := entity.(*entities.Spline); ok {
            fmt.Println("This tool does not yet convert SPLINEs!")
        }
    }
    
    // calculate shift
    shift := core.Point{
        *centerXptr - (max.X + min.X)/2,
        *centerYptr - (max.Y + min.Y)/2,
        0,
    }
    
    // open output file
    outFile := flag.Arg(0) + ".gcode"
    f, err := os.Create(outFile)
    defer f.Close()
    w := bufio.NewWriter(f)
    
    // prepend a few standard commands
    w.WriteString(`G21 ;metric values
G90 ;absolute positioning
M82 ;set extruder to absolute mode
`)
    
    // prepend speed, if any
    if *Fptr > 0 {
        fmt.Fprintf(w, "G1 F%f\n", *Fptr)
    }
    
    // generate G-code
    gcode := GCodeWriter{ E_per_mm: *Eptr }
    for _, p := range polylines {
        p.Translate(shift)
        w.WriteString(gcode.ExtrudePolyline(p))
    }
    
    // print some info
    fmt.Printf("G-code written to %s\n", outFile)
    fmt.Println("PRINT INFO:")
    fmt.Printf("X min: %f; X max: %f\n", min.X, max.X)
    fmt.Printf("Y min: %f; Y max: %f\n", min.Y, max.Y)
    fmt.Printf("Z min: %f; Z max: %f\n", min.Z, max.Z)
    
    w.Flush()
}

func NewPolyline(vv entities.VertexSlice) (Polyline) {
    pp := make([]core.Point, len(vv))
    for i, v := range vv {
        pp[i] = v.Location
    }
    return pp
}

func (pp *Polyline) UpdateBounds(min *core.Point, max *core.Point) {
    for _, v := range *pp {
        min.X = math.Min(min.X, v.X)
        min.Y = math.Min(min.Y, v.Y)
        min.Z = math.Min(min.Z, v.Z)
        max.X = math.Max(max.X, v.X)
        max.Y = math.Max(max.Y, v.Y)
        max.Z = math.Max(max.Z, v.Z)
    }
}

func (pp *Polyline) Translate(shift core.Point) {
    for i := range *pp {
        (*pp)[i].X += shift.X
        (*pp)[i].Y += shift.Y
        (*pp)[i].Z += shift.Z
    }
}

type GCodeWriter struct {
    E_per_mm float64
    cur core.Point
    E float64
}

func (gcode *GCodeWriter) TravelTo(to core.Point) string {
    gcode.cur = to
    return fmt.Sprintf("G1 X%f Y%f Z%f\n", gcode.cur.X, gcode.cur.Y, gcode.cur.Z)
}

func (gcode *GCodeWriter) ExtrudeTo(to core.Point) string {
    gcode.E += distance_to(gcode.cur, to) * gcode.E_per_mm
    gcode.cur = to
    return fmt.Sprintf("G1 X%f Y%f Z%f E%f\n", gcode.cur.X, gcode.cur.Y, gcode.cur.Z, gcode.E)
}

func (gcode *GCodeWriter) ExtrudePolyline(pp Polyline) string {
    if len(pp) == 0 {
        return ""
    }
    
    // if last vertex is lower than the first one, reverse vertices
    // so that we build upwards
    if pp[0].Z > pp[len(pp)-1].Z {
        for i, j := 0, len(pp)-1; i < j; i, j = i+1, j-1 {
            pp[i], pp[j] = pp[j], pp[i]
        }
    }
    
    var out strings.Builder
    for i, p := range pp {
        if (i == 0) {
            out.WriteString(gcode.TravelTo(p))
        } else {
            out.WriteString(gcode.ExtrudeTo(p))
        }
    }
    return out.String()
}

func distance_to(from core.Point, to core.Point) (float64) {
    dX := to.X - from.Y
    dY := to.Y - from.Y
    dZ := to.Z - from.Z
    return math.Sqrt(math.Pow(dX, 2) + math.Pow(dY, 2) + math.Pow(dZ, 2))
}
