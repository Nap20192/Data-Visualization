-- name: ListTopProfitableMovies :many
-- Shows movies with highest revenue and profitability
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
LIMIT 10;

-- name: GenreAverageMetrics :many
-- Analysis of genres by average metrics
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
ORDER BY avg_rating DESC;

-- name: MoviesByDecade :many
-- Number of movies and average metrics by decades
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
ORDER BY decade;

-- name: ActorRoleCounts :many
-- Actors with highest number of roles and average rating of their movies
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
LIMIT 20;

-- name: StudioPerformance :many
-- Top studios by number of movies and average profit
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
LIMIT 15;

-- name: CountryProductionStats :many
-- Geography of film production and average metrics
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
ORDER BY movies_count DESC;

-- name: RuntimeSuccessSegments :many
-- TOPIC 7: MOVIE DURATION AND COMMERCIAL SUCCESS
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
ORDER BY avg_revenue DESC;

-- name: LanguagePopularity :many
-- Analysis of original movie languages
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
ORDER BY movies_count DESC;

-- name: KeywordTrends :many
-- TOPIC 9: KEYWORDS AND TRENDS
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
LIMIT 20;

-- name: DirectorPerformance :many
-- Top directors by average metrics of their movies
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
LIMIT 15;
