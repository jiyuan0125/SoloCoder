package geolib

import (
	"math"
)

const (
	EarthRadiusKm        = 6371.0
	EarthEquatorialRadiusKm = 6378.137
	EarthPolarRadiusKm      = 6356.752
	EarthFlattening         = 1.0 / 298.257223563
	degToRad              = math.Pi / 180.0
	radToDeg              = 180.0 / math.Pi
)

func toRad(deg float64) float64 {
	return deg * degToRad
}

func toDeg(rad float64) float64 {
	return rad * radToDeg
}

func clamp(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

func HaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	phi1 := toRad(lat1)
	phi2 := toRad(lat2)
	deltaPhi := toRad(lat2 - lat1)
	deltaLambda := toRad(lng2 - lng1)

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
		math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)

	clampedArg := clamp(a, 0.0, 1.0)
	c := 2 * math.Atan2(math.Sqrt(clampedArg), math.Sqrt(1-clampedArg))

	return EarthRadiusKm * c
}

func VincentyApproxDistance(lat1, lng1, lat2, lng2 float64) float64 {
	phi1 := toRad(lat1)
	phi2 := toRad(lat2)
	lambda := toRad(math.Abs(lng2 - lng1))
	f := EarthFlattening

	a := EarthEquatorialRadiusKm
	b := EarthPolarRadiusKm

	U1 := math.Atan((1 - f) * math.Tan(phi1))
	U2 := math.Atan((1 - f) * math.Tan(phi2))
	sinU1 := math.Sin(U1)
	cosU1 := math.Cos(U1)
	sinU2 := math.Sin(U2)
	cosU2 := math.Cos(U2)

	lambdaIter := lambda
	lambdaPrev := 0.0
	var sinLambda, cosLambda, sinSigma, cosSigma, sigma, sinAlpha, cosSqAlpha, cos2SigmaM, C float64

	for i := 0; i < 100 && math.Abs(lambdaIter-lambdaPrev) > 1e-12; i++ {
		lambdaPrev = lambdaIter
		sinLambda = math.Sin(lambdaIter)
		cosLambda = math.Cos(lambdaIter)

		sinSigma = math.Sqrt((cosU2*sinLambda)*(cosU2*sinLambda) +
			(cosU1*sinU2-sinU1*cosU2*cosLambda)*(cosU1*sinU2-sinU1*cosU2*cosLambda))

		if sinSigma == 0 {
			return 0.0
		}

		cosSigma = sinU1*sinU2 + cosU1*cosU2*cosLambda
		sigma = math.Atan2(sinSigma, cosSigma)

		sinAlpha = (cosU1 * cosU2 * sinLambda) / sinSigma
		cosSqAlpha = 1 - sinAlpha*sinAlpha

		if cosSqAlpha == 0 {
			cos2SigmaM = 0
		} else {
			cos2SigmaM = cosSigma - 2*sinU1*sinU2/cosSqAlpha
		}

		C = f / 16 * cosSqAlpha * (4 + f*(4-3*cosSqAlpha))

		lambdaIter = lambda + (1-C)*f*sinAlpha*
			(sigma + C*sinSigma*(cos2SigmaM+C*cosSigma*(-1+2*cos2SigmaM*cos2SigmaM)))
	}

	uSq := cosSqAlpha * (a*a - b*b) / (b * b)
	A := 1 + uSq/16384*(4096+uSq*(-768+uSq*(320-175*uSq)))
	B := uSq / 1024 * (256 + uSq*(-128+uSq*(74-47*uSq)))
	deltaSigma := B * sinSigma * (cos2SigmaM + B/4*
		(cosSigma*(-1+2*cos2SigmaM*cos2SigmaM) -
			B/6*cos2SigmaM*(-3+4*sinSigma*sinSigma)*(-3+4*cos2SigmaM*cos2SigmaM)))

	s := b * A * (sigma - deltaSigma)

	return s
}

func Distance(lat1, lng1, lat2, lng2 float64, useEllipsoid bool) (float64, error) {
	if err := ValidateCoordinate(lat1, lng1); err != nil {
		return 0, err
	}
	if err := ValidateCoordinate(lat2, lng2); err != nil {
		return 0, err
	}

	if lat1 == lat2 && lng1 == lng2 {
		return 0, nil
	}

	if useEllipsoid {
		return VincentyApproxDistance(lat1, lng1, lat2, lng2), nil
	}
	return HaversineDistance(lat1, lng1, lat2, lng2), nil
}
