package geofence

import "math"

func crossProduct(o, a, b Point) float64 {
	return (a.Lng-o.Lng)*(b.Lat-o.Lat) - (a.Lat-o.Lat)*(b.Lng-o.Lng)
}

func pointOnSegment(p, a, b Point) bool {
	if math.Abs(crossProduct(a, b, p)) > Epsilon {
		return false
	}
	return p.Lng >= math.Min(a.Lng, b.Lng)-Epsilon &&
		p.Lng <= math.Max(a.Lng, b.Lng)+Epsilon &&
		p.Lat >= math.Min(a.Lat, b.Lat)-Epsilon &&
		p.Lat <= math.Max(a.Lat, b.Lat)+Epsilon
}

func segmentsIntersect(a1, a2, b1, b2 Point) bool {
	c1 := crossProduct(a1, a2, b1)
	c2 := crossProduct(a1, a2, b2)
	c3 := crossProduct(b1, b2, a1)
	c4 := crossProduct(b1, b2, a2)

	if c1*c2 < -Epsilon && c3*c4 < -Epsilon {
		return true
	}

	if pointOnSegment(b1, a1, a2) {
		return true
	}
	if pointOnSegment(b2, a1, a2) {
		return true
	}
	if pointOnSegment(a1, b1, b2) {
		return true
	}
	if pointOnSegment(a2, b1, b2) {
		return true
	}

	return false
}

func distanceToSegment(p, a, b Point) float64 {
	lat1, lng1 := a.ToRadians()
	lat2, lng2 := b.ToRadians()
	latP, lngP := p.ToRadians()

	dLat1 := latP - lat1
	dLng1 := lngP - lng1
	dLat2 := lat2 - lat1
	dLng2 := lng2 - lng1

	dot := dLat1*dLat2 + dLng1*dLng2
	if dot <= 0 {
		return haversineDistance(a, p)
	}

	len2 := dLat2*dLat2 + dLng2*dLng2
	if dot >= len2 {
		return haversineDistance(b, p)
	}

	t := dot / len2
	projLat := lat1 + t*dLat2
	projlng := lng1 + t*dLng2

	return haversineDistance(p, Point{
		Lat: projLat * 180 / math.Pi,
		Lng: projlng * 180 / math.Pi,
	})
}

func haversineDistance(a, b Point) float64 {
	lat1, lng1 := a.ToRadians()
	lat2, lng2 := b.ToRadians()

	dLat := lat2 - lat1
	dLng := lng2 - lng1

	aSqr := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(aSqr), math.Sqrt(1-aSqr))

	return EarthRadius * c
}
