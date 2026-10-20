/**
 * MiMarea Gijón — WebAssembly Client para Safari / iPhone / Navegador
 * Ejecuta el motor matemático de Go compilado en WebAssembly directamente en el dispositivo.
 */

// Registrar Service Worker para caché offline y forzar actualización inmediata
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js').then((reg) => {
      reg.update();
    }).catch(() => {
      // Ignorar si se ejecuta en localhost o sin SSL
    });
  });
}

// 1. Cargar e inicializar Go WebAssembly
const go = new Go();

async function initWebAssembly() {
  const overlay = document.getElementById('wasmLoadingOverlay');
  try {
    let result;
    if (WebAssembly.instantiateStreaming) {
      try {
        result = await WebAssembly.instantiateStreaming(fetch('tides.wasm'), go.importObject);
      } catch (streamErr) {
        const response = await fetch('tides.wasm');
        const buffer = await response.arrayBuffer();
        result = await WebAssembly.instantiate(buffer, go.importObject);
      }
    } else {
      const response = await fetch('tides.wasm');
      const buffer = await response.arrayBuffer();
      result = await WebAssembly.instantiate(buffer, go.importObject);
    }

    // Iniciar el runtime de Go (se ejecuta en segundo plano)
    go.run(result.instance);

    // Ocultar pantalla de carga
    if (overlay) {
      overlay.style.opacity = '0';
      setTimeout(() => { overlay.style.display = 'none'; }, 300);
    }

    // Iniciar la aplicación
    startApplication();
  } catch (err) {
    console.error('Error cargando WebAssembly:', err);
    if (overlay) {
      overlay.innerHTML = `
        <div class="wasm-spinner-box">
          <h3 style="color:#f43f5e">Error al cargar WebAssembly</h3>
          <p class="text-xs text-muted">${err.message}</p>
        </div>
      `;
    }
  }
}

function startApplication() {
  // Estado local de la aplicación
  const state = {
    currentDate: getTodayDateString(),
    currentDatum: 'port', // 'port' (Cero del Puerto LAT) o 'msl' (Nivel medio)
    activeView: 'day',    // 'day' o 'month'
    mobileTab: 'marea',   // 'marea', 'chart', 'astro', 'month'
    dayData: null,
    monthData: null
  };

  // Elementos DOM de Controles
  const dateInput = document.getElementById('dateInput');
  const btnDateDisplay = document.getElementById('btnDateDisplay');
  const dateDisplayCompact = document.getElementById('dateDisplayCompact');
  const btnToday = document.getElementById('btnToday');
  const btnPrevDay = document.getElementById('btnPrevDay');
  const btnNextDay = document.getElementById('btnNextDay');
  const datumSelect = document.getElementById('datumSelect');
  const btnViewDay = document.getElementById('btnViewDay');
  const btnViewMonth = document.getElementById('btnViewMonth');
  const dateDisplayTitle = document.getElementById('dateDisplayTitle');
  const coeffBadge = document.getElementById('coeffBadge');

  // Contenedores Principales
  const dayViewContainer = document.getElementById('dayViewContainer');
  const monthViewContainer = document.getElementById('monthViewContainer');

  // Secciones Internas del Día
  const tumareaHeroCard = document.getElementById('tumareaHeroCard');
  const tidesCardsSection = document.getElementById('tidesCardsSection');
  const quickAstroStrip = document.getElementById('quickAstroStrip');
  const chartSection = document.getElementById('chartSection');
  const astroSection = document.getElementById('astroSection');
  const next30DaysSection = document.getElementById('next30DaysSection');
  const tbody30Days = document.getElementById('tbody30Days');
  const next30DaysDatumLabel = document.getElementById('next30DaysDatumLabel');

  // Elementos del Widget Hero TuMarea
  const tmTrendWord = document.getElementById('tmTrendWord');
  const tmSinceTime = document.getElementById('tmSinceTime');
  const tmArrow = document.getElementById('tmArrow');
  const tmWater = document.getElementById('tmWater');
  const tmLevelMax = document.getElementById('tmLevelMax');
  const tmLevelCurrent = document.getElementById('tmLevelCurrent');
  const tmLevelMin = document.getElementById('tmLevelMin');
  const tmNextName = document.getElementById('tmNextName');
  const tmNextTime = document.getElementById('tmNextTime');
  const tmNextHeight = document.getElementById('tmNextHeight');
  const tmCountdown = document.getElementById('tmCountdown');
  const tmRate10m = document.getElementById('tmRate10m');

  // Cuadrícula de 4 Mareas
  const eventsCardsContainer = document.getElementById('eventsCardsContainer');
  const tidesCardsDateLabel = document.getElementById('tidesCardsDateLabel');

  // Barra Rápida de Astronomía
  const quickSunrise = document.getElementById('quickSunrise');
  const quickMoonIcon = document.getElementById('quickMoonIcon');
  const quickMoonPhase = document.getElementById('quickMoonPhase');
  const quickMoonPercent = document.getElementById('quickMoonPercent');
  const quickSunset = document.getElementById('quickSunset');
  const quickCoeff = document.getElementById('quickCoeff');

  // Gráfico 24 Horas
  const tideChartSvg = document.getElementById('tideChartSvg');
  const chartTooltip = document.getElementById('chartTooltip');
  const chartDatumLabel = document.getElementById('chartDatumLabel');

  // Elementos Detalle Astronomía
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

  // Vista Mensual
  const monthTitle = document.getElementById('monthTitle');
  const monthTableBody = document.getElementById('monthTableBody');

  // Barra de Pestañas Móvil
  const mobileTabBtns = document.querySelectorAll('.mobile-tab-btn');

  // Inicializar UI
  dateInput.value = state.currentDate;
  loadDayData();

  // Actualizar en tiempo real cada 30 segundos
  setInterval(updateRealtimeClock, 30000);

  // ==========================================
  // EVENT LISTENERS
  // ==========================================
  dateInput.addEventListener('change', (e) => {
    if (e.target.value) {
      state.currentDate = e.target.value;
      refreshView();
    }
  });

  if (btnDateDisplay) {
    btnDateDisplay.addEventListener('click', () => {
      try {
        if (typeof dateInput.showPicker === 'function') {
          dateInput.showPicker();
        } else {
          dateInput.focus();
        }
      } catch (err) {
        dateInput.focus();
      }
    });
  }

  btnToday.addEventListener('click', () => {
    state.currentDate = getTodayDateString();
    dateInput.value = state.currentDate;
    refreshView();
  });

  btnPrevDay.addEventListener('click', () => { shiftDate(-1); });
  btnNextDay.addEventListener('click', () => { shiftDate(1); });

  datumSelect.addEventListener('change', (e) => {
    state.currentDatum = e.target.value;
    refreshView();
  });

  btnViewDay.addEventListener('click', () => { switchDesktopView('day'); });
  btnViewMonth.addEventListener('click', () => { switchDesktopView('month'); });

  // Pestañas móviles
  mobileTabBtns.forEach(btn => {
    btn.addEventListener('click', () => {
      setMobileTab(btn.dataset.tab);
    });
  });

  window.addEventListener('resize', applyResponsiveLayout);

  function shiftDate(days) {
    const d = new Date(state.currentDate + 'T12:00:00');
    d.setDate(d.getDate() + days);
    state.currentDate = formatDateToISO(d);
    dateInput.value = state.currentDate;
    refreshView();
  }

  function setMobileTab(tabName) {
    state.mobileTab = tabName;
    mobileTabBtns.forEach(btn => {
      btn.classList.toggle('active', btn.dataset.tab === tabName);
    });
    applyResponsiveLayout();
  }

  function switchDesktopView(view) {
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

  function applyResponsiveLayout() {
    const isMobile = window.innerWidth <= 768;

    if (!isMobile) {
      // Escritorio: según vista activa Día o 30 Días
      if (state.activeView === 'day') {
        dayViewContainer.classList.remove('hidden');
        monthViewContainer.classList.add('hidden');
        tumareaHeroCard.classList.remove('hidden');
        tidesCardsSection.classList.remove('hidden');
        quickAstroStrip.classList.remove('hidden');
        chartSection.classList.remove('hidden');
        astroSection.classList.remove('hidden');
      } else {
        dayViewContainer.classList.add('hidden');
        monthViewContainer.classList.remove('hidden');
        loadMonthData();
      }
      return;
    }

    // Móvil: gestionar pestañas individuales
    if (state.mobileTab === 'month') {
      dayViewContainer.classList.add('hidden');
      monthViewContainer.classList.remove('hidden');
      loadMonthData();
    } else {
      dayViewContainer.classList.remove('hidden');
      monthViewContainer.classList.add('hidden');

      const isMarea = (state.mobileTab === 'marea');
      const isChart = (state.mobileTab === 'chart');
      const isAstro = (state.mobileTab === 'astro');

      tumareaHeroCard.classList.toggle('hidden', !isMarea);
      tidesCardsSection.classList.toggle('hidden', !isMarea);
      quickAstroStrip.classList.toggle('hidden', !isMarea);
      chartSection.classList.toggle('hidden', !isChart);
      astroSection.classList.toggle('hidden', !isAstro);
    }
  }

  function refreshView() {
    if (state.activeView === 'day' || state.mobileTab !== 'month') {
      loadDayData();
    } else {
      loadMonthData();
    }
  }

  // ==========================================
  // CARGA DE DATOS DESDE WEBASSEMBLY
  // ==========================================
  function loadDayData() {
    if (typeof window.calcDayTides !== 'function') return;
    const jsonStr = window.calcDayTides(state.currentDate, state.currentDatum);
    const data = JSON.parse(jsonStr);
    state.dayData = data;
    renderDayView(data);
  }

  function renderDayView(data) {
    const isToday = (data.date === getTodayDateString());

    // 1. Títulos y etiquetas de fecha
    dateDisplayTitle.textContent = data.date_formatted;
    dateDisplayCompact.textContent = formatCompactDate(data.date, isToday);
    coeffBadge.textContent = `Coef. ${data.coefficient} • ${data.coeff_level}`;
    chartDatumLabel.textContent = `Ref: ${data.datum_name}`;
    tidesCardsDateLabel.textContent = isToday ? 'Predicción oficial para hoy' : data.date_formatted;

    // 2. Renderizar Widget Hero TuMarea
    renderTuMareaHero(data, isToday);

    // 3. Renderizar las 4 mareas del día
    renderEventCards(data.events, data.current);

    // 4. Renderizar barra rápida de astros
    renderQuickAstro(data);

    // 5. Renderizar curva de marea con soporte táctil
    renderTideCurve(data.curve_24h, data.events, data.min_height, data.max_height, isToday);

    // 6. Renderizar paneles de sol, luna y solunar
    renderAstroData(data.sun, data.moon, data.moon_phase, data.solunar);

    // 7. Aplicar visibilidad de pestañas
    applyResponsiveLayout();
  }

  // ==========================================
  // WIDGET HERO DIRECTO ESTILO TUMAREA
  // ==========================================
  function renderTuMareaHero(data, isToday) {
    const curr = data.current;
    const minH = data.min_height;
    const maxH = data.max_height;

    // Alturas extremas de referencia en vertical
    tmLevelMax.textContent = `${maxH.toFixed(2)} m`;
    tmLevelMin.textContent = `${minH.toFixed(2)} m`;

    if (isToday && curr) {
      // Estado de tendencia: Subiendo o Bajando
      const isUp = (curr.trend === 'Subiendo');
      tmTrendWord.textContent = isUp ? 'SUBIENDO' : 'BAJANDO';
      tmTrendWord.className = `tumarea-trend-badge ${isUp ? 'trend-up' : 'trend-down'}`;

      // Flecha gigante
      tmArrow.className = `tumarea-big-arrow ${isUp ? 'arrow-up' : 'arrow-down'}`;
      tmArrow.title = isUp ? 'Marea subiendo' : 'Marea bajando';

      // Nivel actual en vivo
      tmLevelCurrent.textContent = `${curr.current_height.toFixed(2)} m`;

      // Calcular porcentaje de agua en el vaso graduado
      const range = Math.max(0.1, maxH - minH);
      let pct = ((curr.current_height - minH) / range) * 100;
      pct = Math.max(10, Math.min(90, pct)); // Mantener entre 10% y 90% para que el barco siempre luzca
      tmWater.style.height = `${pct}%`;

      // Determinar la hora desde la que lleva subiendo o bajando
      const now = new Date();
      const currentMin = now.getHours() * 60 + now.getMinutes();
      let sinceTime = '--:--';

      if (data.events && data.events.length > 0) {
        // Buscar el evento pasado más reciente
        let lastPast = null;
        for (let ev of data.events) {
          const parts = ev.time.split(':');
          const evMin = parseInt(parts[0], 10) * 60 + parseInt(parts[1], 10);
          if (evMin <= currentMin) {
            lastPast = ev;
          }
        }
        if (lastPast) {
          sinceTime = lastPast.time;
        } else {
          // Si es antes de la primera marea del día, la anterior fue de ayer por la noche
          sinceTime = isUp ? 'las 23:30 de ayer' : 'las 23:30 de ayer';
        }
      }
      tmSinceTime.textContent = sinceTime;

      // Anuncio de la próxima marea
      tmNextName.textContent = curr.next_event;
      tmNextTime.textContent = curr.next_event_time;
      tmNextHeight.textContent = `${curr.next_height.toFixed(2)} m`;
      tmCountdown.textContent = curr.countdown_str.startsWith('en ') ? `Faltan ${curr.countdown_str.substring(3)}` : `Faltan ${curr.countdown_str}`;

      // Tasa de cambio de altura en 10 minutos
      let rate10mStr = 'En 10 min cambiará ~4 cm';
      if (data.curve_24h && data.curve_24h.length > 0) {
        const ptNow = data.curve_24h.find(p => p.minute >= currentMin) || data.curve_24h[0];
        const pt10 = data.curve_24h.find(p => p.minute >= currentMin + 10) || data.curve_24h[data.curve_24h.length - 1];
        const diffCm = Math.round((pt10.height - ptNow.height) * 100);
        if (diffCm > 0) {
          rate10mStr = `En 10 min el nivel subirá ~${diffCm} cm`;
        } else if (diffCm < 0) {
          rate10mStr = `En 10 min el nivel bajará ~${Math.abs(diffCm)} cm`;
        } else {
          rate10mStr = 'Nivel estable en el cambio de marea';
        }
      }
      tmRate10m.textContent = rate10mStr;
    } else {
      // Fecha futura o pasada
      tmTrendWord.textContent = 'PREDICCIÓN';
      tmTrendWord.className = 'tumarea-trend-badge trend-up';
      tmSinceTime.textContent = 'todo el día';
      tmArrow.className = 'tumarea-big-arrow arrow-up';

      const midH = (maxH + minH) / 2;
      tmLevelCurrent.textContent = `${midH.toFixed(2)} m (Media)`;
      tmWater.style.height = '50%';

      if (data.events && data.events.length > 0) {
        const first = data.events[0];
        tmNextName.textContent = first.label;
        tmNextTime.textContent = first.time;
        tmNextHeight.textContent = `${first.height.toFixed(2)} m`;
        tmCountdown.textContent = `Primera marea del día`;
      }
      tmRate10m.textContent = `Rango del día: ${(maxH - minH).toFixed(2)} m`;
    }
  }

  // ==========================================
  // CUADRÍCULA DE 4 MAREAS DEL DÍA
  // ==========================================
  function renderEventCards(events, current) {
    eventsCardsContainer.innerHTML = '';
    if (!events || events.length === 0) {
      eventsCardsContainer.innerHTML = '<p class="text-muted">No se detectaron eventos para este día.</p>';
      return;
    }

    const nextEventTime = current ? current.next_event_time : null;

    events.forEach(ev => {
      const card = document.createElement('div');
      const isPleamar = (ev.type === 'pleamar');
      const isNext = (nextEventTime && ev.time === nextEventTime);

      card.className = `event-card ${isPleamar ? 'card-pleamar' : 'card-bajamar'} ${isNext ? 'card-next-highlight' : ''}`;

      card.innerHTML = `
        <div class="event-badge-row">
          <span class="event-type-badge">${isPleamar ? '🌊 PLEAMAR' : '🏖️ BAJAMAR'}</span>
          ${isNext ? '<span class="next-pill">PRÓXIMA</span>' : ''}
        </div>
        <div class="event-time">${ev.time}</div>
        <div class="event-meta-right">
          <span class="event-height">${ev.height.toFixed(2)} m</span>
          <span class="event-coeff-tag">Coef. ${ev.coefficient}</span>
        </div>
      `;
      eventsCardsContainer.appendChild(card);
    });
  }

  // ==========================================
  // BARRA RÁPIDA DE ASTRONOMÍA
  // ==========================================
  function renderQuickAstro(data) {
    quickSunrise.textContent = data.sun.sunrise;
    quickMoonIcon.textContent = getMoonEmoji(data.moon_phase.phase_icon);
    quickMoonPhase.textContent = data.moon_phase.phase_name;
    quickMoonPercent.textContent = `${data.moon_phase.percent}%`;
    quickSunset.textContent = data.sun.sunset;
    quickCoeff.textContent = data.coefficient;
  }

  function updateRealtimeClock() {
    if (state.currentDate !== getTodayDateString()) return;
    loadDayData();
  }

  // ==========================================
  // CURVA DE MAREA INTERACTIVA (CON TOUCH SCRUBBING)
  // ==========================================
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

    const yMin = Math.floor(minH * 2) / 2 - 0.2;
    const yMax = Math.ceil(maxH * 2) / 2 + 0.3;

    function getX(minute) { return padLeft + (minute / 1440) * plotW; }
    function getY(h) { return padTop + plotH - ((h - yMin) / (yMax - yMin)) * plotH; }

    let svgHtml = `
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

    // Líneas horizontales (metros)
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

    // Líneas verticales (horas)
    for (let hr = 0; hr <= 24; hr += 3) {
      const min = hr * 60;
      const x = getX(min);
      const timeLabel = (hr < 10 ? '0' + hr : hr) + ':00';
      svgHtml += `
        <line x1="${x}" y1="${padTop}" x2="${x}" y2="${padTop + plotH}" stroke="#1e3556" stroke-width="1" stroke-dasharray="2,3"/>
        <text x="${x}" y="${svgH - 12}" fill="#8ca3c7" font-size="11" text-anchor="middle">${timeLabel}</text>
      `;
    }

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

    svgHtml += `<path d="${areaD}" fill="url(#tideAreaGrad)"/>`;
    svgHtml += `<path d="${pathD}" fill="none" stroke="#38bdf8" stroke-width="2.5" stroke-linecap="round"/>`;

    // Eventos
    events.forEach(ev => {
      const parts = ev.time.split(':');
      const min = parseInt(parts[0], 10) * 60 + parseInt(parts[1], 10);
      const x = getX(min);
      const y = getY(ev.height);
      const isPleamar = (ev.type === 'pleamar');
      const color = isPleamar ? '#38bdf8' : '#fb923c';

      svgHtml += `<circle cx="${x.toFixed(1)}" cy="${y.toFixed(1)}" r="5" fill="${color}" stroke="#0b1320" stroke-width="2"/>`;
      const textY = isPleamar ? y - 12 : y + 18;
      svgHtml += `<text x="${x.toFixed(1)}" y="${textY.toFixed(1)}" fill="${color}" font-size="11" font-weight="700" text-anchor="middle">${ev.time} (${ev.height.toFixed(2)}m)</text>`;
    });

    // Cursor actual
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
    setupChartScrubber(curvePoints);
  }

  function setupChartScrubber(curvePoints) {
    const wrapper = tideChartSvg.parentElement;

    const handleScrub = (clientX, clientY) => {
      const rect = tideChartSvg.getBoundingClientRect();
      const mouseX = clientX - rect.left;
      const ratio = mouseX / rect.width;
      const minute = Math.round(ratio * 1440);

      if (minute < 0 || minute > 1440) {
        chartTooltip.style.display = 'none';
        return;
      }

      let closest = curvePoints[0];
      let minDiff = 999999;
      for (let pt of curvePoints) {
        const diff = Math.abs(pt.minute - minute);
        if (diff < minDiff) {
          minDiff = diff;
          closest = pt;
        }
      }

      chartTooltip.textContent = `${closest.time} — ${closest.height.toFixed(2)} m`;
      chartTooltip.style.left = `${mouseX}px`;
      chartTooltip.style.top = `${clientY - rect.top}px`;
      chartTooltip.style.display = 'block';
    };

    // Soporte para ratón y eventos táctiles en móvil
    wrapper.onmousemove = (e) => handleScrub(e.clientX, e.clientY);
    wrapper.onmouseleave = () => { chartTooltip.style.display = 'none'; };

    wrapper.ontouchstart = (e) => {
      if (e.touches.length > 0) {
        handleScrub(e.touches[0].clientX, e.touches[0].clientY);
      }
    };
    wrapper.ontouchmove = (e) => {
      if (e.touches.length > 0) {
        handleScrub(e.touches[0].clientX, e.touches[0].clientY);
      }
    };
    wrapper.ontouchend = () => { chartTooltip.style.display = 'none'; };
  }

  // ==========================================
  // PANELES DETALLADOS DE ASTRONOMÍA
  // ==========================================
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

  // ==========================================
  // PREVISIÓN A 30 DÍAS DESDE EL DÍA SELECCIONADO (UNA MAREA POR LÍNEA)
  // ==========================================
  function loadMonthData() {
    if (typeof window.calcNext30DaysTides !== 'function') return;

    const jsonStr = window.calcNext30DaysTides(state.currentDate, state.currentDatum);
    const data = JSON.parse(jsonStr);
    state.monthData = data;
    renderMonthView(data);
  }

  function renderMonthView(data) {
    if (!data || !data.days) return;

    const datumLabel = (state.currentDatum === 'msl') ? 'Nivel Medio (MSL)' : 'Cero del Puerto (LAT)';
    const monthDatumBadge = document.getElementById('monthDatumBadge');
    if (monthDatumBadge) {
      monthDatumBadge.textContent = datumLabel;
    }

    const firstDateParts = state.currentDate.split('-');
    monthTitle.textContent = `Previsión a 30 Días (desde ${firstDateParts[2]}/${firstDateParts[1]}/${firstDateParts[0]})`;
    monthTableBody.innerHTML = '';

    const todayStr = getTodayDateString();
    const monthsShort = ['Ene', 'Feb', 'Mar', 'Abr', 'May', 'Jun', 'Jul', 'Ago', 'Sep', 'Oct', 'Nov', 'Dic'];

    data.days.forEach(day => {
      const tr = document.createElement('tr');
      const isToday = (day.date === todayStr);
      const isSelected = (day.date === state.currentDate);
      tr.className = `row-30day ${isToday ? 'row-today' : ''} ${isSelected && !isToday ? 'row-selected' : ''}`;

      const [y, m, d] = day.date.split('-').map(Number);
      const monthStr = monthsShort[m - 1];
      const moonIcon = getMoonEmoji(day.moon_icon);

      // Una marea por línea
      let tidesHtml = '';
      if (day.events && day.events.length > 0) {
        tidesHtml = day.events.map(ev => {
          const isP = (ev.type === 'pleamar');
          return `
            <div class="tide-line ${isP ? 'tide-p' : 'tide-b'}">
              <strong class="tide-line-time">${ev.time}</strong>
              <span class="tide-line-badge">${isP ? 'P' : 'B'}</span>
              <span class="tide-line-height">${ev.height.toFixed(2)} m</span>
            </div>
          `;
        }).join('');
      } else {
        tidesHtml = '<span class="text-muted text-xs">Sin eventos calculados</span>';
      }

      tr.innerHTML = `
        <td class="col-day">
          <div class="day-badge">
            <span class="day-weekday">${day.weekday.substring(0, 3)}</span>
            <span class="day-num">${day.day}</span>
            <span class="day-month">${monthStr}</span>
          </div>
        </td>
        <td class="col-tides">
          <div class="day-tides-stack">
            ${tidesHtml}
          </div>
        </td>
        <td class="col-meta">
          <div class="meta-stack">
            <span class="coeff-badge-mini" title="${day.coeff_level}">Coef ${day.coefficient}</span>
            <span class="moon-mini" title="${day.moon_phase}">${moonIcon} ${day.moon_percent}%</span>
          </div>
        </td>
      `;

      tr.addEventListener('click', () => {
        state.currentDate = day.date;
        dateInput.value = day.date;
        if (window.innerWidth <= 768) {
          setMobileTab('marea');
        } else {
          switchDesktopView('day');
        }
      });

      monthTableBody.appendChild(tr);
    });
  }

  // ==========================================
  // HELPERS DE FECHAS
  // ==========================================
  function getTodayDateString() {
    return formatDateToISO(new Date());
  }

  function formatDateToISO(d) {
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${y}-${m}-${day}`;
  }

  function formatCompactDate(isoDateStr, isToday) {
    const [y, m, d] = isoDateStr.split('-').map(Number);
    const dateObj = new Date(y, m - 1, d);
    const days = ['Dom', 'Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb'];
    const months = ['Ene', 'Feb', 'Mar', 'Abr', 'May', 'Jun', 'Jul', 'Ago', 'Sep', 'Oct', 'Nov', 'Dic'];
    const dayName = days[dateObj.getDay()];
    const monthName = months[dateObj.getMonth()];

    if (isToday) {
      return `Hoy, ${d} ${monthName}`;
    }
    return `${dayName}, ${d} ${monthName}`;
  }
}

// Iniciar WebAssembly al cargar la ventana
window.addEventListener('DOMContentLoaded', initWebAssembly);
