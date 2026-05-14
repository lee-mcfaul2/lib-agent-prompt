.PHONY: schemas-validate bundle-example verify-example verify-reproducible clean

schemas-validate:
	@echo "validating schemas..."
	@go run ./tools/bundle-builder validate --source-dir .

bundle-example:
	@echo "building example bundle..."
	@go run ./tools/bundle-builder build \
		--source-dir . \
		--out-dir ./out/bundle-0.1.0-example \
		--bundle-version 0.1.0-example \
		--builder-id local-make
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
