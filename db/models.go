package db

type MoviesCountry struct {
	CountryID      int
	CountryIsoCode string
	CountryName    string
}

type MoviesDepartment struct {
	DepartmentID   int
	DepartmentName string
}

type MoviesGenre struct {
	GenreID   int
	GenreName string
}

type MoviesKeyword struct {
	KeywordID   int
	KeywordName string
}

type MoviesLanguage struct {
	LanguageID   int
	LanguageCode string
	LanguageName string
}

type MoviesMovie struct {
	MovieID     int
	Title       string
	Budget      int
	Homepage    string
	Overview    string
	Popularity  float64
	ReleaseDate string
	Revenue     int64
	Runtime     int
	MovieStatus string
	Tagline     string
	VoteAverage float64
	VoteCount   int
}

type MoviesMovieCast struct {
	MovieID       int
	PersonID      int
	CharacterName string
	GenderID      int
	CastOrder     int
}

type MoviesMovieCompany struct {
	MovieID   int
	CompanyID int
}

type MoviesMovieCrew struct {
	MovieID      int
	PersonID     int
	DepartmentID int
	Job          string
}

type MoviesMovieGenre struct {
	MovieID int
	GenreID int
}

type MoviesMovieKeyword struct {
	MovieID   int
	KeywordID int
}

type MoviesMovieLanguage struct {
	MovieID        int
	LanguageID     int
	LanguageRoleID int
}

type MoviesPerson struct {
	PersonID   int
	PersonName string
}

type MoviesProductionCompany struct {
	CompanyID   int
	CompanyName string
}

type MoviesProductionCountry struct {
	MovieID   int
	CountryID int
}
