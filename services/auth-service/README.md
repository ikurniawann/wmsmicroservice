# WMS Microservices - Auth Service

## Overview
Authentication and Authorization microservice for WMS.

## Features
- JWT-based authentication
- User management (CRUD)
- Role-based access control (RBAC)
- Password hashing with bcrypt
- Token refresh mechanism

## Tech Stack
- Go 1.21+
- Echo framework
- GORM
- PostgreSQL
- JWT
- bcrypt

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/logout` - Logout user
- `POST /api/v1/auth/refresh` - Refresh access token

### Users
- `GET /api/v1/users` - List users
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Roles
- `GET /api/v1/roles` - List roles
- `POST /api/v1/roles` - Create role
- `PUT /api/v1/roles/:id` - Update role
- `DELETE /api/v1/roles/:id` - Delete role

## Environment Variables

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=wms
DB_PASSWORD=wms123
DB_NAME=wmsdb_auth
JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h
PORT=8081
```

## Run

```bash
cd services/auth-service
go mod tidy
go run main.go
```

## Docker

```bash
docker build -t wms-auth-service .
docker run -p 8081:8081 wms-auth-service
```
