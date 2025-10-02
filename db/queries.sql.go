package db

import (
	"context"
	"log/slog"
)

const actorRoleCounts = `-- name: ActorRoleCounts :many
SELECT
    p.person_name,
    COUNT(mc.movie_id) as roles_count,
    ROUND(AVG(m.vote_average), 2) as avg_movie_rating,
    ROUND(AVG(m.popularity), 2) as avg_movie_popularity
FROM movies.person p
    JOIN movies.movie_cast mc ON p.person_id = mc.person_id
    JOIN movies.movie m ON mc.movie_id = m.movie_id
WHERE
    m.vote_average > 0
GROUP BY
    p.person_id,
    p.person_name
HAVING
    COUNT(mc.movie_id) >= 5
ORDER BY roles_count DESC, avg_movie_rating DESC
LIMIT 20
`

type ActorRoleCountsRow struct {
	PersonName         string
	RolesCount         int
	AvgMovieRating     float64
	AvgMoviePopularity float64
}

// Actors with highest number of roles and average rating of their movies
func (q *Queries) ActorRoleCounts(ctx context.Context) ([]ActorRoleCountsRow, error) {
	slog.Debug("Executing ActorRoleCounts query")
	rows, err := q.db.Query(ctx, actorRoleCounts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ActorRoleCountsRow
	slog.Debug("Query executed, scanning rows")
	for rows.Next() {
		var i ActorRoleCountsRow
		if err := rows.Scan(
			&i.PersonName,
			&i.RolesCount,
			&i.AvgMovieRating,
			&i.AvgMoviePopularity,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const countryProductionStats = `-- name: CountryProductionStats :many
SELECT
    c.country_name,
    COUNT(m.movie_id) as movies_count,
    ROUND(AVG(m.budget), 0) as avg_budget,
    ROUND(AVG(m.revenue), 0) as avg_revenue,
    ROUND(AVG(m.vote_average), 2) as avg_rating
FROM movies.country c
    JOIN movies.production_country pc ON c.country_id = pc.country_id
    JOIN movies.movie m ON pc.movie_id = m.movie_id
WHERE
    m.budget > 0
    AND m.revenue > 0
GROUP BY
    c.country_id,
    c.country_name
HAVING
    COUNT(m.movie_id) >= 10
ORDER BY movies_count DESC
`

type CountryProductionStatsRow struct {
	CountryName string
	MoviesCount int
	AvgBudget   float64
	AvgRevenue  float64
	AvgRating   float64
}

// Geography of film production and average metrics
func (q *Queries) CountryProductionStats(ctx context.Context) ([]CountryProductionStatsRow, error) {
	rows, err := q.db.Query(ctx, countryProductionStats)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []CountryProductionStatsRow
	for rows.Next() {
		var i CountryProductionStatsRow
		if err := rows.Scan(
			&i.CountryName,
			&i.MoviesCount,
			&i.AvgBudget,
			&i.AvgRevenue,
			&i.AvgRating,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const directorPerformance = `-- name: DirectorPerformance :many
SELECT
    p.person_name as director_name,
    COUNT(m.movie_id) as directed_movies,
    ROUND(AVG(m.vote_average), 2) as avg_rating,
    ROUND(AVG(m.revenue), 0) as avg_revenue,
    ROUND(AVG(m.budget), 0) as avg_budget,
    SUM(m.revenue) as total_box_office
FROM movies.person p
    JOIN movies.movie_crew mc ON p.person_id = mc.person_id
    JOIN movies.movie m ON mc.movie_id = m.movie_id
    JOIN movies.department d ON mc.department_id = d.department_id
WHERE
    d.department_name = 'Directing'
    AND mc.job = 'Director'
    AND m.revenue > 0
    AND m.vote_average > 0
GROUP BY
    p.person_id,
    p.person_name
HAVING
    COUNT(m.movie_id) >= 3
ORDER BY avg_rating DESC, total_box_office DESC
LIMIT 15
`

type DirectorPerformanceRow struct {
	DirectorName   string
	DirectedMovies int
	AvgRating      float64
	AvgRevenue     float64
	AvgBudget      float64
	TotalBoxOffice int
}

// Top directors by average metrics of their movies
func (q *Queries) DirectorPerformance(ctx context.Context) ([]DirectorPerformanceRow, error) {
	rows, err := q.db.Query(ctx, directorPerformance)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DirectorPerformanceRow
	for rows.Next() {
		var i DirectorPerformanceRow
		if err := rows.Scan(
			&i.DirectorName,
			&i.DirectedMovies,
			&i.AvgRating,
			&i.AvgRevenue,
			&i.AvgBudget,
			&i.TotalBoxOffice,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const genreAverageMetrics = `-- name: GenreAverageMetrics :many
SELECT
    g.genre_name,
    COUNT(m.movie_id) as movies_count,
    ROUND(AVG(m.vote_average), 2) as avg_rating,
    ROUND(AVG(m.popularity), 2) as avg_popularity,
    ROUND(AVG(m.revenue), 0) as avg_revenue
FROM movies.genre g
    JOIN movies.movie_genres mg ON g.genre_id = mg.genre_id
    JOIN movies.movie m ON mg.movie_id = m.movie_id
WHERE
    m.vote_average > 0
GROUP BY
    g.genre_id,
    g.genre_name
ORDER BY avg_rating DESC
`

type GenreAverageMetricsRow struct {
	GenreName     string
	MoviesCount   int64
	AvgRating     float64
	AvgPopularity float64
	AvgRevenue    float64
}

// Analysis of genres by average metrics
func (q *Queries) GenreAverageMetrics(ctx context.Context) ([]GenreAverageMetricsRow, error) {
	rows, err := q.db.Query(ctx, genreAverageMetrics)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GenreAverageMetricsRow
	for rows.Next() {
		var i GenreAverageMetricsRow
		if err := rows.Scan(
			&i.GenreName,
			&i.MoviesCount,
			&i.AvgRating,
			&i.AvgPopularity,
			&i.AvgRevenue,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const keywordTrends = `-- name: KeywordTrends :many
SELECT
    k.keyword_name,
    COUNT(m.movie_id) as movies_count,
    ROUND(AVG(m.vote_average), 2) as avg_rating,
    ROUND(AVG(m.revenue), 0) as avg_revenue
FROM movies.keyword k
    JOIN movies.movie_keywords mk ON k.keyword_id = mk.keyword_id
    JOIN movies.movie m ON mk.movie_id = m.movie_id
WHERE
    m.vote_average > 0
GROUP BY
    k.keyword_id,
    k.keyword_name
HAVING
    COUNT(m.movie_id) >= 10
ORDER BY movies_count DESC, avg_rating DESC
LIMIT 20
`

type KeywordTrendsRow struct {
	KeywordName string
	MoviesCount int
	AvgRating   float64
	AvgRevenue  float64
}

// TOPIC 9: KEYWORDS AND TRENDS
func (q *Queries) KeywordTrends(ctx context.Context) ([]KeywordTrendsRow, error) {
	rows, err := q.db.Query(ctx, keywordTrends)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []KeywordTrendsRow
	for rows.Next() {
		var i KeywordTrendsRow
		if err := rows.Scan(
			&i.KeywordName,
			&i.MoviesCount,
			&i.AvgRating,
			&i.AvgRevenue,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const languagePopularity = `-- name: LanguagePopularity :many
SELECT
    l.language_name,
    COUNT(m.movie_id) as movies_count,
    ROUND(AVG(m.vote_average), 2) as avg_rating,
    ROUND(AVG(m.revenue), 0) as avg_revenue,
    ROUND(AVG(m.popularity), 2) as avg_popularity
FROM movies.language l
    JOIN movies.movie_languages ml ON l.language_id = ml.language_id
    JOIN movies.movie m ON ml.movie_id = m.movie_id
WHERE
    m.vote_average > 0
GROUP BY
    l.language_id,
    l.language_name
HAVING
    COUNT(m.movie_id) >= 5
ORDER BY movies_count DESC
`

type LanguagePopularityRow struct {
	LanguageName  string
	MoviesCount   int
	AvgRating     float64
	AvgRevenue    float64
	AvgPopularity float64
}

// Analysis of original movie languages
func (q *Queries) LanguagePopularity(ctx context.Context) ([]LanguagePopularityRow, error) {
	rows, err := q.db.Query(ctx, languagePopularity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []LanguagePopularityRow
	for rows.Next() {
		var i LanguagePopularityRow
		if err := rows.Scan(
			&i.LanguageName,
			&i.MoviesCount,
			&i.AvgRating,
			&i.AvgRevenue,
			&i.AvgPopularity,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listTopProfitableMovies = `-- name: ListTopProfitableMovies :many
SELECT
    title,
    budget,
    revenue,
    (revenue - budget) as profit,
    ROUND(
        (
            revenue::numeric / NULLIF(budget, 0) - 1
        ) * 100,
        2
    ) as roi_percent,
    vote_average
FROM movies.movie
WHERE
    budget > 0
    AND revenue > 0
ORDER BY profit DESC
LIMIT 10
`

type ListTopProfitableMoviesRow struct {
	Title       string
	Budget      int
	Revenue     int64
	Profit      int
	RoiPercent  float64
	VoteAverage float64
}

// Shows movies with highest revenue and profitability
func (q *Queries) ListTopProfitableMovies(ctx context.Context) ([]ListTopProfitableMoviesRow, error) {
	rows, err := q.db.Query(ctx, listTopProfitableMovies)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListTopProfitableMoviesRow
	for rows.Next() {
		var i ListTopProfitableMoviesRow
		if err := rows.Scan(
			&i.Title,
			&i.Budget,
			&i.Revenue,
			&i.Profit,
			&i.RoiPercent,
			&i.VoteAverage,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const moviesByDecade = `-- name: MoviesByDecade :many
SELECT
    FLOOR(
        EXTRACT(
            YEAR
            FROM release_date
        ) / 10
    ) * 10 as decade,
    COUNT(*) as movies_count,
    ROUND(AVG(budget), 0) as avg_budget,
    ROUND(AVG(revenue), 0) as avg_revenue,
    ROUND(AVG(vote_average), 2) as avg_rating,
    ROUND(AVG(runtime), 0) as avg_runtime
FROM movies.movie
WHERE
    release_date IS NOT NULL
    AND EXTRACT(
        YEAR
        FROM release_date
    ) >= 1970
GROUP BY
    FLOOR(
        EXTRACT(
            YEAR
            FROM release_date
        ) / 10
    ) * 10
ORDER BY decade
`

type MoviesByDecadeRow struct {
	Decade      int
	MoviesCount int
	AvgBudget   float64
	AvgRevenue  float64
	AvgRating   float64
	AvgRuntime  float64
}

// Number of movies and average metrics by decades
func (q *Queries) MoviesByDecade(ctx context.Context) ([]MoviesByDecadeRow, error) {
	rows, err := q.db.Query(ctx, moviesByDecade)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []MoviesByDecadeRow
	for rows.Next() {
		var i MoviesByDecadeRow
		if err := rows.Scan(
			&i.Decade,
			&i.MoviesCount,
			&i.AvgBudget,
			&i.AvgRevenue,
			&i.AvgRating,
			&i.AvgRuntime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const runtimeSuccessSegments = `-- name: RuntimeSuccessSegments :many
SELECT
    CASE
        WHEN runtime < 90 THEN 'Short (<90 min)'
        WHEN runtime BETWEEN 90 AND 120  THEN 'Medium (90-120 min)'
        WHEN runtime BETWEEN 121 AND 150  THEN 'Long (121-150 min)'
        ELSE 'Very long (>150 min)'
    END as duration_category,
    COUNT(*) as movies_count,
    ROUND(AVG(revenue), 0) as avg_revenue,
    ROUND(AVG(vote_average), 2) as avg_rating,
    ROUND(AVG(popularity), 2) as avg_popularity
FROM movies.movie
WHERE
    runtime IS NOT NULL
    AND runtime > 0
GROUP BY
    CASE
        WHEN runtime < 90 THEN 'Short (<90 min)'
        WHEN runtime BETWEEN 90 AND 120  THEN 'Medium (90-120 min)'
        WHEN runtime BETWEEN 121 AND 150  THEN 'Long (121-150 min)'
        ELSE 'Very long (>150 min)'
    END
ORDER BY avg_revenue DESC
`

type RuntimeSuccessSegmentsRow struct {
	DurationCategory string
	MoviesCount      int
	AvgRevenue       float64
	AvgRating        float64
	AvgPopularity    float64
}

// TOPIC 7: MOVIE DURATION AND COMMERCIAL SUCCESS
func (q *Queries) RuntimeSuccessSegments(ctx context.Context) ([]RuntimeSuccessSegmentsRow, error) {
	rows, err := q.db.Query(ctx, runtimeSuccessSegments)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RuntimeSuccessSegmentsRow
	for rows.Next() {
		var i RuntimeSuccessSegmentsRow
		if err := rows.Scan(
			&i.DurationCategory,
			&i.MoviesCount,
			&i.AvgRevenue,
			&i.AvgRating,
			&i.AvgPopularity,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const studioPerformance = `-- name: StudioPerformance :many
SELECT
    pc.company_name,
    COUNT(m.movie_id) as movies_count,
    ROUND(AVG(m.revenue), 0) as avg_revenue,
    ROUND(AVG(m.vote_average), 2) as avg_rating,
    SUM(m.revenue) as total_revenue
FROM movies.production_company pc
    JOIN movies.movie_company mcom ON pc.company_id = mcom.company_id
    JOIN movies.movie m ON mcom.movie_id = m.movie_id
WHERE
    m.revenue > 0
GROUP BY
    pc.company_id,
    pc.company_name
HAVING
    COUNT(m.movie_id) >= 3
ORDER BY total_revenue DESC
LIMIT 15
`

type StudioPerformanceRow struct {
	CompanyName  string
	MoviesCount  int
	AvgRevenue   float64
	AvgRating    float64
	TotalRevenue int
}

// Top studios by number of movies and average profit
func (q *Queries) StudioPerformance(ctx context.Context) ([]StudioPerformanceRow, error) {
	rows, err := q.db.Query(ctx, studioPerformance)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []StudioPerformanceRow
	for rows.Next() {
		var i StudioPerformanceRow
		if err := rows.Scan(
			&i.CompanyName,
			&i.MoviesCount,
			&i.AvgRevenue,
			&i.AvgRating,
			&i.TotalRevenue,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
