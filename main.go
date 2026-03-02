package main

import (
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
)

type Input struct {
	Hp, Cp, Sp, Np, Op, Wp, Ap float64
}

type Result struct {
	In Input

	Krs, Krg float64

	// dry
	Hc, Cc, Sc, Nc, Oc, Ac float64
	// combustible
	Hg, Cg, Sg, Ng, Og float64

	SumP, SumC, SumG float64

	Qrn, Qsn, Qhn float64

	WarnP bool
	WarnC bool
	WarnG bool
}

func near100(x float64) bool {
	return math.Abs(x-100.0) > 0.05 // допуск 0.05% (можеш змінити)
}

func parseFormFloat(r *http.Request, name string) (float64, error) {
	return strconv.ParseFloat(r.FormValue(name), 64)
}

var page = template.Must(template.New("page").Parse(`
<!doctype html>
<html lang="uk">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Практична №1 — Завдання 1</title>
  <style>
    :root { --bg:#0b0f19; --card:#121a2b; --text:#e7eefc; --muted:#9db0d0; --line:#23304a; --accent:#7aa2ff; --bad:#ff6b6b; --good:#2ee59d; }
    * { box-sizing: border-box; }
    body { margin:0; font-family: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Arial; background: var(--bg); color: var(--text); }
    a { color: var(--accent); text-decoration: none; }
    a:hover { text-decoration: underline; }
    .wrap { max-width: 980px; margin: 0 auto; padding: 28px 16px 60px; }
    .top { display:flex; gap:14px; align-items:flex-start; justify-content:space-between; flex-wrap: wrap; }
    .title h1 { margin:0; font-size: 26px; letter-spacing: 0.2px; }
    .title p { margin:6px 0 0; color: var(--muted); line-height: 1.4; }
    .card { margin-top: 18px; background: rgba(18,26,43,0.78); border: 1px solid rgba(35,48,74,0.75); box-shadow: 0 10px 30px rgba(0,0,0,0.35); border-radius: 18px; padding: 18px; backdrop-filter: blur(10px); }
    .grid { display:grid; gap: 12px; grid-template-columns: repeat(2, minmax(0,1fr)); }
    @media (min-width: 720px) { .grid { grid-template-columns: repeat(3, minmax(0,1fr)); } }
    .field { display:flex; flex-direction: column; gap: 6px; }
    label { color: var(--muted); font-size: 13px; }
    input[type=number] {
      width:100%;
      padding: 10px 12px;
      border-radius: 12px;
      border: 1px solid rgba(35,48,74,0.9);
      background: rgba(7,10,18,0.55);
      color: var(--text);
      outline: none;
    }
    input[type=number]:focus { border-color: rgba(122,162,255,0.9); box-shadow: 0 0 0 3px rgba(122,162,255,0.18); }
    .actions { display:flex; gap: 10px; flex-wrap: wrap; margin-top: 14px; }
    button {
      border: 0;
      padding: 11px 14px;
      border-radius: 12px;
      cursor: pointer;
      font-weight: 600;
    }
    .primary { background: var(--accent); color: #0b0f19; }
    .ghost { background: rgba(255,255,255,0.06); color: var(--text); border: 1px solid rgba(35,48,74,0.9); }
    .pill { display:inline-flex; gap:8px; align-items:center; padding:6px 10px; border-radius: 999px; border: 1px solid rgba(35,48,74,0.9); background: rgba(255,255,255,0.05); color: var(--muted); font-size: 12px; }
    .section-title { margin: 0 0 10px; font-size: 16px; }
    table { width:100%; border-collapse: collapse; overflow:hidden; border-radius: 14px; border: 1px solid rgba(35,48,74,0.9); }
    th, td { padding: 10px 10px; border-bottom: 1px solid rgba(35,48,74,0.75); text-align: right; }
    th { text-align: left; color: var(--muted); background: rgba(255,255,255,0.04); font-weight: 600; }
    tr:last-child td { border-bottom: 0; }
    .rowhead { text-align:left; }
    .two { display:grid; gap: 12px; grid-template-columns: 1fr; }
    @media (min-width: 900px) { .two { grid-template-columns: 1fr 1fr; } }
    .sum-good { color: var(--good); font-weight: 700; }
    .sum-bad { color: var(--bad); font-weight: 700; }
    .note { color: var(--muted); font-size: 13px; line-height: 1.4; margin-top: 10px; }
    .error { border: 1px solid rgba(255,107,107,0.55); background: rgba(255,107,107,0.08); padding: 12px; border-radius: 14px; color: var(--text); }
    .kv { display:flex; gap:10px; flex-wrap: wrap; margin-top: 10px; }
    .kv b { color: var(--text); }
  </style>
</head>
<body>
  <div class="wrap">
    <div class="top">
  <div class="title">
    <h1>Кушнір Катерина Вікторівна — ТВз-21</h1>
    <p>Практична №1 • Завдання 1 — Веб-калькулятор (суха/горюча маса та нижча теплота згоряння)</p>
  </div>
</div>

    {{if .Error}}
      <div class="card error">
        <b>Помилка:</b> {{.Error}}
      </div>
    {{end}}

    {{if .HasResult}}
      <div class="card">
        <h2 class="section-title">Результати</h2>

        <div class="kv">
          <span class="pill"><b>Kрс</b> = {{printf "%.6f" .Res.Krs}}</span>
          <span class="pill"><b>Kрг</b> = {{printf "%.6f" .Res.Krg}}</span>
          <span class="pill"><b>Qᵣⁿ</b> = {{printf "%.4f" .Res.Qrn}} МДж/кг</span>
          <span class="pill"><b>Qˢⁿ</b> = {{printf "%.4f" .Res.Qsn}} МДж/кг</span>
          <span class="pill"><b>Qʰⁿ</b> = {{printf "%.4f" .Res.Qhn}} МДж/кг</span>
        </div>

        <div class="two" style="margin-top:12px;">
          <div>
            <h3 class="section-title">Робоча маса (вхід)</h3>
            <table>
              <tr><th>Компонент</th><th>Значення</th></tr>
              <tr><td class="rowhead">Hᵖ, %</td><td>{{printf "%.4f" .Res.In.Hp}}</td></tr>
              <tr><td class="rowhead">Cᵖ, %</td><td>{{printf "%.4f" .Res.In.Cp}}</td></tr>
              <tr><td class="rowhead">Sᵖ, %</td><td>{{printf "%.4f" .Res.In.Sp}}</td></tr>
              <tr><td class="rowhead">Nᵖ, %</td><td>{{printf "%.4f" .Res.In.Np}}</td></tr>
              <tr><td class="rowhead">Oᵖ, %</td><td>{{printf "%.4f" .Res.In.Op}}</td></tr>
              <tr><td class="rowhead">Wᵖ, %</td><td>{{printf "%.4f" .Res.In.Wp}}</td></tr>
              <tr><td class="rowhead">Aᵖ, %</td><td>{{printf "%.4f" .Res.In.Ap}}</td></tr>
              <tr>
                <td class="rowhead"><b>Σ, %</b></td>
                <td class="{{if .Res.WarnP}}sum-bad{{else}}sum-good{{end}}">{{printf "%.4f" .Res.SumP}}</td>
              </tr>
            </table>
          </div>

          <div>
            <h3 class="section-title">Нижча теплота згоряння</h3>
            <table>
              <tr><th>Показник</th><th>Значення</th></tr>
              <tr><td class="rowhead">Qᵣⁿ (робоча маса), МДж/кг</td><td>{{printf "%.4f" .Res.Qrn}}</td></tr>
              <tr><td class="rowhead">Qˢⁿ (суха маса), МДж/кг</td><td>{{printf "%.4f" .Res.Qsn}}</td></tr>
              <tr><td class="rowhead">Qʰⁿ (горюча маса), МДж/кг</td><td>{{printf "%.4f" .Res.Qhn}}</td></tr>
            </table>
            <p class="note">
              Підказка: якщо Σ не ≈ 100%, перевір введені дані або округлення.
            </p>
          </div>
        </div>

        <div class="two" style="margin-top:12px;">
          <div>
            <h3 class="section-title">Суха маса</h3>
            <table>
              <tr><th>Компонент</th><th>Значення</th></tr>
              <tr><td class="rowhead">Hᶜ, %</td><td>{{printf "%.4f" .Res.Hc}}</td></tr>
              <tr><td class="rowhead">Cᶜ, %</td><td>{{printf "%.4f" .Res.Cc}}</td></tr>
              <tr><td class="rowhead">Sᶜ, %</td><td>{{printf "%.4f" .Res.Sc}}</td></tr>
              <tr><td class="rowhead">Nᶜ, %</td><td>{{printf "%.4f" .Res.Nc}}</td></tr>
              <tr><td class="rowhead">Oᶜ, %</td><td>{{printf "%.4f" .Res.Oc}}</td></tr>
              <tr><td class="rowhead">Aᶜ, %</td><td>{{printf "%.4f" .Res.Ac}}</td></tr>
              <tr>
                <td class="rowhead"><b>Σ, %</b></td>
                <td class="{{if .Res.WarnC}}sum-bad{{else}}sum-good{{end}}">{{printf "%.4f" .Res.SumC}}</td>
              </tr>
            </table>
          </div>

          <div>
            <h3 class="section-title">Горюча маса</h3>
            <table>
              <tr><th>Компонент</th><th>Значення</th></tr>
              <tr><td class="rowhead">Hᵍ, %</td><td>{{printf "%.4f" .Res.Hg}}</td></tr>
              <tr><td class="rowhead">Cᵍ, %</td><td>{{printf "%.4f" .Res.Cg}}</td></tr>
              <tr><td class="rowhead">Sᵍ, %</td><td>{{printf "%.4f" .Res.Sg}}</td></tr>
              <tr><td class="rowhead">Nᵍ, %</td><td>{{printf "%.4f" .Res.Ng}}</td></tr>
              <tr><td class="rowhead">Oᵍ, %</td><td>{{printf "%.4f" .Res.Og}}</td></tr>
              <tr>
                <td class="rowhead"><b>Σ, %</b></td>
                <td class="{{if .Res.WarnG}}sum-bad{{else}}sum-good{{end}}">{{printf "%.4f" .Res.SumG}}</td>
              </tr>
            </table>
          </div>
        </div>

        <div class="actions">
          <a class="ghost" href="/" style="display:inline-block; padding:11px 14px; border-radius:12px;">← Назад до форми</a>
        </div>
      </div>
    {{end}}

    <div class="card">
      <h2 class="section-title">Ввід даних</h2>
      <form method="post" action="/calculate">
        <div class="grid">
          <div class="field">
            <label>Hᵖ, %</label>
			<input name="Hp" step="any" type="number" placeholder="Введіть значення" value="{{if .Prefill}}{{printf "%.4f" .In.Hp}}{{end}}">
          </div>
          <div class="field">
            <label>Cᵖ, %</label>
			<input name="Cp" step="any" type="number" placeholder="Введіть значення" value="{{if .Prefill}}{{printf "%.4f" .In.Cp}}{{end}}">
          </div>
          <div class="field">
            <label>Sᵖ, %</label>
			<input name="Sp" step="any" type="number" placeholder="Введіть значення" value="{{if .Prefill}}{{printf "%.4f" .In.Sp}}{{end}}">
          </div>
          <div class="field">
            <label>Nᵖ, %</label>
			<input name="Np" step="any" type="number" placeholder="Введіть значення" value="{{if .Prefill}}{{printf "%.4f" .In.Np}}{{end}}">
          </div>
          <div class="field">
            <label>Oᵖ, %</label>
			<input name="Op" step="any" type="number" placeholder="Введіть значення" value="{{if .Prefill}}{{printf "%.4f" .In.Op}}{{end}}">
          </div>
          <div class="field">
            <label>Wᵖ, %</label>
			<input name="Wp" step="any" type="number" placeholder="Введіть значення" value="{{if .Prefill}}{{printf "%.4f" .In.Wp}}{{end}}">
          </div>
          <div class="field">
            <label>Aᵖ, %</label>
			<input name="Ap" step="any" type="number" placeholder="Введіть значення" value="{{if .Prefill}}{{printf "%.4f" .In.Ap}}{{end}}">
          </div>
        </div>

        <div class="actions">
  			<button class="primary" type="submit">Розрахувати</button>
  			<button class="ghost" type="button" onclick="fillVariant5()">Варіант №5</button>
		</div>
      </form>
    </div>

  </div>
  <script>
function fillVariant5() {
  const v = {
    Hp: 1.4, Cp: 70.5, Sp: 1.7, Np: 0.8, Op: 1.9, Wp: 7.0, Ap: 16.7
  };
  for (const k in v) {
	const el = document.querySelector('[name="' + k + '"]');
    if (el) el.value = v[k];
  }
}
</script>
</body>
</html>
`))

type PageData struct {
	In        Input
	HasResult bool
	Prefill   bool // NEW: whether to prefill inputs with numbers
	Res       Result
	Error     string
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		In:      Input{}, // empty
		Prefill: false,   // do NOT prefill
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = page.Execute(w, data)
}

func handleCalculate(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		In:      Input{},
		Prefill: true,
	}

	if r.Method != http.MethodPost {
		data.Error = "Невірний метод запиту."
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = page.Execute(w, data)
		return
	}
	if err := r.ParseForm(); err != nil {
		data.Error = "Не вдалося прочитати форму."
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = page.Execute(w, data)
		return
	}

	// read inputs
	var err error
	if data.In.Hp, err = parseFormFloat(r, "Hp"); err != nil {
		data.Error = "Hp має бути числом."
	}
	if data.In.Cp, err = parseFormFloat(r, "Cp"); err != nil {
		data.Error = "Cp має бути числом."
	}
	if data.In.Sp, err = parseFormFloat(r, "Sp"); err != nil {
		data.Error = "Sp має бути числом."
	}
	if data.In.Np, err = parseFormFloat(r, "Np"); err != nil {
		data.Error = "Np має бути числом."
	}
	if data.In.Op, err = parseFormFloat(r, "Op"); err != nil {
		data.Error = "Op має бути числом."
	}
	if data.In.Wp, err = parseFormFloat(r, "Wp"); err != nil {
		data.Error = "Wp має бути числом."
	}
	if data.In.Ap, err = parseFormFloat(r, "Ap"); err != nil {
		data.Error = "Ap має бути числом."
	}

	// show page even if error
	if data.Error != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = page.Execute(w, data)
		return
	}

	// validate denominators
	if 100.0-data.In.Wp <= 0 || 100.0-data.In.Wp-data.In.Ap <= 0 {
		data.Error = "Wp/Ap некоректні: 100−Wp і 100−Wp−Ap мають бути > 0."
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = page.Execute(w, data)
		return
	}

	// compute
	res := Result{In: data.In}

	res.Krs = 100.0 / (100.0 - res.In.Wp)
	res.Krg = 100.0 / (100.0 - res.In.Wp - res.In.Ap)

	res.Hc = res.In.Hp * res.Krs
	res.Cc = res.In.Cp * res.Krs
	res.Sc = res.In.Sp * res.Krs
	res.Nc = res.In.Np * res.Krs
	res.Oc = res.In.Op * res.Krs
	res.Ac = res.In.Ap * res.Krs

	res.Hg = res.In.Hp * res.Krg
	res.Cg = res.In.Cp * res.Krg
	res.Sg = res.In.Sp * res.Krg
	res.Ng = res.In.Np * res.Krg
	res.Og = res.In.Op * res.Krg

	res.SumP = res.In.Hp + res.In.Cp + res.In.Sp + res.In.Np + res.In.Op + res.In.Wp + res.In.Ap
	res.SumC = res.Hc + res.Cc + res.Sc + res.Nc + res.Oc + res.Ac
	res.SumG = res.Hg + res.Cg + res.Sg + res.Ng + res.Og

	res.WarnP = near100(res.SumP)
	res.WarnC = near100(res.SumC)
	res.WarnG = near100(res.SumG)

	res.Qrn = (339*res.In.Cp + 1030*res.In.Hp - 108.8*(res.In.Op-res.In.Sp) - 25*res.In.Wp) / 1000.0
	res.Qsn = (res.Qrn + 0.025*res.In.Wp) * 100.0 / (100.0 - res.In.Wp)
	res.Qhn = (res.Qrn + 0.025*res.In.Wp) * 100.0 / (100.0 - res.In.Wp - res.In.Ap)

	data.HasResult = true
	data.Res = res

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = page.Execute(w, data)
}

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/calculate", handleCalculate)

	log.Println("Server started: http://localhost:9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}
