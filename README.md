# Open Telemetry LAB

## Execução do projeto em desenvolvimento

Para execução do projeto em ambiente de desenvolvimento deve primeiro criar um arquivo .env dentro da pasta service_b/cmd
Esse arquivo deve ter as informações especificadas no arquivo service_b/cmd/.env.example
WEATHER_API_KEY=

Ao criar o arquivo deve retornar a pasta raiz do projeto e executar o comando:
docker compose up --build

Com esse comando todas as aplicações necessárias para o projeto já estarão em execução

Para testar a aplicação pode utilizar o arquivo test/api.http -=> lembrando que para utilizar esse arquivo será necessário ter a extensão do visual studio code REST Client

ou fazer chamadas HTTP utilizando outro cliente HTTP seguindo o seguinte formato:

POST [http://localhost:8080]/ HTTP/1.1
Content-Type: application/json

{
    "cep": "01330902"
}
