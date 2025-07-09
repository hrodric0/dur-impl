test:
	@echo "=== Executando testes ==="
	go test ./tests -v

coverage:
	@echo "=== Gerando profile de cobertura ==="
	go test ./tests -coverprofile=coverage.out

coverhtml: coverage
	@echo "=== Gerando relatório HTML de cobertura ==="
	go tool cover -html=coverage.out -o coverage.html
	@echo "Abra coverage.html no navegador para ver o relatório."
