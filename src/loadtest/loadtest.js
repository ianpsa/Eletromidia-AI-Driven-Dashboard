import { Writer, BALANCER_ROUND_ROBIN } from "k6/x/kafka";
import { check } from "k6";

// ─── Configuração do Producer ────────────────────────────────────────────────
const writer = new Writer({
  brokers: ["my-cluster-kafka-bootstrap.kafka.svc:9092"],
  topic: "geodata-dev",
  balancer: BALANCER_ROUND_ROBIN, // distribui entre as 50 partições
  batchSize: 500,                  // mensagens por batch
  batchTimeout: 5,                 // ms — flush rápido para alto throughput
  compression: "snappy",           // menor overhead de CPU vs gzip
  autoCreateTopic: false,          // tópico já existe
});

// ─── Cenário de Carga ────────────────────────────────────────────────────────
export const options = {
  scenarios: {
    kafka_high_throughput: {
      executor: "constant-arrival-rate",
      rate: 100000,       // 100.000 mensagens/segundo
      timeUnit: "1s",
      duration: "20s",
      preAllocatedVUs: 200,
      maxVUs: 500,
    },
  },
  thresholds: {
    // Falha o teste se a taxa de erros ultrapassar 1%
    errors: ["rate<0.01"],
  },
};

// ─── Payload de exemplo (campos fixos + variação por iteração) ───────────────
const LOCATIONS = [
  { location_id: 30847, latitude: "-2.358.334",  longitude: "-46.686.511", cidade: "São Paulo",       endereco: "Avenida Cidade Jardim",  uf_estado: "SP" },
  { location_id: 30901, latitude: "-23.550.520", longitude: "-46.633.308", cidade: "São Paulo",       endereco: "Avenida Paulista",        uf_estado: "SP" },
  { location_id: 31042, latitude: "-22.906.847", longitude: "-43.172.897", cidade: "Rio de Janeiro",  endereco: "Avenida Atlântica",       uf_estado: "RJ" },
  { location_id: 31150, latitude: "-19.916.681", longitude: "-43.934.493", cidade: "Belo Horizonte",  endereco: "Avenida Afonso Pena",     uf_estado: "MG" },
  { location_id: 31300, latitude: "-30.034.647", longitude: "-51.217.658", cidade: "Porto Alegre",    endereco: "Avenida Ipiranga",        uf_estado: "RS" },
];

const TARGET_TEMPLATE = {
  idade: {
    "18-19": 0.0773, "20-29": 0.1435, "30-39": 0.2389,
    "40-49": 0.2575, "50-59": 0.1736, "60-69": 0.0796,
    "70-79": 0.0235, "80+":   0.0060,
  },
  genero:        { F: 0.3857, M: 0.6143 },
  classe_social: { A: 0.0531, B1: 0.0839, B2: 0.2467, C1: 0.285, C2: 0.2579, DE: 0.0735 },
};

// ─── Função principal ────────────────────────────────────────────────────────
export default function () {
  const loc     = LOCATIONS[__ITER % LOCATIONS.length];
  const hour    = __ITER % 24;
  const uniques = parseFloat((Math.random() * 3000 + 500).toFixed(1));
  const numero  = Math.floor(Math.random() * 2000) + 1;

  const message = {
    impression_hour: hour,
    location_id:     loc.location_id,
    uniques:         uniques,
    latitude:        loc.latitude,
    longitude:       loc.longitude,
    uf_estado:       loc.uf_estado,
    cidade:          loc.cidade,
    endereco:        loc.endereco,
    numero:          numero,
    target:          JSON.stringify(TARGET_TEMPLATE),
  };

  const result = writer.produce({
    messages: [
      {
        // Chave baseada no location_id → garante ordenação por local dentro da partição
        key:   String(loc.location_id),
        value: JSON.stringify(message),
        headers: {
          "content-type":   "application/json",
          "source":         "k6-loadtest",
          "schema-version": "1.0",
        },
      },
    ],
  });

  check(result, { "mensagem produzida com sucesso": (r) => r === undefined || r === null });
}

// ─── Teardown ────────────────────────────────────────────────────────────────
export function teardown() {
  writer.close();
  console.log("✅ Writer fechado. Teste finalizado.");
}