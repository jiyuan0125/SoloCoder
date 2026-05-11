package s2

import (
	"math"
)

const (
	MaxLevel = 30
	NumFaces = 6
)

type LatLng struct {
	Lat float64
	Lng float64
}

type Point struct {
	X float64
	Y float64
	Z float64
}

type CellID uint64

func (id CellID) Face() int {
	return int((id >> 61) & 0x7)
}

func (id CellID) Level() int {
	if id == 0 {
		return 0
	}
	id64 := uint64(id)
	for level := MaxLevel; level >= 0; level-- {
		shift := 2 * (MaxLevel - level)
		if (id64 & (uint64(1) << shift)) != 0 {
			return level
		}
	}
	return 0
}

func (id CellID) IsValid() bool {
	face := id.Face()
	if face < 0 || face >= NumFaces {
		return false
	}
	if id == 0 {
		return false
	}
	level := id.Level()
	if level < 0 || level > MaxLevel {
		return false
	}
	for l := 0; l < level; l++ {
		shift := 2 * (MaxLevel - l)
		if (uint64(id) & (uint64(1) << shift)) != 0 {
			return false
		}
	}
	return true
}

func (id CellID) Parent(level int) CellID {
	if level < 0 || level > MaxLevel {
		return 0
	}
	currentLevel := id.Level()
	if level >= currentLevel {
		return id
	}
	clearBits := 2 * (MaxLevel - level)
	id64 := uint64(id)
	result := (id64 >> clearBits) << clearBits
	result |= uint64(1) << clearBits
	return CellID(result)
}

func (id CellID) ParentAtLevel(level int) CellID {
	return id.Parent(level)
}

func (id CellID) ImmediateParent() CellID {
	return id.Parent(id.Level() - 1)
}

func (id CellID) Contains(other CellID) bool {
	if id == 0 || other == 0 {
		return false
	}
	if id.Face() != other.Face() {
		return false
	}
	if id.Level() > other.Level() {
		return false
	}
	return other.Parent(id.Level()) == id
}

func (id CellID) IsContainedBy(other CellID) bool {
	return id.Contains(other)
}

func (id CellID) Intersects(other CellID) bool {
	if id == 0 || other == 0 {
		return false
	}
	if id.Face() != other.Face() {
		return false
	}
	level := id.Level()
	if other.Level() < level {
		level = other.Level()
	}
	return id.Parent(level) == other.Parent(level)
}

func LatLngFromPoint(p Point) LatLng {
	return LatLng{
		Lat: math.Atan2(p.Z, math.Sqrt(p.X*p.X+p.Y*p.Y)),
		Lng: math.Atan2(p.Y, p.X),
	}
}

func PointFromLatLng(ll LatLng) Point {
	lat := ll.Lat
	lng := ll.Lng
	cosLat := math.Cos(lat)
	return Point{
		X: cosLat * math.Cos(lng),
		Y: cosLat * math.Sin(lng),
		Z: math.Sin(lat),
	}
}

func CellIDFromPoint(p Point) CellID {
	face := Face(p)
	u, v := FaceUV(face, p)
	ij := [2]int{
		STToIJ(UvToST(u)),
		STToIJ(UvToST(v)),
	}
	return CellIDFromFaceIJ(face, ij[0], ij[1])
}

func CellIDFromLatLng(ll LatLng) CellID {
	return CellIDFromPoint(PointFromLatLng(ll))
}

func Face(p Point) int {
	abs := [3]float64{
		math.Abs(p.X),
		math.Abs(p.Y),
		math.Abs(p.Z),
	}
	face := 0
	maxVal := abs[0]
	if abs[1] > maxVal {
		face = 1
		maxVal = abs[1]
	}
	if abs[2] > maxVal {
		face = 2
	}
	if face == 0 && p.X < 0 {
		face += 3
	}
	if face == 1 && p.Y < 0 {
		face += 3
	}
	if face == 2 && p.Z < 0 {
		face += 3
	}
	return face
}

func FaceUV(face int, p Point) (float64, float64) {
	var u, v float64
	switch face {
	case 0:
		u = p.Y / p.X
		v = p.Z / p.X
	case 1:
		u = -p.X / p.Y
		v = p.Z / p.Y
	case 2:
		u = -p.X / p.Z
		v = -p.Y / p.Z
	case 3:
		u = p.Z / p.X
		v = p.Y / p.X
	case 4:
		u = p.Z / p.Y
		v = -p.X / p.Y
	case 5:
		u = -p.X / p.Z
		v = p.Y / p.Z
	}
	return u, v
}

func UvToST(u float64) float64 {
	if u >= 0 {
		return 0.5 * math.Sqrt(1+3*u)
	}
	return 1 - 0.5*math.Sqrt(1-3*u)
}

func STToIJ(s float64) int {
	const maxSize = 1 << MaxLevel
	k := math.Floor(s * float64(maxSize))
	return int(math.Max(0, math.Min(float64(maxSize-1), k)))
}

func IJToST(i int) float64 {
	const maxSize = 1 << MaxLevel
	return float64(i) / float64(maxSize)
}

func STToUv(s float64) float64 {
	if s >= 0.5 {
		return (1.0/3.0)*(4*s*s - 1)
	}
	return (1.0/3.0)*(1 - 4*(1-s)*(1-s))
}

func FaceUvToPoint(face int, u, v float64) Point {
	var p Point
	switch face {
	case 0:
		p = Point{1, u, v}
	case 1:
		p = Point{-u, 1, v}
	case 2:
		p = Point{-u, -v, 1}
	case 3:
		p = Point{-1, -v, -u}
	case 4:
		p = Point{v, -1, -u}
	case 5:
		p = Point{v, u, -1}
	}
	return p
}

func FaceIJToSTFaceUV(face, i, j int) Point {
	u := STToUv(IJToST(i))
	v := STToUv(IJToST(j))
	return FaceUvToPoint(face, u, v)
}

func CellIDFromFaceIJ(face, i, j int) CellID {
	var id uint64
	id |= uint64(face) << 61
	for k := 0; k < MaxLevel; k++ {
		mask := 1 << (MaxLevel - 1 - k)
		bitI := 0
		if (i & mask) != 0 {
			bitI = 1
		}
		bitJ := 0
		if (j & mask) != 0 {
			bitJ = 1
		}
		shift := 60 - 2*k
		id |= uint64((bitI<<1)|bitJ) << shift
	}
	id |= 1
	return CellID(id)
}

func (id CellID) FaceIJ() (face, i, j int) {
	face = id.Face()
	i = 0
	j = 0
	level := id.Level()
	for k := 0; k < level; k++ {
		shift := 60 - 2*k
		bits := (id >> shift) & 0x3
		i = (i << 1) | int((bits>>1)&1)
		j = (j << 1) | int(bits&1)
	}
	return
}

func (id CellID) Point() Point {
	face, i, j := id.FaceIJ()
	level := id.Level()
	scale := 1 << (MaxLevel - level)
	centerI := (2*i + 1) * scale
	centerJ := (2*j + 1) * scale
	u := STToUv(SIToST(centerI))
	v := STToUv(SIToST(centerJ))
	return FaceUvToPoint(face, u, v)
}

func SIToST(si int) float64 {
	return float64(si) / float64(2*(1<<MaxLevel))
}

func (id CellID) LatLng() LatLng {
	return LatLngFromPoint(id.Point())
}

type Rect struct {
	LatLo float64
	LatHi float64
	LngLo float64
	LngHi float64
}

func (r Rect) IsFullLat() bool {
	return r.LatLo <= -math.Pi/2 && r.LatHi >= math.Pi/2
}

func (r Rect) IsFullLng() bool {
	return r.LngHi-r.LngLo >= 2*math.Pi
}

func (r Rect) Intersects(other Rect) bool {
	if r.LatHi < other.LatLo || other.LatHi < r.LatLo {
		return false
	}
	if r.IsFullLng() || other.IsFullLng() {
		return true
	}
	return lngRangeIntersects(r.LngLo, r.LngHi, other.LngLo, other.LngHi)
}

func (r Rect) Contains(other Rect) bool {
	if r.LatHi < other.LatHi || r.LatLo > other.LatLo {
		return false
	}
	if r.IsFullLng() {
		return true
	}
	if other.IsFullLng() {
		return false
	}
	return lngRangeContains(r.LngLo, r.LngHi, other.LngLo, other.LngHi)
}

func lngRangeIntersects(aLo, aHi, bLo, bHi float64) bool {
	if aHi >= aLo && bHi >= bLo {
		return aHi >= bLo && bHi >= aLo
	}
	return aHi >= bLo || bHi >= aLo || aLo <= bHi || bLo <= aHi
}

func lngRangeContains(aLo, aHi, bLo, bHi float64) bool {
	if aHi >= aLo {
		if bHi >= bLo {
			return aLo <= bLo && aHi >= bHi
		}
		return false
	}
	if bHi >= bLo {
		return aLo <= bLo || aHi >= bHi
	}
	return aLo <= bLo && aHi >= bHi
}

func (id CellID) RectBound() Rect {
	level := id.Level()
	face, i, j := id.FaceIJ()
	scale := 1 << (MaxLevel - level)
	iScaled := i << (MaxLevel - level)
	jScaled := j << (MaxLevel - level)
	lo := FaceIJToSTFaceUV(face, iScaled, jScaled)
	hi := FaceIJToSTFaceUV(face, iScaled+scale, jScaled+scale)
	llLo := LatLngFromPoint(lo)
	llHi := LatLngFromPoint(hi)
	latLo := math.Min(llLo.Lat, llHi.Lat)
	latHi := math.Max(llLo.Lat, llHi.Lat)
	lngLo := math.Min(llLo.Lng, llHi.Lng)
	lngHi := math.Max(llLo.Lng, llHi.Lng)
	return Rect{
		LatLo: latLo,
		LatHi: latHi,
		LngLo: lngLo,
		LngHi: lngHi,
	}
}

func DegToRad(deg float64) float64 {
	return deg * math.Pi / 180.0
}

func RadToDeg(rad float64) float64 {
	return rad * 180.0 / math.Pi
}
