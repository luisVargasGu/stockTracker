# Stock Portfolio Tracker

## **Overview**
The Stock Portfolio Tracker is an application designed to help users manage and monitor their investment portfolios. Users can add stocks to their portfolios, view real-time stock prices, analyze portfolio performance, and receive alerts for significant changes. This project demonstrates the use of modern microservices architecture in Go, with CI/CD deployments using GitHub Actions.

---

## **Features**

### **Core Features**
1. **User Management**
   - User registration and login (with JWT authentication).
   - Profile management.

2. **Portfolio Management**
   - Add, edit, and remove stocks from a portfolio.
   - View portfolio performance over time.

3. **Market Data Integration**
   - Fetch live stock prices using external APIs (e.g., Alpha Vantage, IEX Cloud).
   - Historical price trends and charts.

4. **Notifications**
   - Real-time alerts for significant price movements or portfolio changes.
   - Configurable notification preferences (email or SMS).

5. **Analytics Dashboard**
   - Gain/loss calculation.
   - Diversification analysis.
   - Risk metrics.

### **Stretch Goals**
- Social sharing of portfolios or stocks with friends.
- AI-based stock recommendations.
- Multi-currency support for international users.

---

## **Microservices Architecture**

### **Services Overview**

1. **User Service**
   - Handles user accounts, authentication, and profile data.
   - Database: PostgreSQL

2. **Portfolio Service**
   - Manages user portfolios and tracks owned stocks.
   - Database: PostgreSQL

3. **Market Data Service**
   - Fetches real-time and historical stock data using third-party APIs.
   - Caching for frequently accessed data (e.g., Redis).

4. **Notification Service**
   - Sends email or SMS alerts for portfolio updates.
   - Queue system for handling notification tasks (e.g., RabbitMQ).

5. **API Gateway**
   - Centralized entry point for routing requests to the appropriate microservices.

---

## **Technology Stack**

### **Backend**
- Programming Language: Go
- Frameworks: Gin or Go Kit

### **Databases**
- PostgreSQL: For persistent data storage.
- Redis: For caching and quick lookups.

### **Messaging Queue**
- RabbitMQ: For asynchronous communication between services.

### **Deployment**
- Docker: Containerization of all microservices.
- Kubernetes: Orchestration for production deployments.
- CI/CD: GitHub Actions for automated testing, building, and deploying.

---

## **File Structure**

```
/project-root
│
├── /api-gateway
│   ├── /config          # Configuration files
│   ├── /routes          # API Gateway routes and logic
│   ├── main.go          # Entry point for the API Gateway
│
├── /services
│   ├── /user-service
│   │   ├── /controllers # Handlers for routes
│   │   ├── /models      # Data models
│   │   ├── /repository  # Database layer
│   │   ├── /services    # Business logic
│   │   ├── main.go      # Entry point for the service
│   │   ├── Dockerfile   # Docker config for user service
│   │
│   ├── /portfolio-service (similar structure)
│   ├── /market-data-service (similar structure)
│   ├── /notification-service (similar structure)
│
├── /common
│   ├── /middleware      # Shared middleware (e.g., logging)
│   ├── /utils           # Utility functions
│   ├── /config          # Shared configurations
│
├── /deploy
│   ├── /k8s             # Kubernetes manifests
│   ├── docker-compose.yml # Local setup for all services
│
├── .github
│   ├── /workflows       # CI/CD workflows for GitHub Actions
│
├── README.md            # Documentation (this file)
└── .env                 # Environment variables
```

---

## **Getting Started**

### **Prerequisites**
- Docker and Docker Compose
- Go (v1.20+)
- PostgreSQL
- Redis
- GitHub account for CI/CD

### **Setup Instructions**

1. **Clone the Repository**
   ```bash
   git clone https://github.com/luisVargasGu/stockTracker.git
   cd stockTracker
   ```

2. **Run Locally**
   Use Docker Compose to spin up all services:
   ```bash
   docker-compose up --build
   ```

3. **Environment Variables**
   Create a `.env` file in the project root and populate it with:
   ```env
   DATABASE_URL=postgresql://user:password@localhost:5432/portfolio
   REDIS_URL=redis://localhost:6379
   API_KEY=your_market_data_api_key
   ```

4. **Access the Application**
   - API Gateway: `http://localhost:8080`
   - Swagger UI (if implemented): `http://localhost:8080/swagger`

---

## **CI/CD Pipeline**

### **GitHub Actions**
- Lint and test code on each push and pull request.
- Build Docker images and push to a container registry.
- Deploy to Kubernetes or cloud provider.

Sample workflow:
```yaml
name: CI/CD Pipeline

on:
  push:
    branches:
      - main

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
    - name: Checkout Code
      uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: 1.20

    - name: Run Tests
      run: go test ./...

    - name: Build Docker Images
      run: docker-compose build

    - name: Deploy to Kubernetes
      run: |
        kubectl apply -f deploy/k8s/
```

---

## **Contributing**
Pull requests are welcome! Please ensure your code adheres to the project’s linting and testing standards.

---

## **License**
This project is licensed under the MIT License. See `LICENSE` for details.


