# Auto Message Sender API

Automatic message sending system - Retrieves pending messages from database every 2 minutes and sends them to webhook endpoints.


## Application Flow

```
1. Application Startup
   ├── MongoDB connection
   ├── Redis connection
   └── HTTP server startup

2. Scheduler Start (POST /scheduler/start)
   ├── Log "start" record to database
   ├── Start 2-minute ticker
   └── Background process with goroutine

3. Message Processing Loop (Every 2 minutes)
   ├── Fetch pending messages (2 items)
   ├── Start goroutine for each message
   ├── Send HTTP POST to webhook
   ├── Retry mechanism (3 attempts)
   ├── Update status based on result
   └── Cache to Redis

4. Scheduler Stop (POST /scheduler/stop)
   ├── Stop ticker
   ├── Wait for active jobs
   ├── Log "stop" record to database
   └── Graceful shutdown
```

## Cron Job Approach

**Golang Native Ticker Usage:**
- No Linux crontab or 3rd party cron library used
- Native Go timer with `time.NewTicker(2*time.Minute)`
- Channel based non blocking approach
- Background processing with goroutines


## Other Features
**Graceful Shutdown**
- SIGINT/SIGTERM signals are captured
- Active message sending operations are completed

**Retry Mechanism**
- Webhook calls are retried with exponential backoff
- Maximum 3 retry attempts per message
- Failed messages are marked as `failed` status

**Rate Limiting**
- IP-based rate limiting (100 req/minute)

**Structured Logging**
- Structured logging with Logrus
- Request/response tracking
- Detailed error tracking with error contexts

**Dependency Injection**
- Loose coupling between handlers
- Testable architecture
- Interface based design

**Input Validation**
- Phone number validation 
- Message content validation (160 character limit)
- Pagination parameter validation

**Concurrency and Thread Safety**
- Parallel message processing with goroutines
- Thread safe operations with mutexes
- Job synchronization with WaitGroups

**Audit Trail**
- Every scheduler start/stop operation is logged
- Start-Stop pairs are tracked
- Complete audit trail in database


## Architecture Approach

**Simplified Layered Architecture:**

```
cmd/auto-message-sender/     # Application entry point
├── main.go                  # Dependency wiring & server setup
└── router.go               # HTTP routing

internal/
├── handlers/               # HTTP handlers + business logic
├── dataOperations/        # Data access layer (MongoDB/Redis)
├── models/               # Domain models
├── config/              # Configuration management
└── middleware/         # HTTP middlewares

pkg/                    # Reusable packages
├── mongodb/           # MongoDB client wrapper
├── redisdb/          # Redis client wrapper
└── logger/          # Logging utilities
```

**Why This Architecture:**
- Clean code without over engineering
- Clear responsibility for each layer
- Handlers and data operations are separated
- New features can be easily added
- Code is readable and maintainable

**Alternative Architectures (but not used):**
- Clean Architecture: Too complex for this project
- Hexagonal Pattern: Overkill for small project
- Microservices: Monolith is sufficient

## Local Development

### **Prerequisites**
- Go 1.22+
- MongoDB (local or cloud)
- Redis (local or cloud)

### **1. Clone Repository**
```bash
git clone <repository-url>
cd auto-message-sender
```

### **2. Install Dependencies**
```bash
go mod download
```

### **3. Environment Configuration**
```bash
# Copy and edit env.prod file
cp env.prod .env

# Required environment variables:
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=auto_message_sender
REDIS_HOST=localhost
REDIS_PORT=6379
WEBHOOK_URL=https://webhook.site/your-unique-url
WEBHOOK_AUTH_KEY=your-auth-key
```

### **4. Database Setup**
```bash
# Ensure MongoDB and Redis are running
# Quick setup with Docker:
docker run -d -p 27017:27017 --name mongodb mongo:latest
docker run -d -p 6379:6379 --name redis redis:latest
```

### **5. Run Application**
```bash
# Development mode
go run cmd/auto-message-sender/main.go

# Production build
go build -o bin/app cmd/auto-message-sender/main.go
./bin/app
```

### **6. API Testing**
```bash
# Create test data (Insert the following data into the messages collection)

{"_id":"68a795179072a717e6a17f37","to":"+905551234570","content":"Test mesajı 9","status":"pending","created_at":{"$date":{"$numberLong":"1755798743223"}},"retry_count":{"$numberInt":"0"}}

# Start scheduler
curl -X POST http://localhost:8080/api/v1/scheduler/start

# List sent messages
curl "http://localhost:8080/api/v1/messages/sent?page=1&per_page=10"

# Swagger UI
http://localhost:8080/swagger/index.html
```

### **7. Run with Docker**
```bash
# Build
docker build -t auto-message-sender .

# Run
docker run -p 8080:8080 --env-file .env auto-message-sender
```

## API Endpoints

- `POST /api/v1/scheduler/start` - Start scheduler
- `POST /api/v1/scheduler/stop` - Stop scheduler
- `GET /api/v1/messages/sent` - List sent messages
- `GET /swagger/*` - API documentation

## Proof of requests

<img width="2098" height="699" alt="Screenshot 2025-08-22 at 01 47 19" src="https://github.com/user-attachments/assets/e46bb59c-7008-47a6-9a73-0d63307e1513" />



