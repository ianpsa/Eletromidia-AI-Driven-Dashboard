# Starting up

1. clone it!
2. create a `src/rust-producer/services/csv-kafka-producer/data` folder
3. Run:

<br>

``` bash

docker compose up --build

```

4. Test it!:

``` bash

curl -X POST http://localhost:5000/csv \
  -H "Content-Type: application/json" \
  -d '{"mensagens_por_segundo": <chosen_msg_persec_value>, "tempo_envio_segundos": <runtime>}'

```

ENJOY!!!!