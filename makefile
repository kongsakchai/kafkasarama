.PHONY: up
up:
	@echo "Starting up the application..."
	@docker compose up -d
	@echo "Application is running."

.PHONY: down
down:
	@echo "Stopping the application..."
	@docker compose down
	@echo "Application has been stopped."


.PHONY: producer
producer:
	@echo "Starting the producer..."
	@go run ./producer/main.go

.PHONY: consumer
consumer:
	@echo "Starting the consumer..."
	@go run ./consumer/main.go

.PHONY: group
group:
	@echo "Starting the group consumer..."
	@go run ./consumergroup/.
	