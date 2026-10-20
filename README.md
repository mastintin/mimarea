# 🌊 MiMarea Gijón

**MiMarea Gijón** es un servicio web y aplicación de escritorio ligera en Go para la predicción **perpetua** de mareas, datos astronómicos y actividad solunar en el **Puerto de Gijón (El Musel, Asturias)**.

Inspirado en la presentación visual y utilidad de portales como *tumarea.com*, MiMarea elimina por completo la dependencia de tablas de datos anuales precalculadas o descargadas: utiliza el **modelo de análisis armónico y la mecánica celeste astronómica** para calcular con precisión milimétrica la altura del agua, las pleamares, bajamares y coeficientes para **cualquier fecha del pasado, presente o futuro**.

---

## 🚀 Características Principales

1. **Cálculo Perpetuo (Sin tablas estáticas):**
   - Basado en los constituyentes armónicos astronómicos de la red de mareógrafos **REDMAR (Puertos del Estado / IHM)** para El Musel ($M_2, S_2, N_2, K_2, K_1, O_1, M_4$, etc.).
   - Modulación nodal lunar (ciclo de 18.61 años de regresión de nodos).
   - Efemérides solares y lunares puras (tiempo sidéreo, longitud, declinación, paralaje y anomalía).
   - Coeficiente de marea francés/español estándar (escala 20 a 120).
   - Selector libre para cualquier año y fecha (1990, 2026, 2035, 2050...).

2. **Ventana de Control Ligera (Estilo GoJubila PRO):**
   - Mini ventana de escritorio desarrollada con **Fyne v2**.
   - Indicador de estado en tiempo real (luz esmeralda de actividad).
   - Enlace directo Local (`http://localhost:8080`) y en Red Local (`http://<IP>:8080`) para acceder desde el móvil o tablet en la misma Wi-Fi.
   - Apertura automática del navegador al iniciar y parada ordenada al cerrar la ventana.

3. **Interfaz Web Interactiva y Moderna:**
   - **Estado en tiempo real:** Altura actual del mar, tendencia (Subiendo ⬆ / Bajando ⬇), cuenta atrás a la próxima pleamar o bajamar y temperatura del agua.
   - **Curva de marea de 24 horas:** Gráfico SVG continuo con marcador pulsante en tiempo real y *scrubber* táctil/ratón para consultar la altura en cualquier minuto del día.
   - **Sol y Luz diurna:** Salida (orto), puesta (ocaso), mediodía y duración del día.
   - **Luna y Fase lunar:** Representación visual de fase, porcentaje de iluminación, edad lunar, salida y puesta de la luna.
   - **Actividad Solunar (Teoría de John Alden Knight):** Horarios de Periodos Mayores (tránsito y nadir) y Menores, con calificación de actividad para pesca y actividades náuticas.
   - **Selector de Referencia Vertical (Datum):** Alterna instantáneamente entre el *Cero del Puerto (LAT)* y el *Nivel Medio del Mar (MSL)*.
   - **Vista Mensual:** Tabla completa de todos los días del mes con coeficientes y 4 eventos diarios.

---

## 🛠️ Instalación y Compilación

Requisitos: **Go 1.22+**

```bash
# Clonar o entrar al directorio del proyecto
cd mimarea

# Compilar el ejecutable
make build
# o manualmente:
go build -o mimarea main.go
```

---

## ▶️ Ejecución

### Modo Normal (con Ventana de Control Fyne y navegador)
```bash
./mimarea
# o con Make:
make run
```

### Opciones de línea de comandos
```bash
# Cambiar el puerto HTTP (por defecto 8080)
./mimarea -port 9090

# Modo servidor sin ventana gráfica (headless / consola)
./mimarea -headless

# Iniciar sin abrir el navegador automáticamente
./mimarea -no-browser
```

---

## 📡 API REST Integrada

| Método | Endpoint | Descripción |
|---|---|---|
| `GET` | `/api/status?datum=port` | Estado en tiempo real de la marea en este momento |
| `GET` | `/api/day?date=YYYY-MM-DD&datum=port` | Predicción completa del día (pleamares, curva 24h, sol, luna, solunar) |
| `GET` | `/api/month?year=YYYY&month=MM&datum=port` | Resumen mensual de pleamares, bajamares y coeficientes |
| `GET` | `/api/ping` | Verificación de estado del servicio |

---

## 📍 Estación Oceanográfica de Referencia

- **Localización:** Puerto de El Musel, Gijón (Asturias, España)
- **Coordenadas:** $43^\circ 33'\,\text{N}, 5^\circ 41'\,\text{W}$ ($43.5580^\circ\text{N}, -5.6984^\circ\text{W}$)
- **Huso Horario:** `Europe/Madrid` (CET / CEST con cambio estacional automático)
- **Cero de la Carta:** Bajamar escorada / Cero hidrográfico ($Z_0 \approx 2.38\text{ m}$)
