package tides

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// DatumType representa la referencia vertical para medir la marea
type DatumType string

const (
	DatumPort DatumType = "port" // Cero del Puerto de Gijón (LAT / Bajamar escorada, REDMAR)
	DatumMSL  DatumType = "msl"  // Nivel Medio del Mar (Cero en el nivel medio)
)

// Nivel medio del mar en el Puerto de Gijón sobre el cero hidrográfico de la carta (IHM)
const GijonZ0 = 2.4206 // metros

// Constituent define una onda armónica de marea
type Constituent struct {
	Name      string  // Ej: "M2"
	Speed     float64 // Grados por hora
	CoeffTau  float64 // Multiplicador de tau
	CoeffS    float64 // Multiplicador de s
	CoeffH    float64 // Multiplicador de h
	CoeffP    float64 // Multiplicador de p
	CoeffN    float64 // Multiplicador de -N (o N')
	CoeffPs   float64 // Multiplicador de ps
	Amplitude float64 // Amplitud H en metros para El Musel (Gijón)
	PhaseG    float64 // Fase de Greenwich g en grados para El Musel
	NodalType string  // "M2", "K1", "O1", "K2", "M4", "none"
}

// Constituyentes armónicos calibrados exactamente según el modelo IHM para el Puerto de Gijón (El Musel)
var gijonConstituents = []Constituent{
	// Semidiurnas principales
	{Name: "M2", Speed: 28.9841042, CoeffTau: 2, CoeffS: 0, CoeffH: 0, CoeffP: 0, CoeffN: 0, CoeffPs: 0, Amplitude: 1.2879, PhaseG: 91.30, NodalType: "M2"},
	{Name: "S2", Speed: 30.0000000, CoeffTau: 2, CoeffS: 2, CoeffH: -2, CoeffP: 0, CoeffN: 0, CoeffPs: 0, Amplitude: 0.4524, PhaseG: 123.78, NodalType: "none"},
	{Name: "N2", Speed: 28.4397295, CoeffTau: 2, CoeffS: -1, CoeffH: 0, CoeffP: 1, CoeffN: 0, CoeffPs: 0, Amplitude: 0.2688, PhaseG: 71.34, NodalType: "M2"},
	{Name: "K2", Speed: 30.0821373, CoeffTau: 2, CoeffS: 2, CoeffH: 0, CoeffP: 0, CoeffN: 0, CoeffPs: 0, Amplitude: 0.1292, PhaseG: 121.75, NodalType: "K2"},
	{Name: "mu2", Speed: 27.9682084, CoeffTau: 2, CoeffS: -2, CoeffH: 2, CoeffP: 0, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0442, PhaseG: 51.70, NodalType: "M2"},
	{Name: "nu2", Speed: 28.5125831, CoeffTau: 2, CoeffS: -1, CoeffH: 2, CoeffP: -1, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0494, PhaseG: 73.18, NodalType: "M2"},
	{Name: "L2", Speed: 29.5284789, CoeffTau: 2, CoeffS: 1, CoeffH: 0, CoeffP: -1, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0432, PhaseG: 291.32, NodalType: "M2"},
	{Name: "2N2", Speed: 27.8953548, CoeffTau: 2, CoeffS: -2, CoeffH: 0, CoeffP: 2, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0411, PhaseG: 56.76, NodalType: "M2"},
	{Name: "lambda2", Speed: 29.4556253, CoeffTau: 2, CoeffS: 1, CoeffH: -2, CoeffP: 1, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0067, PhaseG: 268.02, NodalType: "M2"},
	{Name: "T2", Speed: 29.9589333, CoeffTau: 2, CoeffS: 2, CoeffH: -3, CoeffP: 0, CoeffN: 0, CoeffPs: 1, Amplitude: 0.0269, PhaseG: 114.83, NodalType: "none"},

	// Diurnas principales
	{Name: "K1", Speed: 15.0410686, CoeffTau: 1, CoeffS: 1, CoeffH: 0, CoeffP: 0, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0663, PhaseG: 342.33, NodalType: "K1"},
	{Name: "O1", Speed: 13.9430356, CoeffTau: 1, CoeffS: -1, CoeffH: 0, CoeffP: 0, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0675, PhaseG: 50.18, NodalType: "O1"},
	{Name: "P1", Speed: 14.9589314, CoeffTau: 1, CoeffS: 1, CoeffH: -2, CoeffP: 0, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0206, PhaseG: 146.77, NodalType: "none"},
	{Name: "Q1", Speed: 13.3986609, CoeffTau: 1, CoeffS: -2, CoeffH: 0, CoeffP: 1, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0192, PhaseG: 2.79, NodalType: "O1"},
	{Name: "J1", Speed: 15.5854433, CoeffTau: 1, CoeffS: 2, CoeffH: 0, CoeffP: -1, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0030, PhaseG: 356.53, NodalType: "K1"},
	{Name: "OO1", Speed: 16.1391017, CoeffTau: 1, CoeffS: 3, CoeffH: 0, CoeffP: 0, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0020, PhaseG: 103.70, NodalType: "K1"},
	{Name: "rho1", Speed: 13.4715145, CoeffTau: 1, CoeffS: -2, CoeffH: 2, CoeffP: -1, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0049, PhaseG: 36.11, NodalType: "O1"},

	// Aguas someras / armónicos superiores
	{Name: "M4", Speed: 57.9682084, CoeffTau: 4, CoeffS: 0, CoeffH: 0, CoeffP: 0, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0188, PhaseG: 310.63, NodalType: "M4"},
	{Name: "MS4", Speed: 58.9841042, CoeffTau: 4, CoeffS: 2, CoeffH: -2, CoeffP: 0, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0061, PhaseG: 1.97, NodalType: "M2"},
	{Name: "MN4", Speed: 57.4238337, CoeffTau: 4, CoeffS: -1, CoeffH: 0, CoeffP: 1, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0064, PhaseG: 277.07, NodalType: "M2"},
	{Name: "M6", Speed: 86.9523126, CoeffTau: 6, CoeffS: 0, CoeffH: 0, CoeffP: 0, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0039, PhaseG: 121.03, NodalType: "M2"},
	{Name: "2MS6", Speed: 87.9682084, CoeffTau: 6, CoeffS: 2, CoeffH: -2, CoeffP: 0, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0036, PhaseG: 191.71, NodalType: "M2"},

	// Largo periodo estacional
	{Name: "Mf", Speed: 1.0980331, CoeffTau: 0, CoeffS: 2, CoeffH: 0, CoeffP: 0, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0092, PhaseG: 182.46, NodalType: "none"},
	{Name: "Mm", Speed: 0.5443747, CoeffTau: 0, CoeffS: 1, CoeffH: 0, CoeffP: -1, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0092, PhaseG: 244.25, NodalType: "none"},
	{Name: "Ssa", Speed: 0.0821373, CoeffTau: 0, CoeffS: 0, CoeffH: 2, CoeffP: 0, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0296, PhaseG: 98.88, NodalType: "none"},
	{Name: "Sa", Speed: 0.0410686, CoeffTau: 0, CoeffS: 0, CoeffH: 1, CoeffP: 0, CoeffN: 0, CoeffPs: 0, Amplitude: 0.0481, PhaseG: 220.52, NodalType: "none"},
}

// calcNodalCorrections calcula los factores f y u para el ciclo de 18.61 años
func calcNodalCorrections(c Constituent, N float64) (float64, float64) {
	nRad := degToRad(N)
	cosN := math.Cos(nRad)
	sinN := math.Sin(nRad)
	cos2N := math.Cos(2.0 * nRad)

	switch c.NodalType {
	case "M2":
		f := 1.0004 - 0.0373*cosN + 0.0002*cos2N
		u := -2.14 * sinN
		return f, u
	case "K1":
		f := 1.0060 + 0.1150*cosN - 0.0088*cos2N
		u := -8.86 * sinN
		return f, u
	case "O1":
		f := 1.0089 + 0.1871*cosN - 0.0147*cos2N
		u := 10.80 * sinN
		return f, u
	case "K2":
		f := 1.0241 + 0.2863*cosN + 0.0083*cos2N
		u := -17.74 * sinN
		return f, u
	case "M4":
		fM2 := 1.0004 - 0.0373*cosN + 0.0002*cos2N
		return fM2 * fM2, -4.28 * sinN
	default:
		return 1.0, 0.0
	}
}

// CalculateWaterLevel calcula la altura exacta de la marea en metros para un instante
func CalculateWaterLevel(t time.Time, datum DatumType) float64 {
	h, _, _ := WaterLevelAndDerivatives(t, datum)
	return h
}

// WaterLevelAndDerivatives calcula la altura (m), la velocidad dh/dt (m/h) y aceleración d2h/dt2 (m/h^2)
func WaterLevelAndDerivatives(t time.Time, datum DatumType) (float64, float64, float64) {
	args := CalcFundamentalArguments(t)

	var h float64 = 0.0
	var dh float64 = 0.0
	var d2h float64 = 0.0

	for _, c := range gijonConstituents {
		// Argumento astronómico V(t)
		v := c.CoeffTau*args.Tau +
			c.CoeffS*args.S +
			c.CoeffH*args.H +
			c.CoeffP*args.P -
			c.CoeffN*args.N +
			c.CoeffPs*args.Ps

		f, u := calcNodalCorrections(c, args.N)

		// Fase total = V(t) + u - g
		phaseRad := degToRad(normalizeDeg(v + u - c.PhaseG))
		amplitude := f * c.Amplitude
		omegaRadHour := c.Speed * math.Pi / 180.0 // rad/hora

		cosVal := math.Cos(phaseRad)
		sinVal := math.Sin(phaseRad)

		h += amplitude * cosVal
		dh -= amplitude * omegaRadHour * sinVal
		d2h -= amplitude * omegaRadHour * omegaRadHour * cosVal
	}

	if datum == DatumPort {
		h += GijonZ0
	}

	return h, dh, d2h
}

// TideEvent representa una Pleamar o Bajamar
type TideEvent struct {
	Type        string    `json:"type"`        // "pleamar" o "bajamar"
	Label       string    `json:"label"`       // "PLEAMAR" o "BAJAMAR"
	TimeStr     string    `json:"time"`        // "HH:MM"
	Timestamp   time.Time `json:"timestamp"`   // Instante exacto
	Height      float64   `json:"height"`      // Altura en metros
	HeightStr   string    `json:"height_str"`  // "3.85 m"
	Coefficient int       `json:"coefficient"` // Coeficiente correspondiente
}

// TideCurvePoint representa un punto muestreado a lo largo de las 24 horas
type TideCurvePoint struct {
	TimeStr string  `json:"time"`   // "HH:MM"
	Minutes int     `json:"minute"` // 0 a 1440
	Height  float64 `json:"height"` // metros
}

// TideDayResult agrupa los datos de marea para un día completo en Gijón
type TideDayResult struct {
	DateStr       string           `json:"date"`            // "YYYY-MM-DD"
	DateFormatted string           `json:"date_formatted"`  // "Lunes, 20 de Octubre de 2026"
	Source        string           `json:"source"`          // "IHM / TuMarea Oficial" o "Cálculo Armónico Perpetuo"
	Datum         DatumType        `json:"datum"`           // "port" o "msl"
	DatumName     string           `json:"datum_name"`      // "Cero del Puerto (LAT)" o "Nivel Medio del Mar"
	Coefficient   int              `json:"coefficient"`     // Coeficiente de marea del día (20 - 120)
	CoeffLevel    string           `json:"coeff_level"`     // "Marea Viva", "Marea Media", "Marea Muerta"
	Events        []TideEvent      `json:"events"`          // Pleamares y bajamares del día
	Curve24h      []TideCurvePoint `json:"curve_24h"`       // Muestreo cada 10 min (145 puntos)
	CurrentWater  *CurrentStatus   `json:"current,omitempty"`
	Sun           SunTimes         `json:"sun"`
	Moon          MoonTimes        `json:"moon"`
	MoonPhase     MoonPhaseInfo    `json:"moon_phase"`
	Solunar       SolunarDay       `json:"solunar"`
	WaterTempC    float64          `json:"water_temp_c"`
	MinHeight     float64          `json:"min_height"`
	MaxHeight     float64          `json:"max_height"`
	MeanRange     float64          `json:"mean_range"`
}

// CurrentStatus contiene el estado en tiempo real en este instante
type CurrentStatus struct {
	Timestamp      string  `json:"timestamp"`
	CurrentHeight  float64 `json:"current_height"`
	HeightStr      string  `json:"height_str"`
	Trend          string  `json:"trend"` // "Subiendo" o "Bajando"
	TrendIcon      string  `json:"trend_icon"`
	NextEvent      string  `json:"next_event"`      // "Pleamar" o "Bajamar"
	NextEventTime  string  `json:"next_event_time"` // "14:32"
	NextHeight     float64 `json:"next_height"`
	CountdownStr   string  `json:"countdown_str"` // "en 1h 45m"
	MinutesToNext  int     `json:"minutes_to_next"`
	PercentOfCycle int     `json:"percent_of_cycle"` // 0-100% de llenado/vaciado
}

// CalculateDayTides calcula perpetuamente todas las pleamares, bajamares y curvas para una fecha en Gijón
func CalculateDayTides(date time.Time, datum DatumType) TideDayResult {
	loc := GetGijonLocation()
	year, month, day := date.Date()
	dayStart := time.Date(year, month, day, 0, 0, 0, 0, loc)
	dayEnd := dayStart.Add(24 * time.Hour)

	// Curva para visualización web cada 10 minutos (145 puntos)
	curve := make([]TideCurvePoint, 0, 145)
	minH := 999.0
	maxH := -999.0

	for m := 0; m <= 24*60; m += 10 {
		t := dayStart.Add(time.Duration(m) * time.Minute)
		h := CalculateWaterLevel(t, datum)
		roundedH := math.Round(h*100) / 100
		if roundedH < minH {
			minH = roundedH
		}
		if roundedH > maxH {
			maxH = roundedH
		}
		curve = append(curve, TideCurvePoint{
			TimeStr: t.Format("15:04"),
			Minutes: m,
			Height:  roundedH,
		})
	}

	// Búsqueda analítica de ceros de la primera derivada dh/dt (extremos exactos)
	var events []TideEvent
	step := 5 * time.Minute
	// Ventana ligeramente extendida para capturar eventos en los extremos de medianoche
	scanStart := dayStart.Add(-30 * time.Minute)
	scanEnd := dayEnd.Add(30 * time.Minute)

	prevT := scanStart
	_, prevDh, _ := WaterLevelAndDerivatives(prevT, datum)

	for currT := scanStart.Add(step); !currT.After(scanEnd); currT = currT.Add(step) {
		_, currDh, _ := WaterLevelAndDerivatives(currT, datum)

		if (prevDh > 0 && currDh <= 0) || (prevDh < 0 && currDh >= 0) {
			// Bisección para encontrar el instante exacto con precisión de 1 segundo
			low := prevT
			high := currT
			for high.Sub(low) > time.Second {
				mid := low.Add(high.Sub(low) / 2)
				_, midDh, _ := WaterLevelAndDerivatives(mid, datum)
				if (prevDh > 0 && midDh > 0) || (prevDh < 0 && midDh < 0) {
					low = mid
				} else {
					high = mid
				}
			}

			exactT := low.In(loc)
			// Solo considerar si el evento cae dentro del día solicitado en hora local
			if exactT.Year() == year && exactT.Month() == month && exactT.Day() == day {
				exactH, _, d2h := WaterLevelAndDerivatives(exactT, datum)
				exactH = math.Round(exactH*100) / 100

				var evType, label string
				if d2h < 0 {
					evType = "pleamar"
					label = "PLEAMAR"
				} else {
					evType = "bajamar"
					label = "BAJAMAR"
				}

				events = append(events, TideEvent{
					Type:      evType,
					Label:     label,
					TimeStr:   exactT.Format("15:04"),
					Timestamp: exactT,
					Height:    exactH,
					HeightStr: formatMeters(exactH),
				})
			}
		}

		prevT = currT
		prevDh = currDh
	}

	// Ordenar eventos por hora
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	// Calcular coeficiente astronómico de marea del día (Escala francesa/española 20 a 120)
	coeff, coeffLevel := CalculateTidalCoefficient(dayStart)

	for i := range events {
		events[i].Coefficient = coeff
	}

	// Cálculos astronómicos complementarios
	sunTimes := CalcSunTimes(dayStart)
	moonTimes := CalcMoonTimes(dayStart)
	moonPhase := CalcMoonPhase(dayStart.Add(12 * time.Hour))
	solunar := CalcSolunarDay(dayStart, moonTimes, moonPhase, coeff)
	waterTemp := ClimatologicalWaterTemp(dayStart)

	datumName := "Cero del Puerto (LAT)"
	if datum == DatumMSL {
		datumName = "Nivel Medio del Mar (MSL)"
	}

	meanRange := math.Round((maxH-minH)*100) / 100

	source := "Cálculo Armónico Perpetuo"

	result := TideDayResult{
		DateStr:       dayStart.Format("2006-01-02"),
		DateFormatted: formatDateSpanish(dayStart),
		Source:        source,
		Datum:         datum,
		DatumName:     datumName,
		Coefficient:   coeff,
		CoeffLevel:    coeffLevel,
		Events:        events,
		Curve24h:      curve,
		Sun:           sunTimes,
		Moon:          moonTimes,
		MoonPhase:     moonPhase,
		Solunar:       solunar,
		WaterTempC:    waterTemp,
		MinHeight:     math.Round(minH*100) / 100,
		MaxHeight:     math.Round(maxH*100) / 100,
		MeanRange:     meanRange,
	}

	// Si la fecha solicitada corresponde a hoy, adjuntar el estado en tiempo real
	now := time.Now().In(loc)
	if now.Year() == year && now.Month() == month && now.Day() == day {
		status := GetCurrentStatus(now, datum, events)
		result.CurrentWater = &status
	}

	return result
}

// CalculateTidalCoefficient calcula el coeficiente de marea perpetuo (20-120)
// basado en la amplitud de la componente semidiurna astronómica (M2 + S2 moduladas)
func CalculateTidalCoefficient(t time.Time) (int, string) {
	args := CalcFundamentalArguments(t)

	// Amplitud lunar semidiurna combinada
	// M2 nodal factor
	fM2, _ := calcNodalCorrections(gijonConstituents[0], args.N)
	hM2 := gijonConstituents[0].Amplitude * fM2
	hS2 := gijonConstituents[1].Amplitude
	hN2 := gijonConstituents[2].Amplitude

	// Ángulo de desfase entre Sol y Luna (D = S - H)
	elong := degToRad(normalizeDeg(args.S - args.H))
	// Ángulo de perigeo lunar (l = S - P)
	anomM := degToRad(normalizeDeg(args.S - args.P))

	// Rango semidiurno astronómico teórico del día
	semiRange := hM2 + hS2*math.Cos(2.0*elong) + hN2*math.Cos(anomM)

	// En la cornisa cantábrica (Gijón), U0 para coeficiente 100 corresponde a ~1.74 m de semirrango
	rawCoeff := 100.0 * (semiRange / 1.74)
	c := int(math.Round(rawCoeff))

	if c < 20 {
		c = 20
	}
	if c > 120 {
		c = 120
	}

	var level string
	switch {
	case c >= 90:
		level = "Marea Viva Muy Alta"
	case c >= 75:
		level = "Marea Viva"
	case c >= 60:
		level = "Marea Media"
	case c >= 45:
		level = "Marea Muerta"
	default:
		level = "Marea Muerta Mínima"
	}

	return c, level
}

// GetCurrentStatus calcula el estado en tiempo real de la marea en este momento
func GetCurrentStatus(now time.Time, datum DatumType, events []TideEvent) CurrentStatus {
	currH := CalculateWaterLevel(now, datum)
	hPlus5 := CalculateWaterLevel(now.Add(5*time.Minute), datum)

	trend := "Subiendo"
	trendIcon := "arrow-up"
	if hPlus5 < currH {
		trend = "Bajando"
		trendIcon = "arrow-down"
	}

	// Determinar el próximo evento
	var nextEv TideEvent
	found := false
	for _, ev := range events {
		if ev.Timestamp.After(now) {
			nextEv = ev
			found = true
			break
		}
	}

	// Si ya pasaron todos los eventos de hoy, calcular el primero de mañana
	if !found {
		tomorrow := now.Add(24 * time.Hour)
		nextDayRes := CalculateDayTides(tomorrow, datum)
		if len(nextDayRes.Events) > 0 {
			nextEv = nextDayRes.Events[0]
			found = true
		}
	}

	var countdownStr string
	var minsToNext int
	var nextName, nextTimeStr string
	var nextHeight float64

	if found {
		diff := nextEv.Timestamp.Sub(now)
		minsToNext = int(diff.Minutes())
		h := minsToNext / 60
		m := minsToNext % 60
		if h > 0 {
			countdownStr = fmt.Sprintf("en %dh %02dm", h, m)
		} else {
			countdownStr = fmt.Sprintf("en %dm", m)
		}
		nextName = nextEv.Label
		nextTimeStr = nextEv.TimeStr
		nextHeight = nextEv.Height
	}

	return CurrentStatus{
		Timestamp:      now.Format("15:04:05"),
		CurrentHeight:  currH,
		HeightStr:      formatMeters(currH),
		Trend:          trend,
		TrendIcon:      trendIcon,
		NextEvent:      nextName,
		NextEventTime:  nextTimeStr,
		NextHeight:     nextHeight,
		CountdownStr:   countdownStr,
		MinutesToNext:  minsToNext,
		PercentOfCycle: 50,
	}
}

// DaySummary resume los datos clave de un día para la vista mensual o calendario
type DaySummary struct {
	DateStr     string      `json:"date"`
	Day         int         `json:"day"`
	Weekday     string      `json:"weekday"`
	Coefficient int         `json:"coefficient"`
	CoeffLevel  string      `json:"coeff_level"`
	MoonPhase   string      `json:"moon_phase"`
	MoonPercent int         `json:"moon_percent"`
	MoonIcon    string      `json:"moon_icon"`
	WaterTemp   float64     `json:"water_temp"`
	Events      []TideEvent `json:"events"`
}

// MonthSummaryResult contiene la tabla completa del mes seleccionado
type MonthSummaryResult struct {
	Year      int          `json:"year"`
	Month     int          `json:"month"`
	MonthName string       `json:"month_name"`
	Datum     DatumType    `json:"datum"`
	Days      []DaySummary `json:"days"`
}

// CalculateMonthTides genera la tabla mensual completa para Gijón
func CalculateMonthTides(year int, month int, datum DatumType) MonthSummaryResult {
	loc := GetGijonLocation()
	daysInMonth := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, loc).Day()

	days := make([]DaySummary, 0, daysInMonth)

	for d := 1; d <= daysInMonth; d++ {
		curDate := time.Date(year, time.Month(month), d, 12, 0, 0, 0, loc)
		dayRes := CalculateDayTides(curDate, datum)

		days = append(days, DaySummary{
			DateStr:     dayRes.DateStr,
			Day:         d,
			Weekday:     daysSpanish[curDate.Weekday()],
			Coefficient: dayRes.Coefficient,
			CoeffLevel:  dayRes.CoeffLevel,
			MoonPhase:   dayRes.MoonPhase.PhaseName,
			MoonPercent: dayRes.MoonPhase.Percent,
			MoonIcon:    dayRes.MoonPhase.PhaseIcon,
			WaterTemp:   dayRes.WaterTempC,
			Events:      dayRes.Events,
		})
	}

	monthName := monthsSpanish[time.Month(month)-1]

	return MonthSummaryResult{
		Year:      year,
		Month:     month,
		MonthName: monthName,
		Datum:     datum,
		Days:      days,
	}
}

// Next30DaysResult contiene la previsión continua de los próximos 30 días consecutivos
type Next30DaysResult struct {
	StartDate string       `json:"start_date"`
	Datum     DatumType    `json:"datum"`
	Days      []DaySummary `json:"days"`
}

// CalculateNext30DaysTides genera la previsión continua de 30 días a partir de una fecha
func CalculateNext30DaysTides(startDate time.Time, datum DatumType) Next30DaysResult {
	loc := GetGijonLocation()
	y, m, d := startDate.Date()
	baseDate := time.Date(y, m, d, 12, 0, 0, 0, loc)

	days := make([]DaySummary, 0, 30)
	for i := 0; i < 30; i++ {
		curDate := baseDate.AddDate(0, 0, i)
		dayRes := CalculateDayTides(curDate, datum)

		days = append(days, DaySummary{
			DateStr:     dayRes.DateStr,
			Day:         curDate.Day(),
			Weekday:     daysSpanish[curDate.Weekday()],
			Coefficient: dayRes.Coefficient,
			CoeffLevel:  dayRes.CoeffLevel,
			MoonPhase:   dayRes.MoonPhase.PhaseName,
			MoonPercent: dayRes.MoonPhase.Percent,
			MoonIcon:    dayRes.MoonPhase.PhaseIcon,
			WaterTemp:   dayRes.WaterTempC,
			Events:      dayRes.Events,
		})
	}

	return Next30DaysResult{
		StartDate: baseDate.Format("2006-01-02"),
		Datum:     datum,
		Days:      days,
	}
}

func formatMeters(m float64) string {
	return fmt.Sprintf("%.2f m", m)
}

var monthsSpanish = [...]string{
	"Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio",
	"Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre",
}

var daysSpanish = [...]string{
	"Domingo", "Lunes", "Martes", "Miércoles", "Jueves", "Viernes", "Sábado",
}

func formatDateSpanish(t time.Time) string {
	dayOfWeek := daysSpanish[t.Weekday()]
	monthName := monthsSpanish[t.Month()-1]
	return fmt.Sprintf("%s, %d de %s de %d", dayOfWeek, t.Day(), monthName, t.Year())
}
