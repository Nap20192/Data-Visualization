package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 🔹 Подключение к БД
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/movies?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	rows, err := pool.Query(ctx, `
		SELECT movie_id, release_date
		FROM movie
	`)
	if err != nil {
		log.Fatalf("Ошибка запроса: %v", err)
	}
	defer rows.Close()

	type Movie struct {
		ID   int
		Date time.Time
	}
	var movies []Movie
	for rows.Next() {
		var id int
		var release sql.NullTime
		if err := rows.Scan(&id, &release); err != nil {
			log.Fatal(err)
		}
		if release.Valid {
			movies = append(movies, Movie{id, release.Time})
		} else {
			movies = append(movies, Movie{id, randomDate(2000, 2024)})
		}
	}

	counts := make(map[int]map[int]int) // year -> month -> count
	for _, m := range movies {
		y := m.Date.Year()
		mo := int(m.Date.Month())
		if counts[y] == nil {
			counts[y] = make(map[int]int)
		}
		counts[y][mo]++
	}

	var years []int
	for y := range counts {
		years = append(years, y)
	}
	sort.Ints(years)

	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	seriesData := make(map[int][]int, len(years))
	for _, y := range years {
		vals := make([]int, 12)
		for m := 1; m <= 12; m++ {
			vals[m-1] = counts[y][m]
		}
		seriesData[y] = vals
	}

	jsonData, err := json.Marshal(seriesData)
	if err != nil {
		log.Fatalf("marshal series: %v", err)
	}
	jsonYears, _ := json.Marshal(years)
	jsonMonths, _ := json.Marshal(months)

	file := "charts/timeline.html"
	f, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Встраиваем минимальный HTML + ECharts + слайдер
	html := fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="utf-8" />
<title>Monthly Movie Releases</title>
<script src="https://cdn.jsdelivr.net/npm/echarts@5/dist/echarts.min.js"></script>
<style>body{font-family:sans-serif;margin:16px;}#chart{width:100%%;height:520px;}#panel{display:flex;gap:1rem;align-items:center;margin-bottom:12px;}input[type=range]{width:320px;}</style>
</head><body>
<h2>Monthly Movie Releases (Year Filter)</h2>
<div id="panel">Year: <span id="yearVal"></span><input id="yearRange" type="range" min="0" max="%d" step="1" value="0" /></div>
<div id="chart"></div>
<script>
const months=%s;const years=%s;const data=%s;
const chart=echarts.init(document.getElementById('chart'));
function makeOption(yearIdx){const y=years[yearIdx];const vals=data[y];return {title:{text:'Monthly Releases '+y},tooltip:{trigger:'axis'},xAxis:{type:'category',data:months},yAxis:{type:'value',name:'Movies'},series:[{type:'line',smooth:true,data:vals,areaStyle:{opacity:0.18}}]};}
const yr=document.getElementById('yearVal');const slider=document.getElementById('yearRange');
function update(){yr.textContent=years[slider.value];chart.setOption(makeOption(slider.value));}
slider.addEventListener('input',update);update();
</script>
</body></html>`, len(years)-1, jsonMonths, jsonYears, jsonData)

	if _, err := f.WriteString(html); err != nil {
		log.Fatal(err)
	}
}

// Генерация случайной даты между заданными годами
func randomDate(startYear, endYear int) time.Time {
	min := time.Date(startYear, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(endYear, 12, 31, 0, 0, 0, 0, time.UTC).Unix()
	sec := rand.Int63n(max-min) + min
	return time.Unix(sec, 0)
}
