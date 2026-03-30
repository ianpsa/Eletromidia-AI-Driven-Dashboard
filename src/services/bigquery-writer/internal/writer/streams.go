package writer

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	storagepb "cloud.google.com/go/bigquery/storage/apiv1/storagepb"
	"cloud.google.com/go/bigquery/storage/managedwriter"
	"cloud.google.com/go/bigquery/storage/managedwriter/adapt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type tableStream struct {
	name    string
	stream  *managedwriter.ManagedStream
	msgDesc protoreflect.MessageDescriptor
}

type bqStreams struct {
	age         *tableStream
	gender      *tableStream
	socialClass *tableStream
	target      *tableStream
	geodata     *tableStream
}

func (s *bqStreams) all() []*tableStream {
	return []*tableStream{s.age, s.gender, s.socialClass, s.target, s.geodata}
}

func closeStreams(streams []*managedwriter.ManagedStream) {
	for _, ms := range streams {
		_ = ms.Close()
	}
}

// bqSchemaToStorageSchema converts a bigquery.Schema to a storagepb.TableSchema.
func bqSchemaToStorageSchema(schema bigquery.Schema) *storagepb.TableSchema {
	fields := make([]*storagepb.TableFieldSchema, len(schema))
	for i, f := range schema {
		var typ storagepb.TableFieldSchema_Type
		switch f.Type {
		case bigquery.StringFieldType:
			typ = storagepb.TableFieldSchema_STRING
		case bigquery.FloatFieldType:
			typ = storagepb.TableFieldSchema_DOUBLE
		case bigquery.IntegerFieldType:
			typ = storagepb.TableFieldSchema_INT64
		default:
			typ = storagepb.TableFieldSchema_STRING
		}
		fields[i] = &storagepb.TableFieldSchema{
			Name: f.Name,
			Type: typ,
			Mode: storagepb.TableFieldSchema_NULLABLE,
		}
	}
	return &storagepb.TableSchema{Fields: fields}
}

// deriveProtoDescriptor converts a bigquery.Schema into the proto descriptor
// pair needed by the managed writer: a DescriptorProto for stream creation and
// a MessageDescriptor for building dynamic messages at flush time.
func deriveProtoDescriptor(schema bigquery.Schema, scope string) (
	*descriptorpb.DescriptorProto, protoreflect.MessageDescriptor, error,
) {
	storageSchema := bqSchemaToStorageSchema(schema)

	desc, err := adapt.StorageSchemaToProto2Descriptor(storageSchema, scope)
	if err != nil {
		return nil, nil, fmt.Errorf("StorageSchemaToProto2Descriptor(%s): %w", scope, err)
	}

	msgDesc, ok := desc.(protoreflect.MessageDescriptor)
	if !ok {
		return nil, nil, fmt.Errorf("descriptor for %s is not a MessageDescriptor", scope)
	}

	dp, err := adapt.NormalizeDescriptor(msgDesc)
	if err != nil {
		return nil, nil, fmt.Errorf("NormalizeDescriptor(%s): %w", scope, err)
	}

	return dp, msgDesc, nil
}

func initStreams(ctx context.Context, client *managedwriter.Client,
	projectID, datasetID string) (*bqStreams, error) {

	type tableDef struct {
		name     string
		row      any
		fieldPtr **tableStream
	}

	streams := &bqStreams{}

	tables := []tableDef{
		{"age", ageRow{}, &streams.age},
		{"gender", genderRow{}, &streams.gender},
		{"social_class", socialClassRow{}, &streams.socialClass},
		{"target", targetRow{}, &streams.target},
		{"geodata", geodataRow{}, &streams.geodata},
	}

	var created []*managedwriter.ManagedStream

	for _, t := range tables {
		schema, err := bigquery.InferSchema(t.row)
		if err != nil {
			closeStreams(created)
			return nil, fmt.Errorf("InferSchema(%s): %w", t.name, err)
		}

		dp, msgDesc, err := deriveProtoDescriptor(schema, t.name)
		if err != nil {
			closeStreams(created)
			return nil, err
		}

		destTable := fmt.Sprintf(
			"projects/%s/datasets/%s/tables/%s",
			projectID, datasetID, t.name,
		)

		ms, err := client.NewManagedStream(ctx,
			managedwriter.WithDestinationTable(destTable),
			managedwriter.WithType(managedwriter.DefaultStream),
			managedwriter.WithSchemaDescriptor(dp),
			managedwriter.EnableWriteRetries(true),
		)
		if err != nil {
			closeStreams(created)
			return nil, fmt.Errorf("NewManagedStream(%s): %w", t.name, err)
		}
		created = append(created, ms)

		*t.fieldPtr = &tableStream{
			name:    t.name,
			stream:  ms,
			msgDesc: msgDesc,
		}
	}

	return streams, nil
}

// rowData holds field values keyed by BigQuery column name.
type rowData map[string]any

// encodeRow serializes a rowData map into protobuf bytes using the given
// message descriptor. Only STRING, DOUBLE, and INT64 field kinds are supported.
func encodeRow(msgDesc protoreflect.MessageDescriptor, data rowData) ([]byte, error) {
	msg := dynamicpb.NewMessage(msgDesc)
	fields := msgDesc.Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		name := string(fd.Name())
		val, ok := data[name]
		if !ok {
			continue
		}
		switch fd.Kind() {
		case protoreflect.StringKind:
			s, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("field %q: expected string, got %T", name, val)
			}
			msg.Set(fd, protoreflect.ValueOfString(s))
		case protoreflect.DoubleKind:
			f, ok := val.(float64)
			if !ok {
				return nil, fmt.Errorf("field %q: expected float64, got %T", name, val)
			}
			msg.Set(fd, protoreflect.ValueOfFloat64(f))
		case protoreflect.Int64Kind:
			n, ok := val.(int64)
			if !ok {
				return nil, fmt.Errorf("field %q: expected int64, got %T", name, val)
			}
			msg.Set(fd, protoreflect.ValueOfInt64(n))
		default:
			return nil, fmt.Errorf("field %q: unsupported proto kind %v", name, fd.Kind())
		}
	}
	return proto.Marshal(msg)
}

func ageRowToData(r ageRow) rowData {
	return rowData{
		"id": r.ID, "x18_19": r.X1819, "x20_29": r.X2029,
		"x30_39": r.X3039, "x40_49": r.X4049, "x50_59": r.X5059,
		"x60_69": r.X6069, "x70_79": r.X7079, "x80_plus": r.X80Plus,
	}
}

func genderRowToData(r genderRow) rowData {
	return rowData{
		"id": r.ID, "feminine": r.Feminine, "masculine": r.Masculine,
	}
}

func socialClassRowToData(r socialClassRow) rowData {
	return rowData{
		"id": r.ID, "a_class": r.AClass, "b1_class": r.B1Class,
		"b2_class": r.B2Class, "c1_class": r.C1Class, "c2_class": r.C2Class,
		"de_class": r.DEClass,
	}
}

func targetRowToData(r targetRow) rowData {
	return rowData{
		"id": r.ID, "age_id": r.AgeID,
		"gender_id": r.GenderID, "social_class_id": r.SocialClassID,
	}
}

func geodataRowToData(r geodataRow) rowData {
	return rowData{
		"id": r.ID, "impression_hour": r.ImpressionHour,
		"location_id": r.LocationID, "uniques": r.Uniques,
		"latitude": r.Latitude, "longitude": r.Longitude,
		"uf_estado": r.UfEstado, "cidade": r.Cidade,
		"endereco": r.Endereco, "numero": r.Numero,
		"target_id": r.TargetID,
	}
}
