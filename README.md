# 🎬 Movie Database Analytics Platform

A comprehensive movie database analytics platform built with PostgreSQL and Apache Superset for data visualization and analysis.

## 📋 Overview

This project provides a complete setup for analyzing movie industry data using:
- **PostgreSQL 15** - Robust database with comprehensive movie schema
- **Apache Superset** - Modern data visualization and exploration platform
- **Docker Compose** - Simplified deployment and management

## 🗄️ Database Schema

The database contains comprehensive movie industry data with the following main entities:

### Core Tables
- **movies.movie** - Movie details (title, budget, revenue, ratings, release dates)
- **movies.person** - People in the industry (actors, directors, crew)
- **movies.country** - Countries involved in production
- **movies.production_company** - Movie production studios
- **movies.genre** - Movie genres
- **movies.keyword** - Thematic keywords and tags
- **movies.language** - Languages used in films

### Relationship Tables
- **movies.movie_cast** - Actor roles and characters
- **movies.movie_crew** - Director and crew assignments
- **movies.movie_company** - Studio-movie relationships
- **movies.movie_genres** - Genre classifications
- **movies.production_country** - Country-movie relationships
- **movies.movie_languages** - Language usage in films
- **movies.movie_keywords** - Thematic tagging

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose installed
- At least 4GB RAM available
- Ports 5432 and 8088 available

### 1. Clone and Start Services
```bash
git clone <repository-url>
cd dv
docker compose up --build -d
```

### 2. Wait for Initialization
The setup includes automatic initialization:
- PostgreSQL starts and loads movie data from `postgres/` directory
- Superset initializes with admin user and connects to the database

### 3. Access Superset
- **URL:** http://localhost:8088
- **Username:** admin
- **Password:** admin

## 📊 Database Connection in Superset

The database connection is automatically configured, but if needed manually:

**Connection URI:**
```
postgresql://postgres:postgres@postgres:5432/movies
```

**Individual Parameters:**
- **Host:** postgres
- **Port:** 5432
- **Database:** movies
- **Username:** postgres
- **Password:** postgres

## 🔍 Sample Analytics Queries

The project includes comprehensive analytical queries in `/queries/queries.sql`:

### Basic Operations
- Data validation and integrity checks
- Table exploration with LIMIT
- Filtering with WHERE and sorting with ORDER BY
- Aggregations using GROUP BY with COUNT, AVG, MIN, MAX
- Complex JOINs between multiple tables

### Analytics Topics
1. **Movie Industry Evolution** - Budget and revenue trends by decade
2. **Actor Commercial Success** - Top performers by box office
3. **Geographic Analysis** - Production by country and international success
4. **Release Seasonality** - Best months for movie releases
5. **Runtime Analysis** - How movie length affects success
6. **Studio Analysis** - Production company performance
7. **Language Distribution** - Multilingual cinema analysis
8. **Genre Trends** - Popular genres over time
9. **Keyword Analysis** - Trending themes and topics
10. **Director Rankings** - Comprehensive director performance metrics

## 📁 Project Structure

```
dv/
├── docker-compose.yml          # Docker services configuration
├── Dockerfile                  # Custom Superset container
├── init-superset.sh           # Superset initialization script
├── postgres/                  # Database schema and data
│   ├── 01_reference_data.sql  # Countries, languages, genres
│   ├── 02_keyword.sql         # Keywords and themes
│   ├── 03_person.sql          # People (actors, directors)
│   ├── 04_production_company.sql
│   ├── 05_movie.sql           # Core movie data
│   ├── 06_movie_cast.sql      # Actor assignments
│   ├── 07_movie_company.sql   # Studio relationships
│   ├── 08_movie_crew.sql      # Crew assignments
│   ├── 09_movie_genres.sql    # Genre classifications
│   ├── 10_movie_keywords.sql  # Keyword tagging
│   ├── 11_movie_languages.sql # Language usage
│   └── 12_production_country.sql
├── queries/
│   └── queries.sql            # Analytics and test queries
└── test.sql                   # Database validation queries
```

## 🛠️ Management Commands

### Start Services
```bash
docker compose up -d
```

### Stop Services
```bash
docker compose down
```

### Reset Everything (including data)
```bash
docker compose down -v
docker compose up --build -d
```

### View Logs
```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f postgres
docker compose logs -f superset
```

### Database Access
```bash
# Connect to PostgreSQL directly
docker exec -it dv-postgres-1 psql -U postgres -d movies

# Run SQL file
docker exec -i dv-postgres-1 psql -U postgres -d movies < queries/queries.sql
```

## 📈 Using Superset for Analytics

### Creating Your First Chart
1. Go to **SQL Lab** → **SQL Editor**
2. Select the `movies` database
3. Run analytical queries from `queries/queries.sql`
4. Save interesting results as datasets
5. Create charts in the **Chart** section
6. Build dashboards combining multiple charts

### Recommended Visualizations
- **Time Series** - Revenue/ratings trends over time
- **Bar Charts** - Top actors, studios, countries
- **Scatter Plots** - Budget vs Revenue correlation
- **Pie Charts** - Genre distribution
- **Tables** - Detailed rankings and comparisons

## 🔧 Configuration

### Environment Variables
Key configurations in `docker-compose.yml`:
- `POSTGRES_DB=movies`
- `POSTGRES_USER=postgres`
- `POSTGRES_PASSWORD=postgres`
- `SUPERSET_SECRET_KEY=mysecretkey123`

### Custom Configuration
To modify Superset settings, edit `init-superset.sh` or add configuration files to the container.

## 📊 Sample Insights

Using the provided queries, you can discover:
- **Top 10 most profitable movies** with ROI calculations
- **Genre popularity trends** across decades
- **Actor career analytics** with performance metrics
- **Studio market analysis** and competitive positioning
- **Geographic distribution** of movie production
- **Seasonal release patterns** for optimal timing
- **Runtime optimization** for audience engagement

## 🚨 Troubleshooting

### Common Issues

**Superset not accessible:**
```bash
# Check container status
docker compose ps

# Check Superset logs
docker compose logs superset
```

**Database connection issues:**
```bash
# Verify PostgreSQL is running
docker exec dv-postgres-1 pg_isready -U postgres

# Test connection
docker exec -it dv-postgres-1 psql -U postgres -d movies -c "SELECT COUNT(*) FROM movies.movie;"
```

**Data not loaded:**
```bash
# Check if SQL files were processed
docker compose logs postgres | grep "init"

# Manually reload if needed
docker compose down -v
docker compose up --build -d
```

## 📚 Additional Resources

- [Apache Superset Documentation](https://superset.apache.org/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Docker Compose Reference](https://docs.docker.com/compose/)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add new analytical queries or improvements
4. Test with sample data
5. Submit a pull request

## 📄 License

This project is open source and available under the MIT License.
