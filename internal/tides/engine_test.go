package tides

import (
	"testing"
	"time"
)

func TestNov1(t *testing.T) {
	loc := GetGijonLocation()
	for _, yr := range []int{2024, 2025, 2026} {
		t.Logf("=== 1 DE NOVIEMBRE DE %d ===", yr)
		d := time.Date(yr, 11, 1, 12, 0, 0, 0, loc)
		res := CalculateDayTides(d, DatumPort)
		for _, ev := range res.Events {
			t.Logf("  %s a las %s : %.2f m", ev.Label, ev.TimeStr, ev.Height)
		}
	}
}

func TestCalculateWaterLevelStep(t *testing.T) {
	loc := GetGijonLocation()
	t0 := time.Date(2026, 10, 20, 0, 0, 0, 0, loc)
	for m := 0; m <= 60; m += 10 {
		tm := t0.Add(time.Duration(m) * time.Minute)
		h := CalculateWaterLevel(tm, DatumPort)
		args := CalcFundamentalArguments(tm)
		t.Logf("m=%02d: h=%.2f tau=%.2f s=%.2f h=%.2f", m, h, args.Tau, args.S, args.H)
	}
}

func TestCalculateDayTides(t *testing.T) {
	loc := GetGijonLocation()
	testDate := time.Date(2026, 10, 20, 12, 0, 0, 0, loc)

	res := CalculateDayTides(testDate, DatumPort)

	if res.DateStr != "2026-10-20" {
		t.Fatalf("Fecha esperada 2026-10-20, obtenida: %s", res.DateStr)
	}

	if len(res.Events) < 3 || len(res.Events) > 5 {
		t.Fatalf("Esperados entre 3 y 4 eventos de marea en Gijón, obtenidos: %d", len(res.Events))
	}

	t.Logf("Fecha: %s (%s)", res.DateFormatted, res.DatumName)
	t.Logf("Coeficiente de marea: %d (%s)", res.Coefficient, res.CoeffLevel)
	t.Logf("Temperatura del agua estimada: %.1f ºC", res.WaterTempC)
	t.Logf("Sol: Orto %s | Ocaso %s | Duración %s", res.Sun.Sunrise, res.Sun.Sunset, res.Sun.DayLengthStr)
	t.Logf("Luna: Orto %s | Ocaso %s | Fase: %s (%d%%)", res.Moon.Moonrise, res.Moon.Moonset, res.MoonPhase.PhaseName, res.MoonPhase.Percent)
	t.Logf("Actividad Solunar: %s (Rating: %d estrellas)", res.Solunar.ActivityLevel, res.Solunar.OverallRating)

	for _, ev := range res.Events {
		t.Logf(" -> %s a las %s : %.2f m", ev.Label, ev.TimeStr, ev.Height)
		if ev.Height < 0.0 || ev.Height > 5.5 {
			t.Errorf("Altura de marea fuera de rango físico razonable para Gijón: %.2f m", ev.Height)
		}
	}

	// Comprobar alternancia de eventos (pleamar -> bajamar -> pleamar...)
	for i := 1; i < len(res.Events); i++ {
		if res.Events[i].Type == res.Events[i-1].Type {
			t.Errorf("Eventos consecutivos del mismo tipo: %s seguido de %s", res.Events[i-1].Type, res.Events[i].Type)
		}
	}

	if len(res.Curve24h) != 145 {
		t.Errorf("Curva 24h debería tener 145 puntos cada 10 min, obtenidos: %d", len(res.Curve24h))
	}
}

func TestPerpetualFutureAndPast(t *testing.T) {
	loc := GetGijonLocation()

	// Probar año 2035 (perpetuo futuro)
	futureDate := time.Date(2035, 6, 15, 12, 0, 0, 0, loc)
	resFuture := CalculateDayTides(futureDate, DatumPort)
	if len(resFuture.Events) < 3 {
		t.Fatalf("Fallo en predicción perpetua futura 2035: eventos %d", len(resFuture.Events))
	}
	t.Logf("2035-06-15: Coef %d, %d eventos", resFuture.Coefficient, len(resFuture.Events))

	// Probar año 1999 (perpetuo pasado)
	pastDate := time.Date(1999, 1, 1, 12, 0, 0, 0, loc)
	resPast := CalculateDayTides(pastDate, DatumPort)
	if len(resPast.Events) < 3 {
		t.Fatalf("Fallo en predicción perpetua pasada 1999: eventos %d", len(resPast.Events))
	}
	t.Logf("1999-01-01: Coef %d, %d eventos", resPast.Coefficient, len(resPast.Events))
}
