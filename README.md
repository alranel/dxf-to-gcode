# dxf-to-gcode

This tool converts a DXF file containing one or more polylines into a G-code file suitable for simple 3D printing. It performs the following steps:

* polylines are converted into G-code G1 instructions (starting from the endpoint with the lowest Z coordinate)
* for each move, the E value is calculated according to the configured unit value
* the print is centered around the given point (0,0 by default)
* if speed is supplied, a `G1 Fx` command is generated at the beginning

It does not currently support splines, so make sure you convert them to splines.

This tool is written in Go so it can be executed without installing any dependency. You can download the compiled binaries for Mac and Windows in the Releases section of this GitHub repository.

It should be actually included in Slic3r in order to benefit from the algorithms in libslic3r.

## Usage

```
./dxf-to-gcode -E 5 path/to/my.dxf
```

The value supplied to `-E` is the amount of extrusion to be generated for 1mm of linear move of the printing head. Whether this is filament length, steps/mm or a volume depends entirely on what your firmware expects for the "E" G-code value.

## Author

Alessandro Ranellucci
