package internal

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"sort"
	"time"

	"dv/db"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type Charts struct {
	dir  string
	repo *db.Queries
}

func NewCharts(repo *db.Queries, dir string) *Charts {
	return &Charts{
		repo: repo,
		dir:  dir,
	}
}

func (c *Charts) GenerateAllCharts() error {
	type job struct {
		name      string
		desc      string
		run       func() (int, error) // returns row count
		chartType string
	}
	jobs := []job{
		{"pie", "Distribution of movies by runtime duration segment (commercial success proxy)", c.PieChartWithCount, "Pie"},
		{"bar", "Average rating by genre (audience preference across genres)", c.BarChartWithCount, "Bar"},
		{"hbar", "Top studios by total revenue (market share of studios)", c.HorizontalBarWithCount, "Horizontal Bar"},
		{"line", "Average revenue trend by year (temporal performance)", c.LineChartWithCount, "Line"},
		{"hist", "Number of movies by release year (output volume over time)", c.MovieYearHistogramWithCount, "Histogram"},
		// seasonality chart (monthly releases by year) temporarily disabled until monthly query access implemented
		{"scatter", "Budget vs Revenue with ROI color (capital efficiency)", c.ScatterPlotWithCount, "Scatter"},
	}
	start := time.Now()
	for _, j := range jobs {
		rows, err := j.run()
		if err != nil {
			return fmt.Errorf("chart %s failed: %w", j.name, err)
		}
		fmt.Printf("[OK] %-10s %-12s rows=%d -> %s\n", j.chartType, j.name, rows, j.desc)
	}
	fmt.Printf("All charts generated in %s\n", time.Since(start).Round(time.Millisecond))
	return nil
}

func (c *Charts) Histogram() error { // kept for backward compatibility (no row count)
	_, err := c.HistogramWithCount()
	return err
}

// MovieYearHistogramWithCount builds a histogram-like bar chart of movie counts per release year
// using the YearlyTrends query (already filtered) and presents contiguous years on a numeric axis.
func (c *Charts) MovieYearHistogramWithCount() (int, error) {
	data, err := c.repo.YearlyTrends(context.TODO())
	if err != nil {
		return 0, fmt.Errorf("failed to get yearly trends: %w", err)
	}
	if len(data) == 0 {
		return 0, fmt.Errorf("no yearly data")
	}

	// Build map year->count then fill gaps
	yearMap := make(map[int]int64, len(data))
	minYear := int(data[0].Year)
	maxYear := int(data[0].Year)
	total := int64(0)
	maxCount := int64(0)
	for _, r := range data {
		y := int(r.Year)
		yearMap[y] = r.MoviesCount
		total += r.MoviesCount
		if r.MoviesCount > maxCount {
			maxCount = r.MoviesCount
		}
		if y < minYear {
			minYear = y
		}
		if y > maxYear {
			maxYear = y
		}
	}
	spanYears := maxYear - minYear + 1
	xs := make([]interface{}, 0, spanYears)
	bars := make([]opts.BarData, 0, spanYears)
	for y := minYear; y <= maxYear; y++ {
		cnt := yearMap[y]
		xs = append(xs, y)
		bars = append(bars, opts.BarData{Value: cnt})
	}
	avgPerYear := float64(total) / float64(spanYears)

	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Movies Released per Year", Subtitle: fmt.Sprintf("years=%d(%d-%d) total=%d avg≈%.1f max=%d", spanYears, minYear, maxYear, total, avgPerYear, maxCount)}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Year", Type: "category"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Movies"}),
	)
	bar.SetXAxis(xs).AddSeries("Movies", bars).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(false)}),
		charts.WithBarChartOpts(opts.BarChart{BarCategoryGap: "0%", BarGap: "0%"}),
		charts.WithMarkLineNameTypeItemOpts(opts.MarkLineNameTypeItem{Name: "Avg", Type: "average"}),
	)

	// Trend line over all years (including filled zeros)
	trend := make([]opts.LineData, 0, len(bars))
	for _, b := range bars {
		trend = append(trend, opts.LineData{Value: b.Value})
	}
	line := charts.NewLine()
	line.AddSeries("Trend", trend).SetSeriesOptions(
		charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}),
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(false)}),
	)
	bar.Overlap(line)
	return spanYears, c.render(bar, "movies_year_histogram.html")
}

func (c *Charts) HistogramWithCount() (int, error) {
	data, err := c.repo.ListTopProfitableMovies(context.TODO())
	if err != nil {
		return 0, fmt.Errorf("failed to get profitable movies: %w", err)
	}
	if len(data) == 0 {
		return 0, fmt.Errorf("no data for histogram")
	}
	values := make([]float64, 0, len(data))
	for _, m := range data {
		values = append(values, m.RoiPercent)
	}

	clipped := clipQuantiles(values, 0.01, 0.99)
	minV, maxV := minMax(clipped)
	if maxV == minV {
		maxV = minV + 1
	}

	// Freedman–Diaconis bin width
	iqr := quantile(clipped, 0.75) - quantile(clipped, 0.25)
	n := float64(len(clipped))
	width := 0.0
	if iqr > 0 {
		width = 2 * iqr / math.Cbrt(n)
	}
	if width <= 0 { // fallback Sturges
		bins := int(math.Ceil(1 + math.Log2(n)))
		if bins < 5 {
			bins = 5
		}
		width = (maxV - minV) / float64(bins)
	}
	bins := int(math.Ceil((maxV - minV) / width))
	if bins < 5 {
		bins = 5
	}
	if bins > 60 {
		bins = 60
	} // cap to keep chart readable
	width = (maxV - minV) / float64(bins)

	counts := make([]int, bins)
	densities := make([]float64, bins)
	// We will use numeric X (bin mid) to have visually contiguous columns (interval -> center) and remove category gaps
	binMids := make([]float64, bins)
	labels := make([]string, bins) // kept for tooltip formatting
	starts := make([]float64, bins)
	ends := make([]float64, bins)
	for i := 0; i < bins; i++ {
		start := minV + float64(i)*width
		end := start + width
		if i == bins-1 {
			end = maxV
		}
		labels[i] = fmt.Sprintf("%.1f – %.1f", start, end)
		binMids[i] = start + (end-start)/2
		starts[i] = start
		ends[i] = end
	}
	for _, v := range values { // use original distribution (not only clipped) but bin into clipped range
		if v < minV || v > maxV {
			continue
		}
		idx := int((v - minV) / width)
		if idx >= bins {
			idx = bins - 1
		}
		counts[idx]++
	}
	total := float64(len(values))
	for i, c := range counts {
		densities[i] = float64(c) / (total * width)
	}
	avg := mean(values)
	median := quantile(values, 0.5)

	barData := make([]opts.BarData, 0, bins)
	densityLine := make([]opts.LineData, 0, bins)
	for i, ctn := range counts {
		binMid := binMids[i]
		color := "#3498db"
		if binMid >= median {
			color = "#2ecc71"
		}
		barData = append(barData, opts.BarData{Value: []interface{}{binMid, ctn, starts[i], ends[i]}, ItemStyle: &opts.ItemStyle{Color: color}})
		densityLine = append(densityLine, opts.LineData{Value: []interface{}{binMid, densities[i]}})
	}
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "ROI% Histogram", Subtitle: fmt.Sprintf("n=%d mean=%.2f%% median=%.2f%% bins=%d width≈%.2f", len(values), avg, median, bins, width)}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Formatter: "{b}: {c}"}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
		charts.WithXAxisOpts(opts.XAxis{Name: "ROI %", Type: "value"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Count", Position: "left"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Density", Position: "right"}),
	)
	bar.AddSeries("Count", barData).SetSeriesOptions(
		charts.WithBarChartOpts(opts.BarChart{BarGap: "-100%", BarCategoryGap: "0%"}),
	)

	line := charts.NewLine()
	line.AddSeries("Density", densityLine).SetSeriesOptions(
		charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}),
		charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: opts.Float(0.15)}),
	)
	bar.Overlap(line)
	// Add median & average vertical lines via markLine on density series (closest approach)
	line.SetSeriesOptions(
		charts.WithMarkLineNameXAxisItemOpts(
			opts.MarkLineNameXAxisItem{Name: "Median", XAxis: median},
			opts.MarkLineNameXAxisItem{Name: "Mean", XAxis: avg},
		),
	)
	return len(values), c.render(bar, "histogram.html")
}

// --- histogram helpers ---
func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := 0.0
	for _, v := range xs {
		s += v
	}
	return s / float64(len(xs))
}
func minMax(xs []float64) (float64, float64) {
	m1 := math.MaxFloat64
	m2 := -math.MaxFloat64
	for _, v := range xs {
		if v < m1 {
			m1 = v
		}
		if v > m2 {
			m2 = v
		}
	}
	return m1, m2
}
func clipQuantiles(xs []float64, lo, hi float64) []float64 {
	if lo <= 0 && hi >= 1 {
		return xs
	}
	out := make([]float64, 0, len(xs))
	l := quantile(xs, lo)
	h := quantile(xs, hi)
	for _, v := range xs {
		if v >= l && v <= h {
			out = append(out, v)
		}
	}
	return out
}
func quantile(xs []float64, q float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	cp := append([]float64(nil), xs...)
	sort.Float64s(cp)
	if q <= 0 {
		return cp[0]
	}
	if q >= 1 {
		return cp[len(cp)-1]
	}
	pos := q * float64(len(cp)-1)
	i := int(pos)
	frac := pos - float64(i)
	if i+1 < len(cp) {
		return cp[i] + (cp[i+1]-cp[i])*frac
	}
	return cp[i]
}

func (c *Charts) BarChart() error { _, err := c.BarChartWithCount(); return err }
func (c *Charts) BarChartWithCount() (int, error) {
	data, err := c.repo.GenreAverageMetrics(context.TODO())
	if err != nil {
		return 0, err
	}
	genres := make([]string, 0, len(data))
	values := make([]opts.BarData, 0, len(data))
	for _, item := range data {
		genres = append(genres, item.GenreName.String)
		values = append(values, opts.BarData{Value: item.AvgRating})
	}
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Average Rating by Genre"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Genre", Type: "category"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Avg Rating"}),
	)
	bar.SetXAxis(genres).AddSeries("Avg Rating", values).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "top"}),
	)
	return len(data), c.render(bar, "bar.html")
}

func (c *Charts) LineChart() error { _, err := c.LineChartWithCount(); return err }
func (c *Charts) LineChartWithCount() (int, error) {
	data, err := c.repo.YearlyTrends(context.TODO())
	if err != nil {
		return 0, err
	}
	years := make([]string, 0, len(data))
	avgRevenue := make([]opts.LineData, 0, len(data))
	for _, item := range data {
		years = append(years, fmt.Sprintf("%d", item.Year))
		avgRevenue = append(avgRevenue, opts.LineData{Value: item.AvgRevenue})
	}
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Average Revenue by Year"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Year"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Average Revenue"}),
	)
	line.SetXAxis(years).AddSeries("Avg Revenue", avgRevenue).SetSeriesOptions(
		charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}),
	)
	return len(data), c.render(line, "line.html")
}

func (c *Charts) ScatterPlot() error { _, err := c.ScatterPlotWithCount(); return err }
func (c *Charts) ScatterPlotWithCount() (int, error) {
	data, err := c.repo.ListTopProfitableMovies(context.TODO())
	if err != nil {
		return 0, fmt.Errorf("failed to get profitable movies: %w", err)
	}
	points := make([]opts.ScatterData, 0, len(data))
	for _, m := range data {
		budget := m.Budget.Int32
		revenue := int64(0)
		if m.Revenue.Valid {
			revenue = m.Revenue.Int64
		}
		points = append(points, opts.ScatterData{
			Name:  m.Title.String,
			Value: []interface{}{budget, revenue, m.RoiPercent},
		})
	}
	// Sort points by budget so hover / color continuity improves (optional)
	slices.SortFunc(points, func(a, b opts.ScatterData) int {
		return int(a.Value.([]interface{})[0].(int32) - b.Value.([]interface{})[0].(int32))
	})
	scatter := charts.NewScatter()
	scatter.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Budget vs Revenue (Top Profitable Movies)"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Budget ($)"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Revenue ($)"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Formatter: "{b}"}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
	)
	scatter.AddSeries("Movies", points).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(false)}),
	)
	return len(data), c.render(scatter, "scatter.html")
}

func (c *Charts) PieChart() error { _, err := c.PieChartWithCount(); return err }
func (c *Charts) PieChartWithCount() (int, error) {
	data, err := c.repo.RuntimeSuccessSegments(context.TODO())
	if err != nil {
		return 0, fmt.Errorf("failed to get runtime success segments: %w", err)
	}
	items := make([]opts.PieData, 0, len(data))
	for _, d := range data {
		items = append(items, opts.PieData{Name: d.DurationCategory, Value: d.MoviesCount})
	}
	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Movie Duration Distribution"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
	)
	pie.AddSeries("Duration Segments", items).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Formatter: "{b}: {c} ({d}%)"}),
	)
	return len(data), c.render(pie, "pie.html")
}

// HorizontalBar shows top studios by total revenue
func (c *Charts) HorizontalBar() error { _, err := c.HorizontalBarWithCount(); return err }
func (c *Charts) HorizontalBarWithCount() (int, error) {
	data, err := c.repo.StudioPerformance(context.TODO())
	if err != nil {
		return 0, fmt.Errorf("failed to get studio performance: %w", err)
	}
	names := make([]string, 0, len(data))
	values := make([]opts.BarData, 0, len(data))
	for _, s := range data {
		names = append(names, s.CompanyName.String)
		values = append(values, opts.BarData{Value: s.TotalRevenue})
	}
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Top Studios by Total Revenue"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true)}),
		charts.WithYAxisOpts(opts.YAxis{Type: "category", Data: names, Name: "Studio"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Total Revenue"}),
	)
	bar.AddSeries("Revenue", values).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "right"}),
	)
	return len(data), c.render(bar, "studios_horizontal_bar.html")
}

type ChartRenderer interface {
	Render(w io.Writer) error
}

func (c *Charts) render(chart ChartRenderer, filename string) error {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}
	f, err := os.Create(c.dir + filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return chart.Render(f)
}
