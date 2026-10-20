package tides

import (
	"fmt"
	"math"
	"time"
)

// Constantes de ubicación de Gijón (Puerto del Musel)
const (
	GijonLatitude  = 43.55803 // Grados Norte
	GijonLongitude = -5.69835 // Grados Oeste (negativo)
	GijonTZName    = "Europe/Madrid"
)

var gijonLocation *time.Location

func init() {
	loc, err := time.LoadLocation(GijonTZName)
	if err != nil {
		loc = time.FixedZone("CET", 1*3600)
	}
	gijonLocation = loc
}

// GetGijonLocation devuelve la zona horaria oficial de Gijón
func GetGijonLocation() *time.Location {
	return gijonLocation
}

func degToRad(deg float64) float64 {
	return deg * math.Pi / 180.0
}

func radToDeg(rad float64) float64 {
	return rad * 180.0 / math.Pi
}

func normalizeDeg(deg float64) float64 {
	d := math.Mod(deg, 360.0)
	if d < 0 {
		d += 360.0
	}
	return d
}

// JulianDate calcula el día juliano exacto para un tiempo UTC
func JulianDate(t time.Time) float64 {
	utc := t.UTC()
	y := utc.Year()
	m := int(utc.Month())
	d := float64(utc.Day()) +
		float64(utc.Hour())/24.0 +
		float64(utc.Minute())/1440.0 +
		float64(utc.Second())/86400.0 +
		float64(utc.Nanosecond())/8.64e13

	if m <= 2 {
		y -= 1
		m += 12
	}

	a := math.Floor(float64(y) / 100.0)
	b := 2.0 - a + math.Floor(a/4.0)

	jd := math.Floor(365.25*float64(y+4716)) +
		math.Floor(30.6001*float64(m+1)) +
		d + b - 1524.5
	return jd
}

// JulianCenturies calcula los siglos julianos desde J2000.0
func JulianCenturies(t time.Time) float64 {
	return (JulianDate(t) - 2451545.0) / 36525.0
}

// FundamentalArguments contiene los ángulos astronómicos de Brown / Doodson
type FundamentalArguments struct {
	Tau float64 // Tiempo lunar medio (grados)
	S   float64 // Longitud media de la Luna (grados)
	H   float64 // Longitud media del Sol (grados)
	P   float64 // Longitud del perigeo lunar (grados)
	N   float64 // Longitud del nodo ascendente lunar (grados)
	Ps  float64 // Longitud del perigeo solar (grados)
	T   float64 // Siglos julianos
}

// CalcFundamentalArguments calcula los argumentos fundamentales para una fecha
func CalcFundamentalArguments(t time.Time) FundamentalArguments {
	T := JulianCenturies(t)

	// Longitud media de la Luna
	s := 218.3164477 + 481267.88128*T - 0.0015786*T*T + (T*T*T)/538841.0
	// Longitud media del Sol
	h := 280.46646 + 36000.76983*T + 0.0003032*T*T
	// Longitud media del perigeo lunar
	p := 83.353542 + 4069.0137287*T - 0.01032*T*T - (T*T*T)/80053.0
	// Longitud media del nodo lunar ascendente
	n := 125.04452 - 1934.136261*T + 0.0020708*T*T + (T*T*T)/450000.0
	// Longitud media del perigeo solar
	ps := 282.937348 + 1.71953*T + 0.0004568*T*T

	utc := t.UTC()
	solarHourDeg := (float64(utc.Hour()) + float64(utc.Minute())/60.0 + float64(utc.Second())/3600.0) * 15.0

	s = normalizeDeg(s)
	h = normalizeDeg(h)
	p = normalizeDeg(p)
	n = normalizeDeg(n)
	ps = normalizeDeg(ps)

	// tau = solarHourDeg + h - s
	tau := normalizeDeg(solarHourDeg + h - s)

	return FundamentalArguments{
		Tau: tau,
		S:   s,
		H:   h,
		P:   p,
		N:   n,
		Ps:  ps,
		T:   T,
	}
}

// MoonPhaseInfo contiene información detallada sobre la fase de la Luna
type MoonPhaseInfo struct {
	PhaseAngle  float64 `json:"phase_angle"`  // 0 - 360º
	Fraction    float64 `json:"fraction"`     // 0.0 a 1.0 (iluminación)
	Percent     int     `json:"percent"`      // 0 - 100%
	AgeDays     float64 `json:"age_days"`     // 0 a 29.53 días
	PhaseName   string  `json:"phase_name"`   // Ej: "Luna Llena"
	PhaseIcon   string  `json:"phase_icon"`   // Código/Nombre de icono
	Trend       string  `json:"trend"`        // "Creciente" o "Menguante"
	IsSyzygy    bool    `json:"is_syzygy"`    // Cerca de Nueva o Llena (Mareas vivas)
	IsQuadrature bool   `json:"is_quadrature"`// Cerca de Cuartos (Mareas muertas)
}

// CalcMoonPhase calcula con alta precisión la fase lunar para un instante
func CalcMoonPhase(t time.Time) MoonPhaseInfo {
	args := CalcFundamentalArguments(t)
	// Ángulo de elongación lunar D = S - H
	elong := normalizeDeg(args.S - args.H)

	// Fracción iluminada del disco lunar: (1 - cos(D)) / 2
	fraction := (1.0 - math.Cos(degToRad(elong))) / 2.0
	percent := int(math.Round(fraction * 100.0))
	ageDays := (elong / 360.0) * 29.530588853

	var phaseName string
	var phaseIcon string
	var trend string

	if elong >= 180.0 {
		trend = "Menguante"
	} else {
		trend = "Creciente"
	}

	switch {
	case elong >= 354.0 || elong < 6.0:
		phaseName = "Luna Nueva"
		phaseIcon = "new-moon"
		trend = "Luna Nueva"
	case elong >= 6.0 && elong < 84.0:
		phaseName = "Luna Creciente"
		phaseIcon = "waxing-crescent"
	case elong >= 84.0 && elong < 96.0:
		phaseName = "Cuarto Creciente"
		phaseIcon = "first-quarter"
	case elong >= 96.0 && elong < 174.0:
		phaseName = "Gibosa Creciente"
		phaseIcon = "waxing-gibbous"
	case elong >= 174.0 && elong < 186.0:
		phaseName = "Luna Llena"
		phaseIcon = "full-moon"
		trend = "Luna Llena"
	case elong >= 186.0 && elong < 264.0:
		phaseName = "Gibosa Menguante"
		phaseIcon = "waning-gibbous"
	case elong >= 264.0 && elong < 276.0:
		phaseName = "Cuarto Menguante"
		phaseIcon = "last-quarter"
	default:
		phaseName = "Luna Menguante"
		phaseIcon = "waning-crescent"
	}

	isSyzygy := (elong >= 345.0 || elong <= 15.0) || (elong >= 165.0 && elong <= 195.0)
	isQuadrature := (elong >= 75.0 && elong <= 105.0) || (elong >= 255.0 && elong <= 285.0)

	return MoonPhaseInfo{
		PhaseAngle:   elong,
		Fraction:     fraction,
		Percent:      percent,
		AgeDays:      math.Round(ageDays*10) / 10,
		PhaseName:    phaseName,
		PhaseIcon:    phaseIcon,
		Trend:        trend,
		IsSyzygy:     isSyzygy,
		IsQuadrature: isQuadrature,
	}
}

// SunTimes contiene la información solar de orto, ocaso y horas de luz
type SunTimes struct {
	Sunrise      string `json:"sunrise"`       // "HH:MM"
	Sunset       string `json:"sunset"`        // "HH:MM"
	SolarNoon    string `json:"solar_noon"`    // "HH:MM"
	DawnCivil    string `json:"dawn_civil"`    // Crepúsculo matutino "HH:MM"
	DuskCivil    string `json:"dusk_civil"`    // Crepúsculo vespertino "HH:MM"
	DayLengthSec int    `json:"day_length_sec"`// Segundos de sol
	DayLengthStr string `json:"day_length_str"`// "12h 45m"
}

// CalcSunTimes calcula orto, ocaso y crepúsculo solar en Gijón para un día determinado
func CalcSunTimes(date time.Time) SunTimes {
	loc := GetGijonLocation()
	midday := time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, time.UTC)
	T := JulianCenturies(midday)

	// Anomalía media del Sol
	M := degToRad(normalizeDeg(357.52911 + 35999.05029*T - 0.0001537*T*T))
	// Longitud eclíptica del Sol
	L0 := normalizeDeg(280.46646 + 36000.76983*T + 0.0003032*T*T)
	C := (1.914602-0.004817*T-0.000014*T*T)*math.Sin(M) +
		(0.019993-0.000101*T)*math.Sin(2*M) +
		0.000289*math.Sin(3*M)
	sunTrueLong := normalizeDeg(L0 + C)

	// Oblicuidad de la eclíptica
	eps := degToRad(23.439291 - 0.0130042*T)
	// Declinación solar
	sinDec := math.Sin(eps) * math.Sin(degToRad(sunTrueLong))
	cosDec := math.Cos(math.Asin(sinDec))
	dec := math.Asin(sinDec)

	// Ecuación del tiempo (minutos)
	y := math.Tan(eps / 2.0)
	y *= y
	sin2L0 := math.Sin(2.0 * degToRad(L0))
	cos2L0 := math.Cos(2.0 * degToRad(L0))
	sinM := math.Sin(M)
	eqTime := 4.0 * radToDeg(y*sin2L0-2.0*0.016708634*sinM+4.0*0.016708634*y*sinM*cos2L0-0.5*y*y*math.Sin(4.0*degToRad(L0))-1.25*0.016708634*0.016708634*math.Sin(2.0*M))

	// Mediodía solar UTC en minutos desde las 00:00 UTC
	solarNoonUTC := 720.0 - (GijonLongitude * 4.0) - eqTime

	latRad := degToRad(GijonLatitude)

	// Ángulo horario del sol al amanecer/atardecer (depresión estándar -0.833º para refracción y semidiámetro)
	calcHourAngle := func(zenith float64) (float64, bool) {
		cosHA := (math.Cos(degToRad(zenith)) - math.Sin(latRad)*math.Sin(dec)) / (math.Cos(latRad) * cosDec)
		if cosHA > 1.0 || cosHA < -1.0 {
			return 0, false
		}
		return radToDeg(math.Acos(cosHA)), true
	}

	formatTime := func(utcMinutes float64) string {
		for utcMinutes < 0 {
			utcMinutes += 1440
		}
		for utcMinutes >= 1440 {
			utcMinutes -= 1440
		}
		h := int(utcMinutes) / 60
		m := int(utcMinutes) % 60
		s := int(math.Round((utcMinutes - float64(h*60+m)) * 60))
		tUTC := time.Date(date.Year(), date.Month(), date.Day(), h, m, s, 0, time.UTC)
		tLocal := tUTC.In(loc)
		return tLocal.Format("15:04")
	}

	haStandard, okStd := calcHourAngle(90.833)
	haCivil, okCiv := calcHourAngle(96.0)

	var sunriseStr, sunsetStr, dawnStr, duskStr string
	var dayLengthSec int

	if okStd {
		sunriseUTC := solarNoonUTC - haStandard*4.0
		sunsetUTC := solarNoonUTC + haStandard*4.0
		sunriseStr = formatTime(sunriseUTC)
		sunsetStr = formatTime(sunsetUTC)
		dayLengthSec = int(haStandard * 8.0 * 60.0)
	} else {
		sunriseStr = "--:--"
		sunsetStr = "--:--"
	}

	if okCiv {
		dawnStr = formatTime(solarNoonUTC - haCivil*4.0)
		duskStr = formatTime(solarNoonUTC + haCivil*4.0)
	} else {
		dawnStr = sunriseStr
		duskStr = sunsetStr
	}

	noonStr := formatTime(solarNoonUTC)

	hours := dayLengthSec / 3600
	mins := (dayLengthSec % 3600) / 60
	dayLengthStr := fmt.Sprintf("%dh %02dm", hours, mins)

	return SunTimes{
		Sunrise:      sunriseStr,
		Sunset:       sunsetStr,
		SolarNoon:    noonStr,
		DawnCivil:    dawnStr,
		DuskCivil:    duskStr,
		DayLengthSec: dayLengthSec,
		DayLengthStr: dayLengthStr,
	}
}

// MoonTimes contiene la salida, puesta y culminación de la luna
type MoonTimes struct {
	Moonrise   string `json:"moonrise"`    // "HH:MM" o "--:--" si no sale ese día
	Moonset    string `json:"moonset"`     // "HH:MM"
	Transit    string `json:"transit"`     // Tránsito superior / culminación "HH:MM"
	Underfoot  string `json:"underfoot"`   // Tránsito inferior / nadir "HH:MM"
}

// CalcMoonTimes calcula el orto, ocaso y pasos de meridiano de la luna en Gijón
func CalcMoonTimes(date time.Time) MoonTimes {
	loc := GetGijonLocation()
	// Muestreamos la posición lunar a lo largo de las 24 horas del día en pasos de 10 minutos
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)

	var riseTime, setTime, transitTime, underfootTime string
	var prevAlt float64 = -999

	maxAlt := -999.0
	minAlt := 999.0
	var tMax, tMin time.Time

	// Búsqueda de eventos en la jornada (28 horas para asegurar detección)
	for min := 0; min <= 28*60; min += 5 {
		currTime := start.Add(time.Duration(min) * time.Minute)
		alt := calcMoonAltitude(currTime)

		if prevAlt != -999 {
			// Cruce de horizonte (altura lunar = +0.125º por refracción y paralaje)
			if prevAlt < 0.125 && alt >= 0.125 && riseTime == "" && currTime.Day() == date.Day() {
				riseTime = currTime.In(loc).Format("15:04")
			}
			if prevAlt >= 0.125 && alt < 0.125 && setTime == "" && currTime.Day() == date.Day() {
				setTime = currTime.In(loc).Format("15:04")
			}
		}

		if currTime.Day() == date.Day() {
			if alt > maxAlt {
				maxAlt = alt
				tMax = currTime
			}
			if alt < minAlt {
				minAlt = alt
				tMin = currTime
			}
		}
		prevAlt = alt
	}

	if !tMax.IsZero() {
		transitTime = tMax.In(loc).Format("15:04")
	}
	if !tMin.IsZero() {
		underfootTime = tMin.In(loc).Format("15:04")
	}

	if riseTime == "" {
		riseTime = "--:--"
	}
	if setTime == "" {
		setTime = "--:--"
	}
	if transitTime == "" {
		transitTime = "--:--"
	}
	if underfootTime == "" {
		underfootTime = "--:--"
	}

	return MoonTimes{
		Moonrise:  riseTime,
		Moonset:   setTime,
		Transit:   transitTime,
		Underfoot: underfootTime,
	}
}

// calcMoonAltitude calcula la altura topocéntrica aproximada de la Luna en grados para Gijón
func calcMoonAltitude(t time.Time) float64 {
	args := CalcFundamentalArguments(t)

	// Coordenadas eclípticas lunares principales
	L := args.S + 6.289*math.Sin(degToRad(args.S-args.P))
	B := 5.128 * math.Sin(degToRad(args.S-args.N))

	// Oblicuidad
	eps := degToRad(23.439)
	lRad := degToRad(L)
	bRad := degToRad(B)

	// Coordenadas ecuatoriales lunares: ascensión recta y declinación
	sinDec := math.Sin(bRad)*math.Cos(eps) + math.Cos(bRad)*math.Sin(eps)*math.Sin(lRad)
	dec := math.Asin(sinDec)
	y := math.Sin(lRad)*math.Cos(eps) - math.Tan(bRad)*math.Sin(eps)
	x := math.Cos(lRad)
	ra := normalizeDeg(radToDeg(math.Atan2(y, x)))

	// Tiempo sidéreo local en Gijón
	jd := JulianDate(t)
	d := jd - 2451545.0
	gmst := normalizeDeg(280.46061837 + 360.98564736629*d)
	lmst := normalizeDeg(gmst + GijonLongitude)

	// Ángulo horario lunar H
	h := normalizeDeg(lmst - ra)

	// Altura sobre el horizonte
	latRad := degToRad(GijonLatitude)
	hRad := degToRad(h)
	sinAlt := math.Sin(latRad)*math.Sin(dec) + math.Cos(latRad)*math.Cos(dec)*math.Cos(hRad)
	alt := radToDeg(math.Asin(sinAlt))

	return alt
}

// SolunarPeriod representa un periodo solunar de actividad (Teoría de John Alden Knight)
type SolunarPeriod struct {
	Type        string `json:"type"`        // "Mayor" o "Menor"
	Name        string `json:"name"`        // "Periodo Mayor 1", etc.
	Description string `json:"description"` // "Tránsito Lunar (Paso por el cenit)"
	StartTime   string `json:"start_time"`  // "HH:MM"
	EndTime     string `json:"end_time"`    // "HH:MM"
	PeakTime    string `json:"peak_time"`   // "HH:MM"
	RatingStars int    `json:"rating_stars"`// 1 a 4 estrellas
}

// SolunarDay contiene el resumen solunar de la jornada
type SolunarDay struct {
	ActivityLevel string          `json:"activity_level"` // "Excelente", "Alta", "Media", "Baja"
	OverallRating int             `json:"overall_rating"` // 1 a 4 estrellas / peces
	Summary       string          `json:"summary"`
	Periods       []SolunarPeriod `json:"periods"`
}

// CalcSolunarDay calcula los periodos de pesca/actividad solunar para una fecha en Gijón
func CalcSolunarDay(date time.Time, moonTimes MoonTimes, phase MoonPhaseInfo, tideCoeff int) SolunarDay {
	loc := GetGijonLocation()
	parseTime := func(str string) (time.Time, bool) {
		if str == "" || str == "--:--" {
			return time.Time{}, false
		}
		t, err := time.ParseInLocation("2006-01-02 15:04", date.Format("2006-01-02")+" "+str, loc)
		return t, err == nil
	}

	var periods []SolunarPeriod

	// Periodo Mayor 1: Tránsito superior (±60 min)
	if t, ok := parseTime(moonTimes.Transit); ok {
		periods = append(periods, SolunarPeriod{
			Type:        "Mayor",
			Name:        "Periodo Mayor 1",
			Description: "Paso de la Luna por el cenit (máxima atracción gravitatoria)",
			StartTime:   t.Add(-60 * time.Minute).Format("15:04"),
			EndTime:     t.Add(60 * time.Minute).Format("15:04"),
			PeakTime:    t.Format("15:04"),
			RatingStars: 4,
		})
	}

	// Periodo Mayor 2: Tránsito inferior / Nadir (±60 min)
	if t, ok := parseTime(moonTimes.Underfoot); ok {
		periods = append(periods, SolunarPeriod{
			Type:        "Mayor",
			Name:        "Periodo Mayor 2",
			Description: "Paso de la Luna por el nadir (lado opuesto de la Tierra)",
			StartTime:   t.Add(-60 * time.Minute).Format("15:04"),
			EndTime:     t.Add(60 * time.Minute).Format("15:04"),
			PeakTime:    t.Format("15:04"),
			RatingStars: 3,
		})
	}

	// Periodo Menor 1: Salida de la Luna (±45 min)
	if t, ok := parseTime(moonTimes.Moonrise); ok {
		periods = append(periods, SolunarPeriod{
			Type:        "Menor",
			Name:        "Periodo Menor 1",
			Description: "Salida de la Luna sobre el horizonte cantábrico",
			StartTime:   t.Add(-45 * time.Minute).Format("15:04"),
			EndTime:     t.Add(45 * time.Minute).Format("15:04"),
			PeakTime:    t.Format("15:04"),
			RatingStars: 3,
		})
	}

	// Periodo Menor 2: Puesta de la Luna (±45 min)
	if t, ok := parseTime(moonTimes.Moonset); ok {
		periods = append(periods, SolunarPeriod{
			Type:        "Menor",
			Name:        "Periodo Menor 2",
			Description: "Ocaso de la Luna en el horizonte",
			StartTime:   t.Add(-45 * time.Minute).Format("15:04"),
			EndTime:     t.Add(45 * time.Minute).Format("15:04"),
			PeakTime:    t.Format("15:04"),
			RatingStars: 2,
		})
	}

	// Calificación global de actividad solunar:
	// Factores: Fase lunar (Nueva/Llena aportan +2), Coeficiente de marea (>75 aporta +1)
	rating := 2
	if phase.IsSyzygy {
		rating += 2
	} else if phase.IsQuadrature {
		rating -= 1
	}
	if tideCoeff >= 80 {
		rating += 1
	} else if tideCoeff <= 45 {
		rating -= 1
	}
	if rating > 4 {
		rating = 4
	}
	if rating < 1 {
		rating = 1
	}

	var level, summary string
	switch rating {
	case 4:
		level = "Excelente"
		summary = "Actividad muy alta: coincidencia de mareas vivas y máxima influencia gravitacional."
	case 3:
		level = "Alta"
		summary = "Buenos periodos de actividad biológica marina y de pesca en la bahía de Gijón."
	case 2:
		level = "Media"
		summary = "Actividad moderada típica con picos centrados en los periodos mayores."
	default:
		level = "Baja"
		summary = "Marea muerta o desfase lunar; menor movimiento biológico en la costa."
	}

	return SolunarDay{
		ActivityLevel: level,
		OverallRating: rating,
		Summary:       summary,
		Periods:       periods,
	}
}

// ClimatologicalWaterTemp estima la temperatura del agua del mar en El Musel (Gijón)
// según el ciclo sinusoidal anual observado en las boyas de REDMAR
func ClimatologicalWaterTemp(date time.Time) float64 {
	doy := float64(date.YearDay())
	// Mínimo en febrero (~12.4ºC hacia el día 45), Máximo en agosto (~20.8ºC hacia el día 225)
	temp := 16.6 + 4.2*math.Cos(degToRad((doy-225.0)*360.0/365.25))
	return math.Round(temp*10) / 10
}
