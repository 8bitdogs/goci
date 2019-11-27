NAME=goci
VERSION=$(shell cat version)

debug: build clean up logs

build:
	docker build -t ${NAME}:${VERSION} .

up:
	docker run \
		--name ${NAME} \
		-e GIT_USERNAME=${GIT_USERNAME} \
		-e GIT_PASSWORD=${GIT_PASSWORD} \
		-e GIT_WEBHOOK_SECRET=${GIT_WEBHOOK_SECRET} \
		-d \
		${NAME}:${VERSION}

logs:
	docker logs -f ${NAME}

clean:
	-@docker kill ${NAME}
	-@docker rm ${NAME}
