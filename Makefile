help: ## ヘルプを表示します。
	@echo 'targetを下記から指定して実行してください'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## APIプログラムを実行する
	air -c .air.toml

run_no_air: ## airを使用せずにAPIプログラムを実行する
	go run main.go

docs: ## swaggerのドキュメントを生成する
	swag init 

migrate_local: ## ローカル環境のデータベースにマイグレーションを適用する
	FLAVOR=dev go run tools/migrate/migrate.go

migrate_prd: ## 本番環境のデータベースにマイグレーションを適用する
	FLAVOR=prd go run tools/migrate/migrate.go


mockgen: ## interfaceに従ってmockを生成する
	mkdir -p ./tests/mock
	mockgen -source=adapter/llm_adapter.go -destination=tests/mock/llm_adapter_mock.go -package=mock
	mockgen -source=adapter/r2_adapter.go -destination=tests/mock/r2_adapter_mock.go -package=mock
	mockgen -source=adapter/user_adapter.go -destination=tests/mock/user_adapter_mock.go -package=mock

# ==== Cloud Run デプロイ ====
GCP_PROJECT  ?= $(shell gcloud config get-value project)
GCP_REGION   ?= asia-northeast1
SERVICE_NAME ?= trihackathon-api
REPO_NAME    ?= trihackathon
IMAGE        ?= $(GCP_REGION)-docker.pkg.dev/$(GCP_PROJECT)/$(REPO_NAME)/$(SERVICE_NAME)

deploy: ## Cloud Runにデプロイする（gcloud CLIを使用）
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "Error: DATABASE_URL environment variable is required"; \
		echo "Usage: DATABASE_URL=\"postgresql://...\" make deploy"; \
		exit 1; \
	fi
	gcloud run deploy $(SERVICE_NAME) \
		--source . \
		--region $(GCP_REGION) \
		--platform managed \
		--allow-unauthenticated \
		--set-env-vars "DATABASE_URL=$(DATABASE_URL)"

deploy-image: ## Dockerイメージをビルドしてデプロイする
	docker build -t $(IMAGE):latest .
	docker push $(IMAGE):latest
	gcloud run deploy $(SERVICE_NAME) \
		--image $(IMAGE):latest \
		--region $(GCP_REGION) \
		--platform managed \
		--allow-unauthenticated

deploy-cloud-build: ## Cloud Buildでビルド＆デプロイする
	gcloud builds submit --config=cloudbuild.yaml
