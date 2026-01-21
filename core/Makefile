APP_NAME=gist-core
IMAGE_NAME=$(APP_NAME):latest
TAR_FILE=$(APP_NAME)-latest.tar.gz
SSH_HOST=vps
REMOTE_DIR=~/gist

deploy: ssh-check builddocker saveimage copy-image loadimage copy-compose
	@set -e; \
	echo "🚀 Запуск деплоя..."; \
	ssh $(SSH_HOST) "\
		   cd $(REMOTE_DIR) && \
		   docker-compose down && \
		   docker-compose up -d && \
		   docker image prune -f"
	@echo "✅ Деплой завершен!"
	docker image prune -f

ssh-check:
	@echo "🔧 Проверка SSH подключения..."
	@ssh -q $(SSH_HOST) "exit 0" && echo "✅ SSH подключение работает" || \
		(echo "❌ Ошибка SSH"; exit 1)

builddocker:
	@echo "🔨 Сборка Docker образа..."
	docker build -t $(IMAGE_NAME) -f Dockerfile .

saveimage:
	@echo "💾 Сохранение образа в архив..."
	docker save $(IMAGE_NAME) | gzip > $(TAR_FILE)

copy-image:
	@echo "📡 Копирование архива на сервер..."
	@scp $(TAR_FILE) $(SSH_HOST):/tmp/ && \
    		rm -f $(TAR_FILE) && \
    		echo "✅ Архив скопирован на сервер" || \
    		(echo "❌ Ошибка копирования"; exit 1)

loadimage:
	@echo "⬆️  Загрузка образа в Docker на сервере..."
	ssh $(SSH_HOST) "docker load -i /tmp/$(TAR_FILE) && rm -f /tmp/$(TAR_FILE)"
	@echo "✅ Образ загружен"

copy-compose:
	@ssh $(SSH_HOST) "mkdir -p $(REMOTE_DIR)"
	@scp ../compose.yaml $(SSH_HOST):$(REMOTE_DIR)/
	@echo "✅ compose.yaml скопирован"






stop:
	echo "🛑 Остановка проекта..."; \
    	ssh $(SSH_HOST) "\
    	cd $(REMOTE_DIR) && \
    	docker-compose down"

start:
	echo "🚀 Запуск проекта..."; \
          ssh $(SSH_HOST) "\
          cd $(REMOTE_DIR) && \
          docker-compose up -d"

clean:
	echo "🗑 Очистка..."; \
          ssh $(SSH_HOST) "\
          docker image prune -af && \
          docker container prune -f && \
          docker volume prune -f && \
          docker network prune -f \
          "