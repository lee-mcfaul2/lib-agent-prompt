.PHONY: schemas-validate bundle-example verify-example verify-reproducible clean

schemas-validate:
	@echo "validating schemas..."
	@go run ./tools/bundle-builder validate --schema-lib ./schemas

bundle-example:
	@echo "building example bundle..."
	@go run ./tools/bundle-builder build \
		--schema-lib ./schemas \
		--prompts ./prompts/example \
		--services ./schemas/service-references \
		--output ./out/bundle-0.1.0-example \
		--version 0.1.0-example \
		--allow-placeholder
	@go run ./tools/bundle-builder pack ./out/bundle-0.1.0-example

verify-example:
	@echo "verifying example bundle..."
	@go run ./tools/verifier verify ./out/bundle-0.1.0-example.tar.gz

verify-reproducible:
	@echo "rebuilding and comparing digest..."
	@bash ./tests/reproducibility/check.sh

clean:
	rm -rf out/ dist/

.PHONY: demo
demo: bundle-example verify-example
	@echo
	@echo "Demo complete. Bundle at out/bundle-0.1.0-example.tar.gz"
