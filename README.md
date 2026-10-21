# Saldo

**A personal finance management API built with Go.**

Saldo is a backend project designed to help users organize financial transactions, import bank statement data, and analyze income and expenses through flexible queries and financial summaries.

The project focuses on clean separation of responsibilities, maintainable code, automated testing, and practical application of Hexagonal Architecture.

## Features

- **Transaction management** — register and list financial transactions.
- **CSV import** — import transactions from bank statement files.
- **Advanced filtering** — filter transactions by date range, type, and amount.
- **Pagination** — retrieve large transaction collections using `offset` and `limit`.
- **Financial summaries** — aggregate income and expenses over selected periods.
- **REST API** — expose functionality through HTTP endpoints with JSON responses.
- **API documentation** — describe endpoints, parameters, request bodies, and responses using OpenAPI.
- **Automated testing** — unit tests and database integration tests.
- **Database migrations** — manage database schema changes using versioned SQL migrations.

## Tech Stack

| Technology              | Purpose                               |
| ----------------------- | ------------------------------------- |
| Go                      | Backend development                   |
| PostgreSQL 17           | Relational database                   |
| pgx                     | PostgreSQL connectivity               |
| Docker Compose          | Local development environment         |
| Flyway / SQL migrations | Database schema versioning            |
| OpenAPI 3.0.3           | API contract documentation            |
| `mise`                  | Development tools and task automation |

## Architecture

Saldo follows **Hexagonal Architecture**, separating business rules from infrastructure and external interfaces.

The main idea is to keep the core application independent of HTTP handlers, database implementations, and other infrastructure details.

```text
.
├── cmd/
│   └── api/                 # Application entry point
├── internal/
│   ├── adapter/
│   │   ├── http/             # HTTP handlers and routing
│   │   └── ...               # Other adapters
│   ├── transaction/          # Transaction domain and use cases
│   └── ...                   # Other application components
├── migrations/               # Database migrations
├── api/
│   └── openapi.yaml          # OpenAPI specification
├── docker-compose.yml
├── mise.toml
└── go.mod
```

_The directory tree above is illustrative; adjust it to match the current repository structure._

### Architectural principles

- **Domain independence:** business rules should not depend on HTTP or database details.
- **Ports and adapters:** external dependencies interact with the application through defined interfaces.
- **Dependency inversion:** infrastructure implementations depend on application abstractions rather than the other way around.
- **Testability:** use cases can be tested independently of external infrastructure where appropriate.

## Getting Started

### Prerequisites

Install the following tools:

- Go 1.25.4 or a compatible version.
- Docker and Docker Compose.
- `mise` for managing development tools and project tasks.

### 1. Clone the repository

```bash
git clone https://github.com/DovCaio/saldo.git
cd saldo
```

### 2. Configure environment variables

Create a `.env` file in the project root with the variables required by the application.

For example:

```dotenv
DATABASE_URL=postgres://saldo:saldo@localhost:5432/saldo?sslmode=disable
```

Use the actual variable names expected by the application and never commit credentials for real environments.

### 3. Start PostgreSQL

Start the development database using Docker Compose:

```bash
docker compose up -d postgres
```

Make sure the database is healthy and accessible before starting the API.

### 4. Apply database migrations

Run the migration command configured for your development environment. Ensure that the migration tool is installed and its database connection points to the correct database.

### 5. Run the application

```bash
go run ./cmd/api
```

The API is expected to be available at:

```text
http://localhost:8080
```

Check the health endpoint:

```bash
curl http://localhost:8080/health
```

## API Reference

Base URL:

```text
http://localhost:8080
```

| Method | Endpoint                | Description                         |
| ------ | ----------------------- | ----------------------------------- |
| `GET`  | `/health`               | Check API health                    |
| `POST` | `/transactions`         | Register a transaction              |
| `GET`  | `/transactions`         | List and filter transactions        |
| `POST` | `/transactions/import`  | Import transactions from a CSV file |
| `GET`  | `/transactions/summary` | Retrieve financial summaries        |

### List transactions

```http
GET /transactions?date_from=2026-10-01&date_to=2026-10-31&type=expense&offset=0&limit=20
```

Supported query parameters:

| Parameter    | Description                                     |
| ------------ | ----------------------------------------------- |
| `date_from`  | Inclusive start date                            |
| `date_to`    | Inclusive end date                              |
| `type`       | Transaction type: `income` or `expense`         |
| `min_amount` | Minimum transaction amount                      |
| `max_amount` | Maximum transaction amount                      |
| `offset`     | Number of records to skip                       |
| `limit`      | Maximum number of records to return (up to 100) |

All filters are optional and can be combined.

### Register a transaction

```http
POST /transactions
Content-Type: application/json
```

Example request body (adapt the fields to the actual request DTO):

```json
{
  "date": "2026-10-21",
  "description": "Electricity bill",
  "amount": 185.9,
  "type": "expense"
}
```

### Import transactions

```http
POST /transactions/import
Content-Type: multipart/form-data
```

Send a CSV file using the multipart field expected by the import handler.

The import workflow is designed to support transaction ingestion from bank statements. Duplicate detection depends on the transaction identifiers and deduplication rules implemented by the application.

### Retrieve financial summaries

```http
GET /transactions/summary
```

The summary endpoint supports aggregating transaction data over a selected period. Consult `api/openapi.yaml` for the implemented query parameters and response schema.

## API Documentation

The API contract is maintained in [`api/openapi.yaml`](api/openapi.yaml), using the OpenAPI 3.0.3 specification.

To explore the documentation visually:

1. Open the [Swagger Editor](https://editor.swagger.io/).
2. Import `api/openapi.yaml`.
3. Explore the endpoints, query parameters, request bodies, and response schemas.

The OpenAPI specification should remain consistent with the actual HTTP handlers and DTOs.

## Testing

Saldo includes unit tests and database integration tests.

Run the unit tests:

```bash
go test ./...
```

The project also provides `mise` tasks for managing the test database and running integration tests.

For example:

```bash
mise run test-db-up
mise run ci
```

The test database is configured separately from the development database to avoid accidental data loss.

Check `mise.toml` for the exact task definitions and required environment variables.

## Development Workflow

The project uses automated tasks to make local development and testing more consistent.

Typical workflow:

1. Implement or update a use case.
2. Add unit tests for business rules.
3. Update the relevant HTTP adapter and DTOs.
4. Add integration tests for database-dependent behavior.
5. Update the OpenAPI specification when the API contract changes.
6. Run the test suite before committing changes.

## Future Improvements

Potential areas for further development include:

- Support for additional bank statement formats through interchangeable CSV parsers.
- More flexible transaction categorization.
- Expanded financial reporting and time-series analysis.
- Improved duplicate detection across imported statements.
- Additional API integration and end-to-end tests.

## Author

**Caio Jhonatan Alves Pereira**

- GitHub: [@DovCaio](https://github.com/DovCaio)

---

_Saldo is a personal software project built to explore Go, Hexagonal Architecture, database integration, API design, and automated testing through a practical financial management application._
