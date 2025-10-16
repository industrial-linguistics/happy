.PHONY: all build build-local clean test install deploy help

all: build

build:
	@echo "Building binaries..."
	go build -o bin/message-api cmd/message-api.go
	go build -o bin/happywatch cmd/happywatch.go
	go build -o bin/init-db cmd/init-db.go
	@echo "Built binaries in bin/"


clean:
	rm -rf bin/
	@echo "Cleaned build artifacts"

test:
	@echo "Running tests..."
	go test -v ./...


deploy: build
	@echo "Deploying to vhost directories..."
	@echo "Creating directories..."
	doas mkdir -p /var/www/vhosts/happy.industrial-linguistics.com/{data,v1,bin,htdocs}
	@echo "Copying binaries..."
	doas cp bin/message-api /var/www/vhosts/happy.industrial-linguistics.com/v1/
	doas cp bin/init-db /var/www/vhosts/happy.industrial-linguistics.com/bin/
	doas cp bin/happywatch /var/www/vhosts/happy.industrial-linguistics.com/bin/
	@echo "Creating symlinks..."
	cd /var/www/vhosts/happy.industrial-linguistics.com/v1 && \
		doas ln -sf message-api message && \
		doas ln -sf message-api messages && \
		doas ln -sf message-api status
	@echo "Setting permissions..."
	doas chown -R www:www /var/www/vhosts/happy.industrial-linguistics.com
	doas chmod 755 /var/www/vhosts/happy.industrial-linguistics.com/v1/*
	doas chmod 755 /var/www/vhosts/happy.industrial-linguistics.com/bin/*
	@echo "Initializing database (if needed)..."
	test -f /var/www/vhosts/happy.industrial-linguistics.com/data/positive-social.db || \
		doas -u www /var/www/vhosts/happy.industrial-linguistics.com/bin/init-db
	@echo ""
	@echo "✓ Deployed!"
	@echo ""
	@echo "Test with:"
	@echo "  curl http://localhost/v1/status"
	@echo "  curl https://happy.industrial-linguistics.com/v1/status"
	@echo ""
	@echo "Monitor with:"
	@echo "  /var/www/vhosts/happy.industrial-linguistics.com/bin/happywatch"

help:
	@echo "Available targets:"
	@echo "  build       - Build binaries (default)"
	@echo "  deploy      - Deploy to vhost directories (requires doas)"
	@echo "  clean       - Remove build artifacts"
	@echo "  test        - Run tests"
	@echo "  help        - Show this help"
	@echo ""
	@echo "Workflow:"
	@echo "  1. Edit code"
	@echo "  2. make build"
	@echo "  3. make deploy (uses doas to copy to /var/www/vhosts/...)"
