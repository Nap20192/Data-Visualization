DROP TABLE IF EXISTS rating;

CREATE TABLE rating (
    id SERIAL PRIMARY KEY,
    rating_value NUMERIC(2, 1) NOT NULL CHECK (rating_value >= 1 AND rating_value <= 10),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rating_value ON rating(rating_value);

Drop table if exists movie_rating;

CREATE TABLE movie_rating (
    movie_id INT NOT NULL,
    rating_id INT NOT NULL,
    PRIMARY KEY (movie_id, rating_id),
    FOREIGN KEY (movie_id) REFERENCES movie(movie_id) ON DELETE CASCADE,
    FOREIGN KEY (rating_id) REFERENCES rating(id) ON DELETE CASCADE
);
