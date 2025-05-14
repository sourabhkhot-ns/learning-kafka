# Microservices with Kafka

This project demonstrates a microservices architecture using Go and Apache Kafka for event-driven communication between services.

## Architecture

The project consists of three microservices:

1. **Order Service**: Accepts orders via REST API and publishes order events to Kafka
2. **Payment Service**: Processes payments for orders by consuming order events and publishing payment events
3. **Notification Service**: Sends notifications when payments are processed by consuming payment events

## Prerequisites

- Docker and Docker Compose
- Go 1.21 or later (for local development)

## Running the Services

1. Start all services using Docker Compose:
   ```bash
   docker-compose up --build
   ```

2. The services will be available at:
   - Order Service: http://localhost:8080
   - Kafka: localhost:9092
   - Zookeeper: localhost:2181

## Testing the Flow

1. Create a new order:
   ```bash
   curl -X POST http://localhost:8080/orders \
     -H "Content-Type: application/json" \
     -d '{
       "customer_id": 1,
       "restaurant_id": 1,
       "items": [
         {
           "item_id": 1,
           "quantity": 2,
           "price": 10.99
         },
         {
           "item_id": 2,
           "quantity": 1,
           "price": 15.99
         }
       ]
     }'
   ```

2. The following will happen automatically:
   - Order Service will:
     - Create an order with the items
     - Calculate the total amount
     - Publish an order event to Kafka

   - Payment Service will:
     - Consume the order event
     - Process the payment (simulated)
     - Generate a transaction ID
     - Publish a payment event

   - Notification Service will:
     - Consume the payment event
     - Send a detailed notification with order, customer, restaurant, and payment details

## Service Details

### Order Service
- Exposes REST API for creating orders
- Handles order items and calculates total amount
- Publishes to `order-events` Kafka topic
- Runs on port 8080

### Payment Service
- Consumes from `order-events` Kafka topic
- Processes payments (simulated)
- Generates transaction IDs
- Publishes to `payment-events` Kafka topic

### Notification Service
- Consumes from `payment-events` Kafka topic
- Sends detailed notifications including:
  - Order details
  - Customer information
  - Restaurant information
  - Payment status and transaction ID

## Development

To run the services locally:

1. Install dependencies for each service:
   ```bash
   cd Orders && go mod download
   cd ../Payments && go mod download
   cd ../Notifications && go mod download
   ```

2. Run each service:
   ```bash
   # In separate terminals
   cd Orders && go run cmd/main.go
   cd Payments && go run cmd/main.go
   cd Notifications && go run cmd/main.go
   ```

## Monitoring

- Each service logs its activities to stdout
- You can view logs using:
  ```bash
  docker-compose logs -f [service-name]
  ```

## Data Model

The services use the following data model (in-memory for demo purposes):

### Customer
- customer_id (int)
- name (string)
- email (string)
- phone (string)
- address (text)

### Restaurant
- restaurant_id (int)
- name (string)
- location (string)
- contact_info (string)

### MenuItem
- item_id (int)
- name (string)
- price (decimal)
- description (text)
- restaurant_id (int)

### Order
- order_id (int)
- customer_id (int)
- restaurant_id (int)
- order_date (datetime)
- total_amount (decimal)
- status (string)

### OrderItem
- order_item_id (int)
- order_id (int)
- item_id (int)
- quantity (int)
- price (decimal) # learning-kafka
