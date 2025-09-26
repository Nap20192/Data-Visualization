
SELECT 'Фильмы без названия' as check_type, COUNT(*) as count
FROM movies.movie
WHERE
    title IS NULL
    OR title = ''
UNION ALL
SELECT 'Фильмы без даты релиза', COUNT(*)
FROM movies.movie
WHERE
    release_date IS NULL
UNION ALL
SELECT 'Фильмы без бюджета', COUNT(*)
FROM movies.movie
WHERE
    budget IS NULL
    OR budget = 0
UNION ALL
SELECT 'Фильмы без доходов', COUNT(*)
FROM movies.movie
WHERE
    revenue IS NULL
    OR revenue = 0;

SELECT 'Роли без фильмов' as integrity_check, COUNT(*) as violations
FROM movies.movie_cast mc
    LEFT JOIN movies.movie m ON mc.movie_id = m.movie_id
WHERE
    m.movie_id IS NULL
UNION ALL
SELECT 'Роли без актеров', COUNT(*)
FROM movies.movie_cast mc
    LEFT JOIN movies.person p ON mc.person_id = p.person_id
WHERE
    p.person_id IS NULL;

SELECT
    MIN(release_date) as earliest_movie,
    MAX(release_date) as latest_movie,
    COUNT(
        DISTINCT EXTRACT(
            YEAR
            FROM release_date
        )
    ) as unique_years
FROM movies.movie
WHERE
    release_date IS NOT NULL;

SELECT 'Максимальный бюджет' as metric, MAX(budget) as value, title
FROM movies.movie
WHERE
    budget = (
        SELECT MAX(budget)
        FROM movies.movie
    )
UNION ALL
SELECT 'Максимальный доход', MAX(revenue), title
FROM movies.movie
WHERE
    revenue = (
        SELECT MAX(revenue)
        FROM movies.movie
    )
UNION ALL
SELECT 'Самый высокий рейтинг', MAX(vote_average), title
FROM movies.movie
WHERE
    vote_average = (
        SELECT MAX(vote_average)
        FROM movies.movie
    );

