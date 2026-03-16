package storage


import (
   "context"
   "encoding/json"
   "fmt"
   "log"
   "regexp"
   "sync"
   "time"


   "cloud.google.com/go/storage"
   "google.golang.org/api/option"
)


var singleQuoteRe = regexp.MustCompile(`'([^']*)'`)


func normalizeTarget(raw string) string {
   return singleQuoteRe.ReplaceAllString(raw, `"$1"`)
}


type Message struct {
   Topic     string
   Partition int
   Offset    int64
   Timestamp time.Time
   Value     []byte
}


type bufferedRecord struct {
   data      []byte
   topic     string
   partition int
   offset    int64
   date      time.Time
}


type Writer struct {
   client   *storage.Client
   bucket   string
   basePath string
   flushSz  int


   mu     sync.Mutex
   buffer []bufferedRecord
}


func NewWriter(ctx context.Context, bucket, basePath, credentials string, flushSize int) (*Writer, error) {
   var opts []option.ClientOption
   if credentials != "" {
       opts = append(opts, option.WithCredentialsFile(credentials))
   }


   client, err := storage.NewClient(ctx, opts...)
   if err != nil {
       return nil, fmt.Errorf("erro ao criar cliente GCS: %w", err)
   }


   return &Writer{
       client:   client,
       bucket:   bucket,
       basePath: basePath,
       flushSz:  flushSize,
   }, nil
}


func (w *Writer) Close() error {
   return w.client.Close()
}


// Add normaliza o campo target (aspas simples para duplas), enriquece
// o registro com metadados Kafka e adiciona ao buffer.
// Retorna true quando o buffer esta cheio.
func (w *Writer) Add(msg Message) (bool, error) {
   var raw map[string]interface{}
   if err := json.Unmarshal(msg.Value, &raw); err != nil {
       return false, fmt.Errorf("erro ao decodificar mensagem kafka: %w", err)
   }


   if targetStr, ok := raw["target"].(string); ok {
       normalized := normalizeTarget(targetStr)
       var parsed interface{}
       if err := json.Unmarshal([]byte(normalized), &parsed); err == nil {
           raw["target"] = parsed
       }
   }


   raw["_kafka_partition"] = msg.Partition
   raw["_kafka_offset"] = msg.Offset
   raw["_kafka_timestamp"] = msg.Timestamp.UTC().Format(time.RFC3339)


   encoded, err := json.Marshal(raw)
   if err != nil {
       return false, fmt.Errorf("erro ao codificar registro: %w", err)
   }


   ts := msg.Timestamp
   if ts.IsZero() {
       ts = time.Now().UTC()
   }


   w.mu.Lock()
   w.buffer = append(w.buffer, bufferedRecord{
       data:      encoded,
       topic:     msg.Topic,
       partition: msg.Partition,
       offset:    msg.Offset,
       date:      ts.UTC(),
   })
   full := len(w.buffer) >= w.flushSz
   w.mu.Unlock()


   return full, nil
}


type groupKey struct {
   topic     string
   year      int
   month     time.Month
   day       int
   partition int
}


// Flush escreve todos os registros bufferizados no GCS como NDJSON,
// agrupados por topico/data/particao:
//
//  {basePath}/topics/{topic}/year=YYYY/month=MM/day=DD/{partition}-{startOffset}-{endOffset}.json
func (w *Writer) Flush(ctx context.Context) error {
   w.mu.Lock()
   if len(w.buffer) == 0 {
       w.mu.Unlock()
       return nil
   }
   batch := w.buffer
   w.buffer = nil
   w.mu.Unlock()


   groups := make(map[groupKey][]bufferedRecord)
   for _, r := range batch {
       k := groupKey{
           topic:     r.topic,
           year:      r.date.Year(),
           month:     r.date.Month(),
           day:       r.date.Day(),
           partition: r.partition,
       }
       groups[k] = append(groups[k], r)
   }


   var failed []bufferedRecord
   var firstErr error


   for k, records := range groups {
       var content []byte
       for _, r := range records {
           content = append(content, r.data...)
           content = append(content, '\n')
       }


       first := records[0]
       last := records[len(records)-1]


       objectPath := fmt.Sprintf("%s/topics/%s/year=%04d/month=%02d/day=%02d/%d-%d-%d.json",
           w.basePath,
           k.topic,
           k.year, k.month, k.day,
           k.partition,
           first.offset,
           last.offset,
       )


       obj := w.client.Bucket(w.bucket).Object(objectPath)
       wr := obj.NewWriter(ctx)
       wr.ContentType = "application/x-ndjson"


       if _, err := wr.Write(content); err != nil {
           wr.Close()
           failed = append(failed, records...)
           if firstErr == nil {
               firstErr = fmt.Errorf("erro ao escrever no GCS %s: %w", objectPath, err)
           }
           log.Printf("erro ao escrever no GCS %s: %v", objectPath, err)
           continue
       }


       if err := wr.Close(); err != nil {
           failed = append(failed, records...)
           if firstErr == nil {
               firstErr = fmt.Errorf("erro ao finalizar escrita no GCS %s: %w", objectPath, err)
           }
           log.Printf("erro ao finalizar escrita no GCS %s: %v", objectPath, err)
           continue
       }


       log.Printf("flush concluido: %d registros → gs://%s/%s", len(records), w.bucket, objectPath)
   }


   if len(failed) > 0 {
       w.mu.Lock()
       w.buffer = append(failed, w.buffer...)
       w.mu.Unlock()
   }


   return firstErr
}


func (w *Writer) Pending() int {
   w.mu.Lock()
   defer w.mu.Unlock()
   return len(w.buffer)
}

func (w *Writer) BucketReadnessProbe(ctx context.Context) error {
	_, err := w.client.Bucket(w.bucket).Attrs(ctx)
	if err != nil {
		return fmt.Errorf("Erro ao se conectar com o Bucket: %v", err)
	}

	return nil
}
