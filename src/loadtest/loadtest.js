import {
  Writer,
  SchemaRegistry,
  CODEC_SNAPPY,
  SCHEMA_TYPE_STRING,
  SCHEMA_TYPE_JSON,
  BALANCER_ROUND_ROBIN,
} from "k6/x/kafka";

const schemaRegistry = new SchemaRegistry();

const writer = new Writer({
  brokers: ["my-cluster-kafka-bootstrap.kafka.svc:9092"],
  topic: "geodata-dev",
  balancer: BALANCER_ROUND_ROBIN,
  batchSize: 1000,
  batchTimeout: 100,
  compression: CODEC_SNAPPY,
  autoCreateTopic: false,
  async: true,
});

export const options = {
  scenarios: {
    kafka_high_throughput: {
      executor: "constant-arrival-rate",
      rate: 33334,
      timeUnit: "1s",
      duration: "60s",
      preAllocatedVUs: 50,
      maxVUs: 100,
    },
  },
};

const LOCATIONS = [
  { location_id: 30847, latitude: "-2.358.334",  longitude: "-46.686.511", cidade: "São Paulo",      endereco: "Avenida Cidade Jardim", uf_estado: "SP" },
  { location_id: 30901, latitude: "-23.550.520", longitude: "-46.633.308", cidade: "São Paulo",      endereco: "Avenida Paulista",       uf_estado: "SP" },
  { location_id: 31042, latitude: "-22.906.847", longitude: "-43.172.897", cidade: "Rio de Janeiro", endereco: "Avenida Atlântica",      uf_estado: "RJ" },
  { location_id: 31150, latitude: "-19.916.681", longitude: "-43.934.493", cidade: "Belo Horizonte", endereco: "Avenida Afonso Pena",    uf_estado: "MG" },
  { location_id: 31300, latitude: "-30.034.647", longitude: "-51.217.658", cidade: "Porto Alegre",   endereco: "Avenida Ipiranga",       uf_estado: "RS" },
];

const TARGET = {
  idade: {
    "18-19": 0.0773, "20-29": 0.1435, "30-39": 0.2389,
    "40-49": 0.2575, "50-59": 0.1736, "60-69": 0.0796,
    "70-79": 0.0235, "80+":   0.0060,
  },
  genero:        { F: 0.3857, M: 0.6143 },
  classe_social: { A: 0.0531, B1: 0.0839, B2: 0.2467, C1: 0.285, C2: 0.2579, DE: 0.0735 },
};

export default function () {
  const loc     = LOCATIONS[__ITER % LOCATIONS.length];
  const hour    = __ITER % 24;
  const uniques = parseFloat((Math.random() * 3000 + 500).toFixed(1));
  const numero  = Math.floor(Math.random() * 2000) + 1;

  writer.produce({
    messages: [
      {
        // ✅ key: string serializada via SCHEMA_TYPE_STRING
        key: schemaRegistry.serialize({
          data: String(loc.location_id),
          schemaType: SCHEMA_TYPE_STRING,
        }),
        // ✅ value: JSON.stringify + SCHEMA_TYPE_STRING (padrão comprovado)
        value: schemaRegistry.serialize({
          data: JSON.stringify({
            impression_hour: hour,
            location_id:     loc.location_id,
            uniques:         uniques,
            latitude:        loc.latitude,
            longitude:       loc.longitude,
            uf_estado:       loc.uf_estado,
            cidade:          loc.cidade,
            endereco:        loc.endereco,
            numero:          numero,
            target:          TARGET,
          }),
          schemaType: SCHEMA_TYPE_STRING,
        }),
        headers: {
          "content-type":   "application/json",
          "source":         "k6-loadtest",
          "schema-version": "1.0",
        },
      },
    ],
  });
}

export function teardown() {
  writer.close();
  console.log("✅ Writer fechado. Teste finalizado.");
}