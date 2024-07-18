.PHONY: test
test:
	go test -v -race ./...

.PHONY: test-integration
test-integration:
	@if [ -z "$$(docker ps -q -f name=bank_simulator)" ]; then \
		echo "Bank simulator is not running. Starting it..."; \
		docker-compose up -d payment; \
		echo "Waiting for container to be ready..."; \
		sleep 5; \
	else \
		echo "Payment container is already running."; \
	fi
	go test -v -race -tags=integration ./...