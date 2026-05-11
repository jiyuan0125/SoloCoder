package geolib

import (
	"math"
)

func GreatCirclePoints(center Coordinate, radiusKm float64, numPoints int) []Coordinate {
	if numPoints < 36 {
		numPoints = 36
	}

	phi0 := toRad(center.Lat)
	lambda0 := toRad(center.Lng)
	angularDistance := radiusKm / EarthRadiusKm

	points := make([]Coordinate, numPoints)

	for i := 0; i < numPoints; i++ {
		bearing := 2 * math.Pi * float64(i) / float64(numPoints)

		phi := math.Asin(
			math.Sin(phi0)*math.Cos(angularDistance) +
				math.Cos(phi0)*math.Sin(angularDistance)*math.Cos(bearing))

		lambda := lambda0 + math.Atan2(
			math.Sin(bearing)*math.Sin(angularDistance)*math.Cos(phi0),
			math.Cos(angularDistance)-math.Sin(phi0)*math.Sin(phi))

		lat := toDeg(phi)
		lng := toDeg(lambda)

		lat = clamp(lat, -90, 90)
		lng = normalizeLng(lng)

		points[i] = Coordinate{
			Lat: lat,
			Lng: lng,
		}
	}

	return points
}

func GreatCirclePointsEllipsoid(center Coordinate, radiusKm float64, numPoints int) []Coordinate {
	if numPoints < 36 {
		numPoints = 36
	}

	a := EarthEquatorialRadiusKm
	b := EarthPolarRadiusKm
	f := EarthFlattening

	phi1 := toRad(center.Lat)
	lambda1 := toRad(center.Lng)

	tanU1 := (1 - f) * math.Tan(phi1)
	cosU1 := 1 / math.Sqrt(1+tanU1*tanU1)
	sinU1 := tanU1 * cosU1

	points := make([]Coordinate, numPoints)

	for i := 0; i < numPoints; i++ {
		bearing := 2 * math.Pi * float64(i) / float64(numPoints)

		sinAlpha := cosU1 * math.Sin(bearing)
		cosSqAlpha := 1 - sinAlpha*sinAlpha
		uSq := cosSqAlpha * (a*a - b*b) / (b * b)

		A := 1 + uSq/16384*(4096+uSq*(-768+uSq*(320-175*uSq)))
		B := uSq / 1024 * (256 + uSq*(-128+uSq*(74-47*uSq)))

		s := radiusKm
		sigma := s / (b * A)
		sigmaP := 2 * math.Pi

		var sinSigma, cosSigma, cos2SigmaM float64
		for math.Abs(sigma-sigmaP) > 1e-12 {
			cos2SigmaM = math.Cos(2*math.Asin(math.Sin(sigma)/2)*2 + sigma)
			sinSigma = math.Sin(sigma)
			cosSigma = math.Cos(sigma)

			deltaSigma := B * sinSigma * (cos2SigmaM + B/4*
				(cosSigma*(-1+2*cos2SigmaM*cos2SigmaM) -
					B/6*cos2SigmaM*(-3+4*sinSigma*sinSigma)*(-3+4*cos2SigmaM*cos2SigmaM)))
			sigmaP = sigma
			sigma = s/(b*A) + deltaSigma
		}

		tmp := sinU1*sinSigma - cosU1*cosSigma*math.Cos(bearing)
		phi2 := math.Atan2(sinU1*cosSigma+cosU1*sinSigma*math.Cos(bearing),
			(1-f)*math.Sqrt(sinAlpha*sinAlpha+tmp*tmp))
		lambda := math.Atan2(sinSigma*math.Sin(bearing),
			cosU1*cosSigma-sinU1*sinSigma*math.Cos(bearing))
		C := f / 16 * cosSqAlpha * (4 + f*(4-3*cosSqAlpha))
		L := lambda - (1-C)*f*sinAlpha*
			(sigma + C*sinSigma*(cos2SigmaM+C*cosSigma*(-1+2*cos2SigmaM*cos2SigmaM)))
		lambda2 := lambda1 + L

		lat := toDeg(phi2)
		lng := toDeg(lambda2)

		lat = clamp(lat, -90, 90)
		lng = normalizeLng(lng)

		points[i] = Coordinate{
			Lat: lat,
			Lng: lng,
		}
	}

	return points
}
