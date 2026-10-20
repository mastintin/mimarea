/**
 * MiMarea Gijón — Frontend Client
 * Interfaz interactiva y cálculo perpetuo para el Puerto de Gijón (El Musel)
 */

document.addEventListener('DOMContentLoaded', () => {
  // Estado local de la aplicación
  const state = {
    currentDate: getTodayDateString(),
    currentDatum: 'port', // 'port' (Cero del Puerto LAT) o 'msl' (Nivel medio)
    activeView: 'day',    // 'day' o 'month'
    dayData: null,
    monthData: null,
    pollInterval: null
  };

  // Elementos del DOM
  const dateInput = document.getElementById('dateInput');
  const btnToday = document.getElementById('btnToday');
  const btnPrevDay = document.getElementById('btnPrevDay');
  const btnNextDay = document.getElementById('btnNextDay');
  const datumSelect = document.getElementById('datumSelect');
  const btnViewDay = document.getElementById('btnViewDay');
  const btnViewMonth = document.getElementById('btnViewMonth');
  const dateDisplayTitle = document.getElementById('dateDisplayTitle');
  const coeffBadge = document.getElementById('coeffBadge');

  const realtimeCard = document.getElementById('realtimeCard');
  const rtHeight = document.getElementById('rtHeight');
  const rtTrendPill = document.getElementById('rtTrendPill');
  const rtTrendIcon = document.getElementById('rtTrendIcon');
  const rtTrendText = document.getElementById('rtTrendText');
  const rtNextTitle = document.getElementById('rtNextTitle');
  const rtCountdown = document.getElementById('rtCountdown');
  const rtCoeffVal = document.getElementById('rtCoeffVal');
  const rtCoeffDesc = document.getElementById('rtCoeffDesc');
  const rtWaterTemp = document.getElementById('rtWaterTemp');

  const dayViewContainer = document.getElementById('dayViewContainer');
  const monthViewContainer = document.getElementById('monthViewContainer');
  const eventsCardsContainer = document.getElementById('eventsCardsContainer');
  const tideChartSvg = document.getElementById('tideChartSvg');
  const chartTooltip = document.getElementById('chartTooltip');
  const chartDatumLabel = document.getElementById('chartDatumLabel');

  // Astro elementos
  const sunSunrise = document.getElementById('sunSunrise');
  const sunSunset = document.getElementById('sunSunset');
  const sunNoon = document.getElementById('sunNoon');
  const sunDuration = document.getElementById('sunDuration');
  const sunDawn = document.getElementById('sunDawn');

  const moonIconGraphic = document.getElementById('moonIconGraphic');
  const moonPhaseName = document.getElementById('moonPhaseName');
  const moonPercent = document.getElementById('moonPercent');
  const moonAge = document.getElementById('moonAge');
  const moonRise = document.getElementById('moonRise');
  const moonSet = document.getElementById('moonSet');
  const moonTransit = document.getElementById('moonTransit');

  const solunarStars = document.getElementById('solunarStars');
  const solunarLevel = document.getElementById('solunarLevel');
  const solunarSummary = document.getElementById('solunarSummary');
  const solunarPeriodsList = document.getElementById('solunarPeriodsList');

  // Mes elementos
  const monthTitle = document.getElementById('monthTitle');
  const monthTableBody = document.getElementById('monthTableBody');

  // 1. Inicialización
  dateInput.value = state.currentDate;
  loadDayData();

  // Polling en segundo plano para actualizar tiempo real cada 30 segundos
  state.pollInterval = setInterval(updateRealtimeStatus, 30000);

  // 2. Listeners de eventos
  dateInput.addEventListener('change', (e) => {
    if (e.target.value) {
      state.currentDate = e.target.value;
      if (state.activeView === 'day') {
        loadDayData();
      } else {
        loadMonthData();
      }
    }
  });

  btnToday.addEventListener('click', () => {
    state.currentDate = getTodayDateString();
    dateInput.value = state.currentDate;
    if (state.activeView === 'day') {
      loadDayData();
    } else {
      loadMonthData();
    }
  });

  btnPrevDay.addEventListener('click', () => {
    shiftDate(-1);
  });

  btnNextDay.addEventListener('click', () => {
    shiftDate(1);
  });

  datumSelect.addEventListener('change', (e) => {
    state.currentDatum = e.target.value;
    if (state.activeView === 'day') {
      loadDayData();
    } else {
      loadMonthData();
    }
  });

  btnViewDay.addEventListener('click', () => {
    switchView('day');
  });

  btnViewMonth.addEventListener('click', () => {
    switchView('month');
  });

  function shiftDate(days) {
    const d = new Date(state.currentDate + 'T12:00:00');
    d.setDate(d.getDate() + days);
    state.currentDate = formatDateToISO(d);
    dateInput.value = state.currentDate;
    if (state.activeView === 'day') {
      loadDayData();
    } else {
      loadMonthData();
    }
  }

  function switchView(view) {
    state.activeView = view;
    if (view === 'day') {
      btnViewDay.classList.add('active');
      btnViewMonth.classList.remove('active');
      dayViewContainer.classList.remove('hidden');
      monthViewContainer.classList.add('hidden');
      loadDayData();
    } else {
      btnViewMonth.classList.add('active');
      btnViewDay.classList.remove('active');
      dayViewContainer.classList.add('hidden');
      monthViewContainer.classList.remove('hidden');
      loadMonthData();
    }
  }

  // 3. Cargar datos del día
  async function loadDayData() {
    try {
      const url = `/api/day?date=${state.currentDate}&datum=${state.currentDatum}`;
      const res = await fetch(url);
      if (!res.ok) throw new Error('Error al cargar datos del día');
      const data = await res.json();
      state.dayData = data;
      renderDayView(data);
    } catch (err) {
      console.error('Error cargando datos de marea:', err);
    }
  }

  // 4. Renderizar vista diaria
  function renderDayView(data) {
    dateDisplayTitle.textContent = data.date_formatted;
    coeffBadge.textContent = `Coef. ${data.coefficient} • ${data.coeff_level}`;
    chartDatumLabel.textContent = `Ref: ${data.datum_name}`;

    // Actualizar banner en tiempo real
    const isToday = (data.date === getTodayDateString());
    if (isToday && data.current) {
      realtimeCard.classList.remove('hidden');
      renderRealtimeStatus(data.current, data.coefficient, data.coeff_level, data.water_temp_c);
    } else {
      realtimeCard.classList.add('hidden');
    }

    // Renderizar tarjetas de eventos (pleamares y bajamares)
    renderEventCards(data.events);

    // Renderizar gráfico interactivo de la curva de marea
    renderTideCurve(data.curve_24h, data.events, data.min_height, data.max_height, isToday);

    // Renderizar datos astronómicos
    renderAstroData(data.sun, data.moon, data.moon_phase, data.solunar);
  }

  // 5. Renderizar banner en tiempo real
  function renderRealtimeStatus(curr, coeff, coeffLevel, waterTemp) {
    rtHeight.textContent = `${curr.current_height.toFixed(2)} m`;
    rtTrendText.textContent = curr.trend;
    rtTrendIcon.textContent = curr.trend === 'Subiendo' ? '⬆' : '⬇';

    if (curr.trend === 'Subiendo') {
      rtTrendPill.className = 'trend-badge trend-up';
    } else {
      rtTrendPill.className = 'trend-badge trend-down';
    }

    rtNextTitle.textContent = `${curr.next_event} (${curr.next_event_time})`;
    rtCountdown.textContent = curr.countdown_str;

    rtCoeffVal.textContent = coeff;
    rtCoeffDesc.textContent = coeffLevel;

    rtWaterTemp.textContent = `${waterTemp.toFixed(1)} °C`;
  }

  async function updateRealtimeStatus() {
    if (state.currentDate !== getTodayDateString() || state.activeView !== 'day') return;
    try {
      const res = await fetch(`/api/status?datum=${state.currentDatum}`);
      if (!res.ok) return;
      const data = await res.json();
      if (data.current) {
        renderRealtimeStatus(data.current, data.coefficient, data.coeff_level, data.water_temp_c);
      }
    } catch (e) {
      // Ignorar errores transitorios de polling
    }
  }

  // 6. Tarjetas de pleamares y bajamares
  function renderEventCards(events) {
    eventsCardsContainer.innerHTML = '';
    if (!events || events.length === 0) {
      eventsCardsContainer.innerHTML = '<p class="text-muted">No se detectaron eventos para este día.</p>';
      return;
    }

    events.forEach(ev => {
      const card = document.createElement('div');
      const isPleamar = (ev.type === 'pleamar');
      card.className = `event-card ${isPleamar ? 'card-pleamar' : 'card-bajamar'}`;

      card.innerHTML = `
        <div class="event-info">
          <span class="event-type-badge">${isPleamar ? '🌊 PLEAMAR' : '🏖️ BAJAMAR'}</span>
          <span class="event-time">${ev.time}</span>
        </div>
        <div>
          <div class="event-height">${ev.height_str}</div>
          <div class="event-coeff-tag">Coef. ${ev.coefficient}</div>
        </div>
      `;
      eventsCardsContainer.appendChild(card);
    });
  }

  // 7. Renderizar curva de marea de 24 horas en SVG interactivo
  function renderTideCurve(curvePoints, events, minH, maxH, isToday) {
    if (!curvePoints || curvePoints.length === 0) return;

    const svgW = 960;
    const svgH = 280;
    const padTop = 35;
    const padBottom = 35;
    const padLeft = 45;
    const padRight = 25;

    const plotW = svgW - padLeft - padRight;
    const plotH = svgH - padTop - padBottom;

    // Rango vertical de marea (mínimo con margen)
    const yMin = Math.floor(minH * 2) / 2 - 0.2;
    const yMax = Math.ceil(maxH * 2) / 2 + 0.3;

    function getX(minute) {
      return padLeft + (minute / 1440) * plotW;
    }

    function getY(h) {
      return padTop + plotH - ((h - yMin) / (yMax - yMin)) * plotH;
    }

    let svgHtml = '';

    // Defs para degradados y filtros
    svgHtml += `
      <defs>
        <linearGradient id="tideAreaGrad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#0284c7" stop-opacity="0.45"/>
          <stop offset="100%" stop-color="#0284c7" stop-opacity="0.03"/>
        </linearGradient>
        <filter id="glow" x="-20%" y="-20%" width="140%" height="140%">
          <feGaussianBlur stdDeviation="3" result="blur" />
          <feComposite in="SourceGraphic" in2="blur" operator="over" />
        </filter>
      </defs>
    `;

    // Líneas de cuadrícula horizontales (Alturas en metros)
    const stepH = (yMax - yMin > 3.0) ? 1.0 : 0.5;
    for (let h = Math.ceil(yMin); h <= yMax; h += stepH) {
      const y = getY(h);
      if (y >= padTop && y <= padTop + plotH) {
        svgHtml += `
          <line x1="${padLeft}" y1="${y}" x2="${svgW - padRight}" y2="${y}" stroke="#1e3556" stroke-width="1" stroke-dasharray="3,3"/>
          <text x="${padLeft - 8}" y="${y + 4}" fill="#8ca3c7" font-size="11" text-anchor="end">${h.toFixed(1)}m</text>
        `;
      }
    }

    // Líneas de cuadrícula verticales (Horas: 00:00, 03:00, 06:00, 09:00, 12:00, 15:00, 18:00, 21:00, 24:00)
    for (let hr = 0; hr <= 24; hr += 3) {
      const min = hr * 60;
      const x = getX(min);
      const timeLabel = (hr < 10 ? '0' + hr : hr) + ':00';
      svgHtml += `
        <line x1="${x}" y1="${padTop}" x2="${x}" y2="${padTop + plotH}" stroke="#1e3556" stroke-width="1" stroke-dasharray="2,3"/>
        <text x="${x}" y="${svgH - 12}" fill="#8ca3c7" font-size="11" text-anchor="middle">${timeLabel}</text>
      `;
    }

    // Construcción de la ruta poligonal continua y del área sombreada
    let pathD = '';
    curvePoints.forEach((pt, idx) => {
      const x = getX(pt.minute);
      const y = getY(pt.height);
      if (idx === 0) {
        pathD += `M ${x.toFixed(1)} ${y.toFixed(1)}`;
      } else {
        pathD += ` L ${x.toFixed(1)} ${y.toFixed(1)}`;
      }
    });

    const xFirst = getX(curvePoints[0].minute);
    const xLast = getX(curvePoints[curvePoints.length - 1].minute);
    const yBottom = padTop + plotH;
    const areaD = `${pathD} L ${xLast.toFixed(1)} ${yBottom} L ${xFirst.toFixed(1)} ${yBottom} Z`;

    // Área rellena
    svgHtml += `<path d="${areaD}" fill="url(#tideAreaGrad)"/>`;
    // Trazo de la curva de agua
    svgHtml += `<path d="${pathD}" fill="none" stroke="#38bdf8" stroke-width="2.5" stroke-linecap="round"/>`;

    // Marcadores de pleamar y bajamar
    events.forEach(ev => {
      const parts = ev.time.split(':');
      const min = parseInt(parts[0], 10) * 60 + parseInt(parts[1], 10);
      const x = getX(min);
      const y = getY(ev.height);
      const isPleamar = (ev.type === 'pleamar');
      const color = isPleamar ? '#38bdf8' : '#fb923c';

      // Círculo del evento
      svgHtml += `
        <circle cx="${x.toFixed(1)}" cy="${y.toFixed(1)}" r="5" fill="${color}" stroke="#0b1320" stroke-width="2"/>
      `;

      // Etiqueta con hora y altura
      const textY = isPleamar ? y - 12 : y + 18;
      svgHtml += `
        <text x="${x.toFixed(1)}" y="${textY.toFixed(1)}" fill="${color}" font-size="11" font-weight="700" text-anchor="middle">
          ${ev.time} (${ev.height.toFixed(2)}m)
        </text>
      `;
    });

    // Marcador de hora actual si corresponde al día de hoy
    if (isToday) {
      const now = new Date();
      const currentMin = now.getHours() * 60 + now.getMinutes();
      if (currentMin >= 0 && currentMin <= 1440) {
        const nowX = getX(currentMin);
        svgHtml += `
          <line x1="${nowX.toFixed(1)}" y1="${padTop}" x2="${nowX.toFixed(1)}" y2="${padTop + plotH}" stroke="#e11d48" stroke-width="1.8" stroke-dasharray="3,3"/>
          <circle cx="${nowX.toFixed(1)}" cy="${padTop + 4}" r="4" fill="#e11d48" filter="url(#glow)"/>
        `;
      }
    }

    tideChartSvg.innerHTML = svgHtml;

    // Interacción hover / scrubber sobre la curva
    setupChartScrubber(curvePoints, padLeft, padRight, svgW);
  }

  function setupChartScrubber(curvePoints, padLeft, padRight, svgW) {
    const wrapper = tideChartSvg.parentElement;

    wrapper.onmousemove = (e) => {
      const rect = tideChartSvg.getBoundingClientRect();
      const mouseX = e.clientX - rect.left;
      const ratio = mouseX / rect.width;
      const minute = Math.round(ratio * 1440);

      if (minute < 0 || minute > 1440) {
        chartTooltip.style.display = 'none';
        return;
      }

      // Buscar el punto más cercano en la curva
      let closest = curvePoints[0];
      let minDiff = 999999;
      for (let pt of curvePoints) {
        const diff = Math.abs(pt.minute - minute);
        if (diff < minDiff) {
          minDiff = diff;
          closest = pt;
        }
      }

      const hStr = closest.height.toFixed(2) + ' m';
      chartTooltip.textContent = `${closest.time} — ${hStr}`;
      chartTooltip.style.left = `${mouseX}px`;
      chartTooltip.style.top = `${e.clientY - rect.top}px`;
      chartTooltip.style.display = 'block';
    };

    wrapper.onmouseleave = () => {
      chartTooltip.style.display = 'none';
    };
  }

  // 8. Renderizar información solar, lunar y solunar
  function renderAstroData(sun, moon, moonPhase, solunar) {
    sunSunrise.textContent = sun.sunrise;
    sunSunset.textContent = sun.sunset;
    sunNoon.textContent = sun.solar_noon;
    sunDuration.textContent = sun.day_length_str;
    sunDawn.textContent = sun.dawn_civil;

    moonPhaseName.textContent = moonPhase.phase_name;
    moonPercent.textContent = `Iluminación: ${moonPhase.percent}% (${moonPhase.trend})`;
    moonAge.textContent = `Edad lunar: ${moonPhase.age_days} días`;
    moonRise.textContent = moon.moonrise;
    moonSet.textContent = moon.moonset;
    moonTransit.textContent = moon.transit;

    moonIconGraphic.textContent = getMoonEmoji(moonPhase.phase_icon);

    // Solunar
    solunarLevel.textContent = solunar.activity_level;
    solunarSummary.textContent = solunar.summary;
    solunarStars.textContent = '⭐'.repeat(solunar.overall_rating);

    solunarPeriodsList.innerHTML = '';
    if (solunar.periods && solunar.periods.length > 0) {
      solunar.periods.forEach(p => {
        const item = document.createElement('div');
        item.className = 'solunar-period-item';
        item.innerHTML = `
          <div>
            <div class="solunar-period-type">${p.type === 'Mayor' ? '⭐' : '•'} ${p.name}</div>
            <div class="text-xs text-muted">${p.description}</div>
          </div>
          <div class="solunar-period-time">${p.start_time} - ${p.end_time}</div>
        `;
        solunarPeriodsList.appendChild(item);
      });
    } else {
      solunarPeriodsList.innerHTML = '<p class="text-muted text-xs">Sin periodos destacados hoy.</p>';
    }
  }

  function getMoonEmoji(iconCode) {
    switch (iconCode) {
      case 'new-moon': return '🌑';
      case 'waxing-crescent': return '🌒';
      case 'first-quarter': return '🌓';
      case 'waxing-gibbous': return '🌔';
      case 'full-moon': return '🌕';
      case 'waning-gibbous': return '🌖';
      case 'last-quarter': return '🌗';
      case 'waning-crescent': return '🌘';
      default: return '🌕';
    }
  }

  // 9. Cargar datos mensuales
  async function loadMonthData() {
    try {
      const d = new Date(state.currentDate + 'T12:00:00');
      const year = d.getFullYear();
      const month = d.getMonth() + 1;

      const url = `/api/month?year=${year}&month=${month}&datum=${state.currentDatum}`;
      const res = await fetch(url);
      if (!res.ok) throw new Error('Error al cargar datos mensuales');
      const data = await res.json();
      state.monthData = data;
      renderMonthView(data);
    } catch (err) {
      console.error('Error cargando mes:', err);
    }
  }

  function renderMonthView(data) {
    monthTitle.textContent = `Tabla de Mareas — ${data.month_name} de ${data.year}`;
    monthTableBody.innerHTML = '';

    const todayStr = getTodayDateString();

    data.days.forEach(day => {
      const tr = document.createElement('tr');
      if (day.date === todayStr) {
        tr.className = 'row-today';
      }

      const moonIcon = getMoonEmoji(day.moon_icon);

      let eventsCells = '';
      for (let i = 0; i < 4; i++) {
        if (day.events && day.events[i]) {
          const ev = day.events[i];
          const isPleamar = (ev.type === 'pleamar');
          const cls = isPleamar ? 'tide-cell-pleamar' : 'tide-cell-bajamar';
          const prefix = isPleamar ? '▲' : '▼';
          eventsCells += `<td><span class="tide-cell ${cls}">${prefix} ${ev.time} (${ev.height.toFixed(2)}m)</span></td>`;
        } else {
          eventsCells += `<td><span class="text-muted">—</span></td>`;
        }
      }

      tr.innerHTML = `
        <td><strong>${day.day}</strong> <span class="text-muted text-xs">${day.weekday.substring(0, 3)}</span></td>
        <td><span class="coeff-pill">${day.coefficient}</span></td>
        <td title="${day.moon_phase}">${moonIcon} ${day.moon_percent}%</td>
        ${eventsCells}
      `;

      tr.style.cursor = 'pointer';
      tr.addEventListener('click', () => {
        state.currentDate = day.date;
        dateInput.value = day.date;
        switchView('day');
      });

      monthTableBody.appendChild(tr);
    });
  }

  function getTodayDateString() {
    const d = new Date();
    return formatDateToISO(d);
  }

  function formatDateToISO(d) {
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${y}-${m}-${day}`;
  }
});
