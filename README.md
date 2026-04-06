# WMS Microservices

Warehouse Management System - Modern Microservices Architecture

## 🏗️ Architecture

**Old:** .NET Framework 4.5 + ASP.NET Web Forms + SQL Server

**New:** Go + Next.js + PostgreSQL + Microservices

## 📁 Project Structure

```
wmsmicroservices/
├── services/              # Go microservices
│   ├── auth-service/      # Authentication & Authorization
│   ├── wms-service/       # Core WMS (Warehouse Management)
│   ├── pos-service/       # Point of Sales
│   ├── inventory-service/ # Inventory Management
│   ├── accounting-service/# Accounting & Finance
│   ├── hr-service/        # Human Capital Management
│   ├── restaurant-service/# Restaurant Module
│   └── reporting-service/ # Reports & Analytics
├── frontend/              # Next.js 14 Frontend
│   └── wms-web/
├── infrastructure/        # Docker, K8s, Terraform
├── shared/                # Shared libraries
├── scripts/               # Migration & utility scripts
└── docs/                  # Documentation
```

## 🛠️ Tech Stack

### Backend (Go)
- **Web Framework:** Echo/Gin
- **ORM:** GORM
- **Auth:** JWT
- **Database:** PostgreSQL
- **Cache:** Redis
- **Message Queue:** RabbitMQ/NATS

### Frontend (Next.js)
- **Framework:** Next.js 14 + React 18
- **Language:** TypeScript
- **Styling:** Tailwind CSS + Shadcn/ui
- **State:** Zustand
- **Data Fetching:** TanStack Query
- **Forms:** React Hook Form + Zod

### Infrastructure
- **Container:** Docker + Docker Compose
- **Orchestration:** Kubernetes
- **API Gateway:** Kong/Nginx
- **Monitoring:** Prometheus + Grafana

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+
- Node.js 20+
- PostgreSQL 15+

### Run with Docker Compose

```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

### Development Mode

```bash
# Start infrastructure
docker-compose up -d postgres redis rabbitmq

# Run individual services
cd services/auth-service && go run main.go
cd services/wms-service && go run main.go

# Run frontend
cd frontend/wms-web && npm run dev
```

## 📋 Migration Roadmap

| Phase | Module | Status |
|-------|--------|--------|
| 1 | Auth Service | 🚧 In Progress |
| 2 | WMS Core | 📋 Planned |
| 3 | POS Service | 📋 Planned |
| 4 | Inventory | 📋 Planned |
| 5 | Accounting | 📋 Planned |
| 6 | HR | 📋 Planned |
| 7 | Restaurant | 📋 Planned |
| 8 | Reporting | 📋 Planned |

## 📖 Documentation

- [Architecture](./docs/architecture.md)
- [API Specification](./docs/api-spec.md)
- [Database Migration](./docs/database-migration.md)
- [Development Guide](./docs/development.md)

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📄 License

[MIT](LICENSE)

## 🙏 Acknowledgments

- Original WMS .NET team
- Open source community
