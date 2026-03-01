package main

import (
  "fmt"
  "html/template"
  "log"
  "net/http"
  "strconv"
)

var tpl = template.Must(template.New("index").Parse(`
<!doctype html>
<html lang="uk">
<head>
  <meta charset="utf-8">
  <title>Практична №1 — Завдання 1</title>
</head>
<body>
  <h1>Завдання 1: Веб-калькулятор</h1>

  <p>
    Введіть склад палива на робочу масу (Hᵖ, Cᵖ, Sᵖ, Nᵖ, Oᵖ, Wᵖ, Aᵖ),
    як вимагається в практичній. 
  </p>

  <form method="post" action="/calculate">
    <label>Hᵖ, %: <input name="Hp" step="any" type="number" value="1.4"></label><br>
    <label>Cᵖ, %: <input name="Cp" step="any" type="number" value="70.5"></label><br>
    <label>Sᵖ, %: <input name="Sp" step="any" type="number" value="1.7"></label><br>
    <label>Nᵖ, %: <input name="Np" step="any" type="number" value="0.8"></label><br>
    <label>Oᵖ, %: <input name="Op" step="any" type="number" value="1.9"></label><br>
    <label>Wᵖ, %: <input name="Wp" step="any" type="number" value="7.0"></label><br>
    <label>Aᵖ, %: <input name="Ap" step="any" type="number" value="16.7"></label><br><br>

    <button type="submit">Розрахувати</button>
  </form>
</body>
</html>
`))

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = tpl.Execute(w, nil)
}

func handleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad form", http.StatusBadRequest)
		return
	}

	// helper: safely parse float from form
	parse := func(name string) (float64, error) {
		// Using ParseFloat to support decimals like 1.4
		return strconv.ParseFloat(r.FormValue(name), 64)
	}

	Hp, err := parse("Hp")
	if err != nil { http.Error(w, "Bad Hp", http.StatusBadRequest); return }
	Cp, err := parse("Cp")
	if err != nil { http.Error(w, "Bad Cp", http.StatusBadRequest); return }
	Sp, err := parse("Sp")
	if err != nil { http.Error(w, "Bad Sp", http.StatusBadRequest); return }
	Np, err := parse("Np")
	if err != nil { http.Error(w, "Bad Np", http.StatusBadRequest); return }
	Op, err := parse("Op")
	if err != nil { http.Error(w, "Bad Op", http.StatusBadRequest); return }
	Wp, err := parse("Wp")
	if err != nil { http.Error(w, "Bad Wp", http.StatusBadRequest); return }
	Ap, err := parse("Ap")
	if err != nil { http.Error(w, "Bad Ap", http.StatusBadRequest); return }

	// basic validation: denominator must be > 0
	if 100-Wp <= 0 || 100-Wp-Ap <= 0 {
		http.Error(w, "Wp/Ap values make denominators invalid (100-Wp and 100-Wp-Ap must be > 0).", http.StatusBadRequest)
		return
	}

	// === From practical (control example): coefficients of transition ===
	Krs := 100.0 / (100.0 - Wp)          // to dry mass :contentReference[oaicite:8]{index=8}
	Krg := 100.0 / (100.0 - Wp - Ap)     // to combustible mass :contentReference[oaicite:9]{index=9}

	// dry mass composition
	Hc := Hp * Krs
	Cc := Cp * Krs
	Sc := Sp * Krs
	Nc := Np * Krs
	Oc := Op * Krs
	Ac := Ap * Krs

	// combustible mass composition
	Hg := Hp * Krg
	Cg := Cp * Krg
	Sg := Sp * Krg
	Ng := Np * Krg
	Og := Op * Krg

	// sums for check (as in example)
	sumP := Hp + Cp + Sp + Np + Op + Wp + Ap
	sumC := Hc + Cc + Sc + Nc + Oc + Ac
	sumG := Hg + Cg + Sg + Ng + Og

	// === Lower heating value for working mass (formula 1.2 in practical) === :contentReference[oaicite:10]{index=10}
	Qrn := (339*Cp + 1030*Hp - 108.8*(Op-Sp) - 25*Wp) / 1000.0

	// recalculation to dry and combustible mass (table 1.2 logic used in example) :contentReference[oaicite:11]{index=11}
	Qsn := (Qrn + 0.025*Wp) * 100.0 / (100.0 - Wp)
	Qhn := (Qrn + 0.025*Wp) * 100.0 / (100.0 - Wp - Ap)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
<!doctype html>
<html lang="uk"><head><meta charset="utf-8"><title>Результати — Завдання 1</title></head>
<body>
  <h1>Завдання 1 — Результати</h1>

  <h2>Вхідні дані (робоча маса)</h2>
  <ul>
    <li>Hᵖ=%.4f%%, Cᵖ=%.4f%%, Sᵖ=%.4f%%, Nᵖ=%.4f%%, Oᵖ=%.4f%%, Wᵖ=%.4f%%, Aᵖ=%.4f%%</li>
    <li>Перевірка суми: Σ=%.4f%%</li>
  </ul>

  <h2>Коефіцієнти переходу (табл. 1.1 / контрольний приклад)</h2>
  <ul>
    <li>K<sub>рс</sub> = 100/(100−Wᵖ) = %.6f</li>
    <li>K<sub>рг</sub> = 100/(100−Wᵖ−Aᵖ) = %.6f</li>
  </ul>

  <h2>Суха маса</h2>
  <ul>
    <li>Hᶜ=%.4f%%, Cᶜ=%.4f%%, Sᶜ=%.4f%%, Nᶜ=%.4f%%, Oᶜ=%.4f%%, Aᶜ=%.4f%%</li>
    <li>Перевірка суми: Σ=%.4f%%</li>
  </ul>

  <h2>Горюча маса</h2>
  <ul>
    <li>Hᵍ=%.4f%%, Cᵍ=%.4f%%, Sᵍ=%.4f%%, Nᵍ=%.4f%%, Oᵍ=%.4f%%</li>
    <li>Перевірка суми: Σ=%.4f%%</li>
  </ul>

  <h2>Нижча теплота згоряння</h2>
  <ul>
    <li>Q<sub>r</sub><sup>n</sup> (робоча маса) = %.4f МДж/кг</li>
    <li>Q<sub>s</sub><sup>n</sup> (суха маса) = %.4f МДж/кг</li>
    <li>Q<sub>h</sub><sup>n</sup> (горюча маса) = %.4f МДж/кг</li>
  </ul>

  <p><a href="/">← Назад до форми</a></p>
</body></html>
`,
		Hp, Cp, Sp, Np, Op, Wp, Ap, sumP,
		Krs, Krg,
		Hc, Cc, Sc, Nc, Oc, Ac, sumC,
		Hg, Cg, Sg, Ng, Og, sumG,
		Qrn, Qsn, Qhn,
	)
}

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/calculate", handleCalculate)

	log.Println("Server started: http://localhost:9090")
	log.Fatal(http.ListenAndServe(":9090", nil))
}