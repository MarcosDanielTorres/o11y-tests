# 
# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0
#
#
#-- config vars
ZIP_NAME = layer-lrap.zip
LAYER_NAME = lrap
LAMBDA_LAYER_MARKER := .lambda-layer
CARGO_FEATURES := 

#-- current-condition vars
# Check if Docker is available or running-- needed by `cargo cross`.
#    modify if not cross-compiling or if using different tooling
#DOCKER_RUNNING := $(shell docker ps > /dev/null 2>&1 && echo -n yes)
#RS_FILES := $(shell find src -name "*.rs")
#RS_FILES := $(shell find src -name "*.rs")



.phony: build clean cargo zip clean-build clean-cargo deploy-layer doc

# * Build both x86_64 and aarch64 binaries
# * create a Layer '.zip'
# * use AWS CLI to publish Lambda layer
#
default: deploy-layer

clean: clean-build clean-cargo

clean-build: 
	@rm -rf build
	@rm -f .lambda-layer

clean-cargo:
	@cargo clean

build:
	@echo Building Rust application for x86_64
	@cross build --release --target x86_64-unknown-linux-gnu ${CARGO_FEATURES}
	@cp target/x86_64-unknown-linux-gnu/release/aws-lambda-runtime-api-proxy-rs build/lrap_x86_64

doc: 
	@cargo doc
	@echo
	@echo "Docs are located in target/doc/aws_lambda_runtime_api_proxy_rs/index.html"

build-RustExampleExtensionLayer:
	@echo Building Rust application for x86_64 aaaa

	@mkdir -p "$(ARTIFACTS_DIR)/extensions"
	@cross build --release --target x86_64-unknown-linux-gnu ${CARGO_FEATURES}
	@cp target/x86_64-unknown-linux-gnu/release/aws-lambda-runtime-api-proxy-rs $(ARTIFACTS_DIR)/layer_x86_64
	@cp opt/entrypoint $(ARTIFACTS_DIR)/extensions/$(LAYER_NAME)
	@cp opt/wrapper $(ARTIFACTS_DIR)/wrapper