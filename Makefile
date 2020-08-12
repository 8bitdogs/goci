NAME=goci
VERSION=$(shell cat version)

debug: build clean up logs

build:
	docker build -t ${NAME}:${VERSION} .

up:
	docker run \
		--name ${NAME} \
		-v $(CURDIR):/usr/local/ci \
		-e GITHUB_WEBHOOK_SECRET='' \
		-e GITHUB_TOKEN='' \
		-d \
		${NAME}:${VERSION} \
		goci -ci-config /usr/local/ci/ci.json

logs:
	docker logs -f ${NAME}

clean:
	-@docker kill ${NAME}
	-@docker rm ${NAME}
