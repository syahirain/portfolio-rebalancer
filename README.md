# Portfolio Rebalancer

A distributed system for managing and rebalancing investment portfolios. This project demonstrates a microservices architecture using Go, Kafka, and Elasticsearch.

## Architecture

The system consists of two main services decoupled by a message broker:

1.  **API Service**:
    *   Exposes HTTP endpoints for portfolio management.
    *   Validates user input and portfolio allocations.
    *   Persists portfolio state to **Elasticsearch**.
    *   Publishes rebalancing requests to a **Kafka** topic.

2.  **Consumer Service**:
    *   Consumes rebalancing requests from **Kafka**.
    *   Calculates the difference between current and new allocations.
    *   Generates necessary transactions to achieve the target allocation.
    *   Stores rebalancing transactions in **Elasticsearch**.

3.  **Infrastructure**:
    *   **Kafka**: Ensures asynchronous and reliable communication between the API and Consumer services.
    *   **Elasticsearch**: Serves as the primary data store for portfolios and transactions.
    *   **Zookeeper**: Manages the Kafka cluster.

## Tech Stack

*   **Language**: Go (Golang)
*   **Messaging**: Apache Kafka (using `segmentio/kafka-go`)
*   **Database**: Elasticsearch (using `elastic/go-elasticsearch`)
*   **Containerization**: Docker & Docker Compose

## Getting Started

### Prerequisites

*   Docker
*   Docker Compose

### Installation & Running

1.  Clone the repository:
    ```bash
    git clone https://github.com/your-username/portfolio-rebalancer.git
    cd portfolio-rebalancer
    ```

2.  Start the services using Docker Compose:
    ```bash
    docker-compose up --build
    ```

    This command will start:
    *   Zookeeper
    *   Kafka
    *   Elasticsearch
    *   API Service (exposed on port 8080)
    *   Consumer Service

3.  The API will be available at `http://localhost:8080`.

## API Reference

### 1. Create Portfolio

Creates a new investment portfolio for a user.

*   **Endpoint**: `POST /portfolio`
*   **Content-Type**: `application/json`
*   **Body Parameters**:
    *   `user_id` (string): Unique identifier for the user.
    *   `allocation` (object): Key-value pairs of asset names and their percentage allocation. Percentages must sum to 100.

**Example Request:**

```json
{
    "user_id": "1",
    "allocation": {
        "stocks": 60,
        "bonds": 30,
        "gold": 10
    }
}
```

**Example Response (Success):**

```json
{
    "success": true,
    "data": {
        "user_id": "1",
        "allocation": {
            "stocks": 60,
            "bonds": 30,
            "gold": 10
        }
    },
    "message": "Portfolio request accepted"
}
```

**Example Response (Error):**

```json
{
    "success": false,
    "message": "allocation percentages must sum to 100"
}
```

### 2. Trigger Rebalance

Triggers a rebalancing operation to adjust the portfolio to a new target allocation.

*   **Endpoint**: `POST /rebalance`
*   **Content-Type**: `application/json`
*   **Body Parameters**:
    *   `user_id` (string): Unique identifier for the user.
    *   `new_allocation` (object): The desired target allocation.

**Example Request:**

```json
{
    "user_id": "1",
    "new_allocation": {
        "stocks": 70,
        "bonds": 20,
        "gold": 10
    }
}
```

**Example Response (Success):**

```json
{
    "success": true,
    "data": {
        "user_id": "1",
        "new_allocation": {
            "stocks": 70,
            "bonds": 20,
            "gold": 10
        }
    },
    "message": "Rebalance request accepted"
}
```

**Example Response (Error):**

```json
{
    "success": false,
    "message": "New allocation is the same as current allocation"
}
```

## Fault Tolerance

The system is designed with several fault tolerance mechanisms:

*   **Asynchronous Processing**: Using Kafka decouples the request ingestion from processing, allowing the system to handle bursts of traffic without overwhelming the consumer.
*   **Graceful Shutdown**: Both API and Consumer services handle OS signals (SIGINT, SIGTERM) to ensure clean shutdowns (e.g., closing connections, finishing in-flight requests).
*   **Retry Mechanism**: The Consumer service implements exponential backoff (up to 5 attempts) when saving transactions to Elasticsearch to handle transient failures.
*   **Idempotency**: The system checks for duplicate rebalance requests using allocation hashes to prevent redundant processing.
*   **Container Recovery**: Docker Compose is configured with `restart: on-failure` to automatically restart services if they crash.


## Models

*   **Portfolio**
    *   `UserID`: Unique user identifier.
    *   `Allocation`: Map of asset classes to their percentage allocation (e.g., `{"stocks": 60, "bonds": 30, "gold": 10}`).

*   **UpdatedPortfolio**
    *   `UserID`: Unique user identifier.
    *   `NewAllocation`: The updated allocation provided by the 3rd party provider.

*   **RebalancePortfolioKafka**
    *   `UserID`: Unique user identifier.
    *   `NewAllocation`: The target allocation.
    *   `CurrentAllocation`: The user's original allocation before market changes.

*   **RebalanceTransaction**
    *   `UserID`: Unique user identifier.
    *   `Action`: Type of transaction (`BUY` or `SELL`).
    *   `Asset`: The asset class (e.g., `stocks`, `bonds`).
    *   `RebalancePercent`: The percentage of the asset to buy or sell.

*   **RebalanceRequest**
    *   `UserID`: Unique user identifier.
    *   `AllocationHash`: Hash of the updated allocation JSON (used for idempotency).

*   **APIResponse**
    *   `Success`: Boolean indicating request success.
    *   `Data`: Payload data (optional).
    *   `Message`: Error message or status description (optional).