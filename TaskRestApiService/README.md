# Rest Api service

Installation:
- install docker 
- run install_service.sh

to generate swagger doc run:

`go install github.com/swaggo/swag/cmd/swag@latest`

`swag init`

swagger will be available at /swagger/index.html

Логи доступны через elasticSearch по индексу service-logs* 
Можно смотреть логи в kebana через discover - FROM service-logs*