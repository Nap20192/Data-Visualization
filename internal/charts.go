package internal

import (
	"context"
	"fmt"
	"io"
	"os"

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
	generators := []func() error{
		c.PieChart,
		c.BarChart,
		c.LineChart,
		c.ScatterPlot,
		c.Histogram,
	}
	for _, gen := range generators {
		if err := gen(); err != nil {
			return err
		}
	}
	return nil
}

// PIE: фильмы по странам
func (c *Charts) PieChart() error {
	data, err := c.repo.CountryProductionStats(context.TODO())
	if err != nil {
		return err
	}

	items := make([]opts.PieData, 0, len(data))
	for _, item := range data {
		items = append(items, opts.PieData{
			Name:  item.CountryName.String,
			Value: item.MoviesCount,
		})
	}

	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Film Production by Country"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
	)
	pie.AddSeries("Countries", items).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Show:      opts.Bool(true),
				Formatter: "{b}: {d}%",
			}),
		)

	return c.render(pie, "pie.html")
}

// BAR: средний рейтинг по жанрам
func (c *Charts) BarChart() error {
	data, err := c.repo.GenreAverageMetrics(context.TODO())
	if err != nil {
		return err
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
	)
	bar.SetXAxis(genres).
		AddSeries("Avg Rating", values)

	return c.render(bar, "bar.html")
}

// LINE: динамика выручки по десятилетиям
func (c *Charts) LineChart() error {
	data, err := c.repo.DecadeTrends(context.TODO())
	if err != nil {
		return err
	}
	decades := make([]string, 0, len(data))
	avgRevenue := make([]opts.LineData, 0, len(data))
	for _, item := range data {
		decades = append(decades, fmt.Sprintf("%d", item.Decade))
		avgRevenue = append(avgRevenue, opts.LineData{Value: item.AvgRevenue})
	}

	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Average Revenue by Decade"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
	)
	line.SetXAxis(decades).
		AddSeries("Avg Revenue", avgRevenue).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}))

	return c.render(line, "line.html")
}

func (c *Charts) ScatterPlot() error {
	data, err := c.repo.ListTopProfitableMovies(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get profitable movies: %w", err)
	}

	points := make([]opts.ScatterData, 0, len(data))
	for _, m := range data {
		budget := m.Budget.Int32

		revenue := int64(0)

		if m.Revenue.Valid {
			revenue = m.Revenue.Int64
		}

		points = append(points, opts.ScatterData{
			Name: m.Title.String,
			Value: []interface{}{
				budget,
				revenue,
				m.RoiPercent,
			},
		})
	}

	scatter := charts.NewScatter()
	scatter.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Budget vs Revenue (Top Profitable Movies)"}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Budget ($)",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Revenue ($)",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show: opts.Bool(true),
			Formatter: "{b}",
		}),
	)

	// убираем метки прямо на точках, чтобы не засорять график
	scatter.AddSeries("Movies", points).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{Show: opts.Bool(false)}),
		)

	return c.render(scatter, "scatter.html")
}

func (c *Charts) Histogram() error {

	data, err := c.repo.RuntimeSuccessSegments(context.TODO())

	if err != nil {
		return err
	}

	categories := make([]string, 0, len(data))
	values := make([]opts.BarData, 0, len(data))

	for _, r := range data {
		categories = append(categories, r.DurationCategory)
		values = append(values, opts.BarData{Value: r.MoviesCount})
	}

	hist := charts.NewBar()

	hist.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Movies by Runtime Segments"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true)}),
	)
	hist.SetXAxis(categories).AddSeries("Count", values)

	return c.render(hist, "histogram.html")
}

type ChartRenderer interface {
	Render(w io.Writer) error
}

func (c *Charts) render(chart ChartRenderer, filename string) error {
	f, err := os.Create(c.dir + filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return chart.Render(f)
}
