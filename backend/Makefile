APP=mp3tom4r-server
FLAGS="-s -w"
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${APP} -ldflags ${FLAGS}
build-deploy:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${APP} -ldflags ${FLAGS}
	rsync -avz ./ --include=${APP} --exclude="*" vps:/var/local/mp3tom4r/
	ssh vps "chmod a+x /var/local/mp3tom4r/${APP} && cd /var/local/dockers/mp3tom4r && docker-compose restart server"
clean:
	@rm ${APP}

