# Ascend

> 🏔️ Comprehensive weightlifting performance tracking platform

Ascend is a full-stack application designed to help athletes monitor progress, analyze form through video analysis, and achieve their strength training goals.

## Project Structure

```
ascend/
├── api/                    # Go backend API
│   ├── cmd/               # Application entry points
│   ├── internal/          # Private application code
│   ├── pkg/               # Public libraries
│   └── Dockerfile         # Production Docker image
├── mobile/                # React Native mobile app
│   ├── src/               # Application source code
│   ├── ios/               # iOS native code
│   ├── android/           # Android native code
│   └── README.md          # Mobile app documentation
├── scripts/               # Helper scripts
│   ├── start-dev.sh      # Start development environment
│   ├── stop-dev.sh       # Stop development environment
│   └── reset-dev.sh      # Reset development environment
├── docker-compose.yml     # Docker services configuration
├── docker-compose.dev.yml # Development overrides
└── DOCKER_SETUP.md       # Complete Docker documentation
```

## Quick Start

### Prerequisites

- Docker Desktop (or Docker Engine + Docker Compose)
- Node.js 18+ (for mobile development)
- Go 1.21+ (optional, for local API development)

### 1. Start Backend Services

```bash
# Start all backend services with Docker
./scripts/start-dev.sh
```

This starts:
- **PostgreSQL** (port 5432) - Database
- **Redis** (port 6379) - Cache
- **Go API** (port 8080) - Backend with hot-reload
- **Adminer** (port 8081) - Database UI
- **MinIO** (ports 9000, 9001) - Object storage

### 2. Start Mobile App

```bash
cd mobile
npm install

# iOS
npm run ios

# Android
npm run android
```

### 3. Access Services

- **API:** http://localhost:8080
- **API Health:** http://localhost:8080/health
- **Database UI:** http://localhost:8081
- **Object Storage UI:** http://localhost:9001

## Features

### Mobile App
- 📊 Comprehensive workout logging
- 📈 Progress analytics and ACWR monitoring
- 🎥 Video recording and form analysis
- 📴 Full offline mode with automatic sync
- 💪 1RM calculations and tracking
- 🔔 Smart notifications and reminders
- 🎯 Performance optimized

### Backend API
- 🚀 RESTful API built with Go
- 🔐 JWT authentication
- 💾 PostgreSQL database
- ⚡ Redis caching
- 📹 Video upload and storage
- 🔄 Real-time synchronization
- 📊 Analytics endpoints

## Development

### Backend Development

#### With Docker (Recommended)
```bash
# Start with hot-reload
./scripts/start-dev.sh

# View logs
docker-compose logs -f api

# Stop
./scripts/stop-dev.sh
```

#### Local Development
```bash
cd api

# Install dependencies
go mod download

# Run migrations
# (configure .env first)

# Start server
go run cmd/api/main.go
```

### Mobile Development

See [mobile/README.md](mobile/README.md) for complete mobile app documentation.

```bash
cd mobile

# Install dependencies
npm install

# iOS
npm run ios

# Android
npm run android

# Tests
npm test

# Type check
npm run type-check

# Lint
npm run lint
```

## Configuration

### Backend Configuration

Create `api/.env` from `api/.env.example`:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=ascend
DB_PASSWORD=ascend_dev_password
DB_NAME=ascend

JWT_SECRET=your_secret_key
PORT=8080
```

### Mobile Configuration

Create `mobile/.env` from `mobile/.env.example`:

```env
# For iOS Simulator or Android Emulator
API_BASE_URL=http://localhost:8080

# For Android Emulator (special host)
API_BASE_URL=http://10.0.2.2:8080

# For physical device (use your computer's IP)
API_BASE_URL=http://192.168.1.XXX:8080
```

## Testing

### Backend Tests
```bash
cd api
go test ./...

# With coverage
go test -cover ./...

# Verbose
go test -v ./...
```

### Mobile Tests
```bash
cd mobile

# All tests
npm test

# Watch mode
npm test -- --watch

# Coverage
npm test -- --coverage
```

## Documentation

- **[DOCKER_SETUP.md](DOCKER_SETUP.md)** - Complete Docker and docker-compose guide
- **[mobile/README.md](mobile/README.md)** - Mobile app documentation
- **[mobile/RELEASE_CHECKLIST.md](mobile/RELEASE_CHECKLIST.md)** - Release process
- **[api/README.md](api/README.md)** - API documentation (if available)

## Architecture

### System Architecture

```
┌─────────────┐
│   Mobile    │
│     App     │ ◄──── React Native + TypeScript
│ (iOS/Android)│
└──────┬──────┘
       │ HTTP/REST
       ▼
┌─────────────┐
│   API       │
│   Server    │ ◄──── Go + Chi Router
│  (Port 8080)│
└──────┬──────┘
       │
   ┌───┴───┬─────────┬─────────┐
   ▼       ▼         ▼         ▼
┌──────┐ ┌───┐  ┌────────┐ ┌──────┐
│ PostgreSQL Redis │ MinIO │ │ Other│
│  DB   │ │Cache  │ Object │ │Services
│       │ │     │ │Storage │ │      │
└───────┘ └───┘  └────────┘ └──────┘
```

### Mobile App Architecture

- **Offline-First** - WatermelonDB with automatic sync
- **Redux Toolkit** - Global state management
- **React Navigation** - Navigation and routing
- **Performance Optimized** - Image caching, query optimization
- **Error Boundaries** - Comprehensive error handling

### Backend Architecture

- **Go + Chi** - High-performance HTTP router
- **PostgreSQL** - Relational database
- **Redis** - Caching and sessions
- **MinIO** - S3-compatible object storage
- **JWT** - Stateless authentication

## Deployment

### Mobile App

See [mobile/README.md](mobile/README.md) for iOS App Store and Google Play Store deployment instructions.

### Backend API

#### Docker Deployment

```bash
# Build production image
docker build -t ascend-api:latest ./api

# Run with docker-compose
docker-compose -f docker-compose.yml up -d
```

#### Manual Deployment

```bash
cd api

# Build binary
CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

# Run
./api
```

## Troubleshooting

### Docker Issues

```bash
# Check if Docker is running
docker info

# View logs
docker-compose logs -f

# Reset everything
./scripts/reset-dev.sh
```

### Mobile App Issues

```bash
# Clear Metro cache
npm start -- --reset-cache

# Clean iOS
cd ios && pod deintegrate && pod install && cd ..

# Clean Android
cd android && ./gradlew clean && cd ..
```

### Database Issues

```bash
# Check database
docker-compose exec postgres psql -U ascend -d ascend

# Reset database
./scripts/stop-dev.sh --clean
./scripts/start-dev.sh
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details

## Support

For support:
- Email: support@ascend.app
- GitHub Issues: [github.com/your-org/ascend/issues](https://github.com/your-org/ascend/issues)

## Acknowledgments

- Go community for excellent libraries
- React Native team for the mobile framework
- WatermelonDB for offline-first database
- All open source contributors

---

Built with ❤️ for the strength training community
